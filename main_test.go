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

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Utility function to compart two configuration struct instances
func compareConfigs(c1 configuration, c2 configuration) bool {

	res := true
	res = res && c1.dbUrl == c2.dbUrl
	res = res && c1.dbPort == c2.dbPort
	res = res && c1.dbUser == c2.dbUser
	res = res && c1.dbPassword == c2.dbPassword
	res = res && c1.dbTargetDB == c2.dbTargetDB
	res = res && c1.symbol == c2.symbol
	res = res && c1.instance == c2.instance
	res = res && c1.mode == c2.mode
	res = res && c1.maxDepth == c2.maxDepth
	res = res && c1.jout == c2.jout
	res = res && len(c1.excludedBefore) == len(c2.excludedBefore)
	for i, item := range c1.excludedBefore {
		res = res && item == c2.excludedBefore[i]
	}
	res = res && len(c1.excludedAfter) == len(c2.excludedAfter)
	for i, item := range c1.excludedAfter {
		res = res && item == c2.excludedAfter[i]
	}
	return res
}

// Tests the ability to extract the configuration from command line arguments
func testConfig(t *testing.T) {

	var testConfig configuration = configuration{
		dbUrl:          "dummy",
		dbPort:         1234,
		dbUser:         "dummy",
		dbPassword:     "dummy",
		dbTargetDB:     "dummy",
		symbol:         "dummy",
		instance:       1234,
		mode:           1234,
		excludedBefore: []string{"dummy1", "dummy2", "dummy3"},
		excludedAfter:  []string{"dummyA", "dummyB", "dummyC"},
		maxDepth:       1234, //0: no limit
		jout:           "jsonOutputPlain",
		cmdlineNeeds:   map[string]bool{},
	}

	os.Args = []string{"nav"}
	conf, err := argsParse(cmdLineItemInit())
	if err == nil {
		t.Error("Error validating empty command line input against mandatory args")
	}

	if !compareConfigs(conf, defaultConfig) {
		t.Error("Unexpected change in default config")
	}

	os.Args = []string{"nav", "-i", "1", "-s"}
	conf, err = argsParse(cmdLineItemInit())
	if err == nil {
		t.Error("Error Missing switch argument not detected", conf)
	}

	os.Args = []string{"nav", "-i", "a", "-s", "symb"}
	conf, err = argsParse(cmdLineItemInit())
	if err == nil {
		t.Error("Error switch arg type mismatch not detected", conf)
	}

	os.Args = []string{"nav", "-i", "a", "-s", "symb", "-f"}
	conf, err = argsParse(cmdLineItemInit())
	if err == nil {
		t.Error("Error Missing optional switch argument not detected", conf)
	}

	_, filename, _, _ := runtime.Caller(0)
	current := filepath.Dir(filename)

	os.Args = []string{"nav", "-i", "1", "-s", "symb", "-f", current + "/t_files/dummy.json"}
	conf, err = argsParse(cmdLineItemInit())
	if err == nil {
		t.Error("Undetected missing file", conf)
	}
	if !compareConfigs(conf, defaultConfig) {
		t.Error("Unexpected change in default config")
	}

	os.Args = []string{"nav", "-i", "1", "-s", "symb", "-f", current + "/t_files/test1.json"}
	conf, err = argsParse(cmdLineItemInit())
	if err != nil {
		t.Error("Unexpected conf error while reading from existing file", err, current+"/t_files/test1.json")
	}
	if !compareConfigs(conf, testConfig) {
		t.Error("Unexpected difference between actual and loaded config", conf, testConfig)
	}

	tmp := testConfig
	tmp.dbUser = "new"
	os.Args = []string{"nav", "-i", "1", "-s", "symb", "-f", current + "/t_files/test1.json", "-u", "new"}
	conf, err = argsParse(cmdLineItemInit())
	if err != nil {
		t.Error("Unexpected conf error while reading from existing file", err, current+"/t_files/test1.json")
	}
	if !compareConfigs(conf, tmp) {
		t.Error("Unexpected difference between actual and loaded modified config", conf, testConfig)
	}

}
