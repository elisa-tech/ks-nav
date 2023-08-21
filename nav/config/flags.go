package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
	c "nav/constants"
)

func initFlagSet(fs *pflag.FlagSet, configPath *string) {
	fs.StringVarP(configPath, "config", "f", "", "path to `config` file")

	fs.StringP("symbol", "s", "", "name of the `symbol` to start the navigation from")
	fs.StringP("output-type", "j", c.DefaultOutputType, "`type` of output: graphOnly, jsonOutputPlain, jsonOutputB64 or jsonOutputGZB64")
	fs.IntP("max-depth", "x", c.DefaultMaxDepth, "max `number` of levels in call flow exploration (0=No limit)")
	fs.IntP("mode", "m", int(c.DefaultMode), "`mode` of plotting: 1=Symbols, 2=Subsystems, 3=Subsystems with labels, 4=Target subsystem isolation")
	fs.StringSliceP("excluded-before", "b", nil, "list of `symbols` to exclude before the target symbol")
	fs.StringSliceP("excluded-after", "a", nil, "list of `symbols` to exclude after the target symbol")
	fs.StringSliceP("target-subsys", "t", nil, "list of `subsystems` to include in the output\n")

	fs.StringP("db-driver", "e", c.DefaultDBDriver, "database `driver`: mysql, postgres or sqlite3")
	fs.IntP("db-instance", "i", c.DefaultDBInstance, "database `instance`")
	fs.IntP("output-format", "g", int(c.DefaultGOutputType), "Output format 1=dot 2=png 3=jpg 4=svg")
	fs.StringP("DBDSN", "d", "", "database `DSN` in the engine specific format\n"+
		"postgres: \"host=dbhost.com port=5432 user=username password=<password> dbname=kernel_bin sslmode=disable\"\n"+
		"mysql: \"username:@tcp(dbhost.com:3306)/dbname?multiStatements=true\"\n"+
		"sqlite3: \"file:db_file.db\"\n")

	fs.BoolP("help", "h", false, "show this help message and exit")
}

func parseCommandLine(fs *pflag.FlagSet, configPath *string, args []string) error {
	fs.SortFlags = false // disable sorting so flags get printed in the order they are defined
	initFlagSet(fs, configPath)

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("unable to parse args: %w", err)
	}

	help, err := fs.GetBool("help")
	if err != nil {
		return fmt.Errorf("unable to get help flag: %w", err)
	}
	if help {
		fmt.Printf("%s\n\n%s\n\nFlags:\n", c.AppName, c.AppUsage)
		fs.PrintDefaults()
		os.Exit(c.OSExitSuccess)
	}
	return nil
}

// If a flag is set, set the corresponding field in the config struct.
func setFlags(fs *pflag.FlagSet, cfg *ConfValues) {
	var flagToField = map[string]interface{}{
		"symbol":          &cfg.Symbol,
		"output-type":     &cfg.Type,
		"max-depth":       &cfg.MaxDepth,
		"mode":            &cfg.Mode,
		"excluded-before": &cfg.ExcludedBefore,
		"excluded-after":  &cfg.ExcludedAfter,
		"target-subsys":   &cfg.TargetSubsys,
		"db-driver":       &cfg.DBDriver,
		"DBDSN":           &cfg.DBDSN,
		"db-instance":     &cfg.DBInstance,
		"output-format":   &cfg.Graphviz,
	}

	fs.VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			if field, ok := flagToField[f.Name]; ok {
				setFlag(field, f.Value)
			}
		}
	})
}

func setFlag(field interface{}, value pflag.Value) {
	switch f := field.(type) {
	case *string:
		*f = value.String()
	case *int:
		if i, err := strconv.Atoi(value.String()); err == nil {
			*f = i
		}
	case *[]string:
		if sl, ok := value.(pflag.SliceValue); ok {
			*f = sl.GetSlice()
		}
	case *c.OutMode:
		if i, err := strconv.Atoi(value.String()); err == nil {
			*f = c.OutMode(i)
		}
	case *c.OutIMode:
		if i, err := strconv.Atoi(value.String()); err == nil {
			*f = c.OutIMode(i)
		}
	}
}
