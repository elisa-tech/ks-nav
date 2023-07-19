package config

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config Tests", func() {
	Describe("CLI", func() {
		When("The CLI is invoked with no arguments", func() {
			It("Should fail and inform the user of missing configuration", func() {
				os.Args = []string{"nav"}
				_, err := initConfig()
				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprintf("%s", err)).To(Equal("error: no configuration was specified, please specify a configuration file or use the command line flags; use -h or --help for more information"))
			})
		})

		When("The CLI is invoked with missing argument", func() {
			It("Should fail and inform the user of missing argument", func() {
				os.Args = []string{"nav", "--symbol"}
				_, err := initConfig()
				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprintf("%s", err)).To(Equal("error initializing command line flags: unable to parse args: flag needs an argument: --symbol"))
			})
		})

		When("The CLI is invoked with an invalid flag", func() {
			It("Should fail and inform the user about the invalid flag", func() {
				os.Args = []string{"nav", "--invalid"}
				_, err := initConfig()
				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprintf("%s", err)).To(Equal("error initializing command line flags: unable to parse args: unknown flag: --invalid"))
			})
		})

		When("The CLI is invoked with an invalid argument", func() {
			It("Should fail and inform the user about the invalid argument", func() {
				os.Args = []string{"nav", "--mode", "invalid"}
				_, err := initConfig()
				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprintf("%s", err)).To(Equal("error initializing command line flags: unable to parse args: invalid argument \"invalid\" for \"-m, --mode\" flag: strconv.ParseInt: parsing \"invalid\": invalid syntax"))
			})
		})

		When("The CLI is invoked with an invalid mode value", func() {
			It("Should fail and inform the user about the invalid mode", func() {
				os.Args = []string{"nav", "-s", "symbol", "--mode", "5"}
				_, err := initConfig()
				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprintf("%s", err)).To(Equal("invalid configuration: invalid output mode: 5\nChoose one of the following: 1=Symbols, 2=Subsystems, 3=Subsystems with labels, 4=Target subsystem isolation"))
			})
		})

		When("The CLI is invoked with an invalid depth value", func() {
			It("Should fail and inform the user about the invalid depth", func() {
				os.Args = []string{"nav", "-s", "symbol", "-x", "-1"}
				_, err := initConfig()
				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprintf("%s", err)).To(Equal("invalid configuration: invalid depth: -1"))
			})
		})

		When("The CLI is invoked with an invalid database instance", func() {
			It("Should fail and inform the user about the invalid database instance", func() {
				os.Args = []string{"nav", "-s", "symbol", "-i", "-1"}
				_, err := initConfig()
				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprintf("%s", err)).To(Equal("invalid configuration: invalid database instance: -1"))
			})
		})

		When("The CLI is invoked with an invalid database driver", func() {
			It("Should fail and inform the user about the invalid database driver", func() {
				os.Args = []string{"nav", "-s", "symbol", "-e", "invalidDriver"}
				_, err := initConfig()
				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprintf("%s", err)).To(Equal("invalid configuration: invalid database driver: invalidDriver\nChoose one of the following: mysql, postgres or sqlite3"))
			})
		})
	})

	Describe("config", func() {
		wd, _ := os.Getwd()
		testPath := filepath.Join(wd, "test_files/test.json")
		testMissingPath := filepath.Join(wd, "test_files/test_missing.json")
		testRewrite := filepath.Join(wd, "test_files/test_rewrite.json")
		var expected ConfValues

		BeforeEach(func() {
			expected = ConfValues{
				Symbol:         "__arm64_sys_getppid",
				Type:           "graphOnly",
				ExcludedBefore: []string{},
				ExcludedAfter:  []string{".*rcu.*"},
				TargetSubsys:   []string{},
				DBDriver:       "postgres",
				DBDSN:          "host=dbs.hqhome163.com port=5432 user=alessandro password=<password> dbname=kern_bin_new sslmode=disable",
				DBInstance:     1,
				MaxDepth:       1,
				Mode:           2,
			}
		})

		When("The -f (--config) flag is invoked", func() {
			It("Should parse the config file", func() {
				os.Args = []string{"nav", "-f", testPath}
				conf, err := initConfig()
				Expect(conf).To(Equal(expected))
				Expect(err).To(BeNil())
			})

			It("Should fail and inform the user about the invalid config file path", func() {
				os.Args = []string{"nav", "--config", "invalid"}
				_, err := initConfig()
				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprintf("%s", err)).To(Equal("error: problem while opening config file: open invalid: no such file or directory"))
			})

			It("Should parse the config file and use defaults for missing values", func() {
				os.Args = []string{"nav", "--config", testMissingPath}
				conf, err := initConfig()
				Expect(conf).To(Equal(expected))
				Expect(err).To(BeNil())
			})

			It("CLI should overwrite the values from the config", func() {
				os.Args = []string{"nav", "-f", testRewrite, "-s", "__arm64_sys_getppid", "-m", "2", "-i", "1"}
				conf, err := initConfig()
				Expect(conf).To(Equal(expected))
				Expect(err).To(BeNil())
			})
		})
	})
})
