/*
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *
 *   Name: nav - Kernel source code analysis tool
 *   Description: Extract call trees for kernel API
 *
 *   Author: Alessandro Carminati <acarmina@redhat.com>
 *   Author: Maurizio Papini <mpapini@redhat.com>
 *
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *
 *   Copyright (c) 2022 Red Hat, Inc. All rights reserved.
 *
 *   This copyrighted material is made available to anyone wishing
 *   to use, modify, copy, or redistribute it subject to the terms
 *   and conditions of the GNU General Public License version 2.
 *
 *   This program is distributed in the hope that it will be
 *   useful, but WITHOUT ANY WARRANTY; without even the implied
 *   warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
 *   PURPOSE. See the GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public
 *   License along with this program; if not, write to the Free
 *   Software Foundation, Inc., 51 Franklin Street, Fifth Floor,
 *   Boston, MA 02110-1301, USA.
 *
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 */

package main // import "golang.org/x/tools/cmd/gorename"

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

var confFn = "conf.json"

const (
	appName  string = "App Name: nav"
	appDescr string = "Descr: kernel symbol navigator"
)

type argFunc func(*configuration, []string) error

// Command line switch elements
type cmdLineItems struct {
	id        int
	switchStr string
	helpStr   string
	hasArg    bool
	needed    bool
	function  argFunc
}

// Represents the application configuration
type configuration struct {
	dbUrl          string
	dbPort         int
	dbUser         string
	dbPassword     string
	dbTargetDB     string
	symbol         string
	instance       int
	mode           outMode
	excludedBefore []string
	excludedAfter  []string
	targetSybsys   []string
	maxDepth       int
	jout           string
	cmdlineNeeds   map[string]bool
}

// Instance of default configuration values
var defaultConfig configuration = configuration{
	dbUrl:          "dbs.hqhome163.com",
	dbPort:         5432,
	dbUser:         "alessandro",
	dbPassword:     "<password>",
	dbTargetDB:     "kernel_bin",
	symbol:         "",
	instance:       0,
	mode:           printSubsys,
	excludedBefore: []string{},
	excludedAfter:  []string{},
	targetSybsys:   []string{},
	maxDepth:       0, //0: no limit
	jout:           "GraphOnly",
	cmdlineNeeds:   map[string]bool{},
}

// Inserts a commandline item item, which is composed by:
// * switch string
// * switch descriptio
// * if the switch requires an additiona argument
// * a pointer to the function that manages the switch
// * the configuration that gets updated
func pushCmdLineItem(switchStr string, helpStr string, hasArg bool, needed bool, function argFunc, cmdLine *[]cmdLineItems) {
	*cmdLine = append(*cmdLine, cmdLineItems{id: len(*cmdLine) + 1, switchStr: switchStr, helpStr: helpStr, hasArg: hasArg, needed: needed, function: function})
}

// This function initializes configuration parser subsystem
// Inserts all the commandline switches suppported by the application
func cmdLineItemInit() []cmdLineItems {
	var res []cmdLineItems

	pushCmdLineItem("-j", "Force Json output with subsystems data", true, false, funcOuttype, &res)
	pushCmdLineItem("-s", "Specifies symbol", true, true, funcSymbol, &res)
	pushCmdLineItem("-i", "Specifies instance", true, true, funcInstance, &res)
	pushCmdLineItem("-f", "Specifies config file", true, false, funcJconf, &res)
	pushCmdLineItem("-u", "Forces use specified database userid", true, false, funcDBUser, &res)
	pushCmdLineItem("-p", "Forecs use specified password", true, false, funcDBPass, &res)
	pushCmdLineItem("-d", "Forecs use specified DBhost", true, false, funcDBHost, &res)
	pushCmdLineItem("-p", "Forecs use specified DBPort", true, false, funcDBPort, &res)
	pushCmdLineItem("-m", "Sets display mode 2=subsystems,1=all", true, false, funcMode, &res)
	pushCmdLineItem("-x", "Specify Max depth in call flow exploration", true, false, funcDepth, &res)
	pushCmdLineItem("-h", "This Help", false, false, funcHelp, &res)

	return res
}

func funcHelp(conf *configuration, fn []string) error {
	return errors.New("Command Help")
}

func funcOuttype(conf *configuration, jout []string) error {
	(*conf).jout = jout[0]
	return nil
}

func funcJconf(conf *configuration, fn []string) error {
	jsonFile, err := os.Open(fn[0])
	if err != nil {
		return err
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	jsonFile.Close()
	err = json.Unmarshal(byteValue, conf)
	if err != nil {
		return err
	}
	return nil
}

func funcSymbol(conf *configuration, fn []string) error {
	(*conf).symbol = fn[0]
	return nil
}

func funcDBUser(conf *configuration, user []string) error {
	(*conf).dbUser = user[0]
	return nil
}

func funcDBPass(conf *configuration, pass []string) error {
	(*conf).dbPassword = pass[0]
	return nil
}

func funcDBHost(conf *configuration, host []string) error {
	(*conf).dbUrl = host[0]
	return nil
}

func funcDBPort(conf *configuration, port []string) error {
	s, err := strconv.Atoi(port[0])
	if err != nil {
		return err
	}
	(*conf).dbPort = s
	return nil
}

func funcDepth(conf *configuration, depth []string) error {
	s, err := strconv.Atoi(depth[0])
	if err != nil {
		return err
	}
	if s < 0 {
		return errors.New("Depth must be >= 0")
	}
	(*conf).maxDepth = s
	return nil
}

func funcInstance(conf *configuration, instance []string) error {
	s, err := strconv.Atoi(instance[0])
	if err != nil {
		return err
	}
	(*conf).instance = s
	return nil
}

func funcMode(conf *configuration, mode []string) error {
	s, err := strconv.Atoi(mode[0])
	if err != nil {
		return err
	}
	if outMode(s) < printAll || outMode(s) >= OutModeLast {
		return errors.New("unsupported mode")
	}
	(*conf).mode = outMode(s)
	return nil
}

// Uses commandline args to generate the help string
func printHelp(lines []cmdLineItems) {

	fmt.Println(appName)
	fmt.Println(appDescr)
	for _, item := range lines {
		fmt.Printf(
			"\t%s\t%s\t%s\n",
			item.switchStr,
			func(a bool) string {
				if a {
					return "<v>"
				}
				return ""
			}(item.hasArg),
			item.helpStr,
		)
	}
}

// Used to parse the command line and generate the command line
func argsParse(lines []cmdLineItems) (configuration, error) {
	var extra bool = false
	var conf configuration = defaultConfig
	var f argFunc

	for _, item := range lines {
		if item.needed {
			conf.cmdlineNeeds[item.switchStr] = false
		}
	}

	for _, osArg := range os.Args[1:] {
		if !extra {
			for _, arg := range lines {
				if arg.switchStr == osArg {
					if arg.needed {
						conf.cmdlineNeeds[arg.switchStr] = true
					}
					if arg.hasArg {
						f = arg.function
						extra = true
						break
					}
					err := arg.function(&conf, []string{})
					if err != nil {
						return defaultConfig, err
					}
				}
			}
			continue
		}
		if extra {
			err := f(&conf, []string{osArg})
			if err != nil {
				return defaultConfig, err
			}
			extra = false
		}

	}
	if extra {
		return defaultConfig, errors.New("Missing switch arg")
	}

	res := true
	for _, element := range conf.cmdlineNeeds {
		res = res && element
	}
	if res {
		return conf, nil
	}
	return defaultConfig, errors.New("Missing needed arg")
}
