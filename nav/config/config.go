package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	c "nav/constants"
)

type Config struct {
	ConfValues ConfValues
	// TODO: add DB connection instance here
}

type ConfValues struct {
	Symbol         string     `json:"symbol"`
	Type           string     `json:"output_type"`
	DBDriver       string     `json:"db_driver"`
	DBDSN          string     `json:"DBDSN"`
	ExcludedBefore []string   `json:"excluded_before"`
	ExcludedAfter  []string   `json:"excluded_after"`
	TargetSubsys   []string   `json:"target_subsys"`
	MaxDepth       int        `json:"max_depth"`
	Mode           c.OutMode  `json:"mode"`
	Graphviz       c.OutIMode `json:"out_type"`
	DBInstance     int        `json:"db_instance"`
}

// New creates a new Config instance and returns a pointer to it.
func New() (*Config, error) {
	config := Config{}
	confValues, err := initConfig()
	config.ConfValues = confValues
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// initConfig gathers the configuration values from command line and config file (if provided) and returns
// them in a ConfValues struct. If both are provided, the command line flags take precedence.
func initConfig() (ConfValues, error) {
	if len(os.Args) == 1 {
		return ConfValues{}, fmt.Errorf("error: no configuration was specified, please specify a configuration file or use the command line flags; " +
			"use -h or --help for more information")
	}
	fs := pflag.NewFlagSet("nav", pflag.ContinueOnError)
	var confValues ConfValues
	var configPath string

	err := parseCommandLine(fs, &configPath, os.Args)
	if err != nil {
		return ConfValues{}, fmt.Errorf("error initializing command line flags: %w", err)
	}

	if configPath != "" {
		err = loadConfigFile(&confValues, configPath)
		if err != nil {
			return ConfValues{}, fmt.Errorf("error: %w", err)
		}
	}

	setFlags(fs, &confValues)
	if err := confValues.validate(); err != nil {
		return ConfValues{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return confValues, nil
}

func loadConfigFile(cfg *ConfValues, configPath string) error {
	cleanedPath := filepath.Clean(configPath)

	if cleanedPath != configPath {
		return fmt.Errorf("invalid file path, did you mean %s", cleanedPath)
	}

	jsonFile, err := os.Open(cleanedPath)
	if err != nil {
		return fmt.Errorf("problem while opening config file: %w", err)
	}
	defer func() {
		closeErr := jsonFile.Close()
		if err == nil {
			err = closeErr
		}
	}()

	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, cfg)
	if err != nil {
		return fmt.Errorf("problem while parsing config file: %w", err)
	}
	return nil
}

func (cfg *ConfValues) validate() error {
	if cfg.Symbol == "" {
		return fmt.Errorf("symbol must be specified")
	}
	if cfg.MaxDepth < 0 {
		return fmt.Errorf("invalid depth: %d", cfg.MaxDepth)
	}
	if err := validateDBInstance(&cfg.DBInstance); err != nil {
		return err
	}
	if err := validateDBDriver(&cfg.DBDriver); err != nil {
		return err
	}
	if err := validateMode(&cfg.Mode); err != nil {
		return err
	}
	if err := validateType(&cfg.Type); err != nil {
		return err
	}
	if err := validateGType(&cfg.Graphviz); err != nil {
		return err
	}

	return nil
}

func validateDBInstance(i *int) error {
	if *i < 0 {
		return fmt.Errorf("invalid database instance: %d", *i)
	}
	switch *i {
	case 0:
		*i = c.DefaultDBInstance
		fmt.Printf("No database instance specified. Defaulting to %d.\n", c.DefaultDBInstance)
		return nil
	default:
		return nil
	}
}

func validateDBDriver(d *string) error {
	switch *d {
	case "":
		*d = c.DefaultDBDriver
		fmt.Printf("No database driver specified. Defaulting to %s.\n", c.DefaultDBDriver)
		return nil
	case "mysql", "postgres", "sqlite3":
		return nil
	default:
		return fmt.Errorf("invalid database driver: %s\nChoose one of the following: mysql, postgres or sqlite3", *d)
	}
}

func validateMode(m *c.OutMode) error {
	switch *m {
	case 0:
		*m = c.DefaultMode
		fmt.Printf("No output mode specified. Defaulting to %d.\n", c.DefaultMode)
		return nil
	case c.PrintAll, c.PrintSubsys, c.PrintSubsysWs, c.PrintTargeted, c.GDataFunc, c.GDataSubs:
		return nil
	default:
		return fmt.Errorf("invalid output mode: %d\nChoose one of the following: 1=Symbols, 2=Subsystems, "+
			"3=Subsystems with labels, 4=Target subsystem isolation", *m)
	}
}

func validateType(t *string) error {
	switch *t {
	case "":
		*t = c.DefaultOutputType
		fmt.Printf("No output type specified. Defaulting to %s.\n", c.DefaultOutputType)
		return nil
	case "graphOnly", "jsonOutputPlain", "jsonOutputB64", "jsonOutputGZB64":
		return nil
	default:
		return fmt.Errorf("invalid output type: %s\nChoose one of the following: graphOnly, jsonOutputPlain, jsonOutputB64 or jsonOutputGZB64", *t)
	}
}

func validateGType(t *c.OutIMode) error {
        switch *t {
        case 0:
                *t = c.DefaultGOutputType
                fmt.Printf("No output format specified. Defaulting to %d.\n", c.DefaultGOutputType)
                return nil
        case c.OText, c.OPNG, c.OJPG, c.OSVG:
                return nil
        default:
                return fmt.Errorf("invalid graphviz output type: %s\nSee help for more details.", *t)
        }
}
