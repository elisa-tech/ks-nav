/*
 * Copyright (c) 2022 Red Hat, Inc.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Utility function to compare two configuration struct instances.
func compareConfigs(c1 configuration, c2 configuration) bool {

	res := true
	res = res && c1.DBDriver == c2.DBDriver
	res = res && c1.DBDSN == c2.DBDSN
	res = res && c1.Symbol == c2.Symbol
	res = res && c1.Instance == c2.Instance
	res = res && c1.Mode == c2.Mode
	res = res && c1.MaxDepth == c2.MaxDepth
	res = res && c1.Jout == c2.Jout
	res = res && len(c1.ExcludedBefore) == len(c2.ExcludedBefore)
	for i, item := range c1.ExcludedBefore {
		res = res && item == c2.ExcludedBefore[i]
	}
	res = res && len(c1.ExcludedAfter) == len(c2.ExcludedAfter)
	for i, item := range c1.ExcludedAfter {
		res = res && item == c2.ExcludedAfter[i]
	}
	return res
}

// Tests the ability to extract the configuration from command line arguments
func TestConfig(t *testing.T) {

	var testConfig = configuration{
		DBDriver:       "dummy",
		DBDSN:          "dummy",
		Symbol:         "dummy",
		Instance:       1234,
		Mode:           1234,
		ExcludedBefore: []string{"dummy1", "dummy2", "dummy3"},
		ExcludedAfter:  []string{"dummyA", "dummyB", "dummyC"},
		MaxDepth:       1234, //0: no limit
		Jout:           "jsonOutputPlain",
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
	tmp.DBDriver = "new"
	os.Args = []string{"nav", "-i", "1", "-s", "symb", "-f", current + "/t_files/test1.json", "-e", "new"}
	conf, err = argsParse(cmdLineItemInit())
	if err != nil {
		t.Error("Unexpected conf error while reading from existing file", err, current+"/t_files/test1.json")
	}
	if !compareConfigs(conf, tmp) {
		t.Error("Unexpected difference between actual and loaded modified config", conf, testConfig)
	}

}
