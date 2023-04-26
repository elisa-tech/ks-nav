/*
 * Copyright (c) 2022 Red Hat, Inc.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

package main

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"nav/config"
	c "nav/constants"
	"os"
	"strings"
)

const jsonOutputFMT string = "{\"graph\": \"%s\",\"graph_type\":\"%s\",\"symbols\": [%s]}"

var fmtDot = []string{
	"",
	"\"%s\"->\"%s\" \n",
	"\\\"%s\\\"->\\\"%s\\\" \\\\\\n",
	"\"%s\"->\"%s\" \n",
	"\"%s\"->\"%s\" \n",
}

var fmtDotHeader = []string{
	"",
	"digraph G {\nrankdir=\"LR\"\n",
	"digraph G {\\\\\\nrankdir=\"LR\"\\\\\\n",
	"digraph G {\nrankdir=\"LR\"\n",
	"digraph G {\nrankdir=\"LR\"\n",
}

var fmtDotNodeHighlightWSymb = "\"%[1]s\" [shape=record style=\"rounded,filled,bold\" fillcolor=yellow label=\"%[1]s|%[2]s\"]\n"
var fmtDotNodeHighlightWoSymb = "\"%[1]s\" [shape=record style=\"rounded,filled,bold\" fillcolor=yellow label=\"%[1]s\"]\n"

func opt2num(s string) int {
	var opt = map[string]int{
		"graphOnly":       c.GraphOnly,
		"jsonOutputPlain": c.JsonOutputPlain,
		"jsonOutputB64":   c.JsonOutputB64,
		"jsonOutputGZB64": c.JsonOutputGZB64,
	}
	val, ok := opt[s]
	if !ok {
		return 0
	}
	return val
}

func decorateLine(l string, r string, adjm []adjM) string {
	var res = " [label=\""

	for _, item := range adjm {
		if (item.l.subsys == l) && (item.r.subsys == r) {
			tmp := fmt.Sprintf("%s([%s]%s),\\n", item.r.symbol, item.r.addressRef, item.r.sourceRef)
			if !strings.Contains(res, tmp) {
				res += fmt.Sprintf("%s([%s]%s),\\n", item.r.symbol, item.r.addressRef, item.r.sourceRef)
			}
		}
	}
	res += "\"]"
	return res
}

func decorate(dotStr string, adjm []adjM) string {
	var res string

	dotBody := strings.Split(dotStr, "\n")
	for i, line := range dotBody {
		split := strings.Split(line, "->")
		if len(split) == 2 {
			res = res + dotBody[i] + decorateLine(strings.TrimSpace(strings.ReplaceAll(split[0], "\"", "")), strings.TrimSpace(strings.ReplaceAll(split[1], "\"", "")), adjm) + "\n"
		}
	}
	return res
}

func generateOutput(db *sql.DB, cfg *config.Config) (string, error) {
	var graphOutput string
	var jsonOutput string
	var prod = map[string]int{}
	var visited []int
	var entryName string
	var output string
	var adjm []adjM

	conf := cfg.ConfValues

	cache := make(map[int][]entry)
	cache2 := make(map[int]entry)
	cache3 := make(map[string]string)

	start, err := sym2num(db, conf.Symbol, conf.DBInstance)
	if err != nil {
		fmt.Println("Symbol not found")
		return "", err
	}

	graphOutput = fmtDotHeader[opt2num(conf.Type)]
	entry, err := getEntryById(db, start, conf.DBInstance, cache2)
	if err != nil {
		return "", err
	} else {
		entryName = entry.symbol
	}
	startSubsys, _ := getSubsysFromSymbolName(db, entryName, conf.DBInstance, cache3)
	if startSubsys == "" {
		startSubsys = SUBSYS_UNDEF
	}

	if (conf.Mode == c.PrintTargeted) && len(conf.TargetSubsys) == 0 {
		targSubsysTmp, err := getSubsysFromSymbolName(db, conf.Symbol, conf.DBInstance, cache3)
		if err != nil {
			panic(err)
		}
		conf.TargetSubsys = append(conf.TargetSubsys, targSubsysTmp)
	}

	navigate(db, start, node{startSubsys, entryName, "entry point", "0x0"}, conf.TargetSubsys, &visited, &adjm, prod, conf.DBInstance, Cache{cache, cache2, cache3}, conf.Mode, conf.ExcludedAfter, conf.ExcludedBefore, 0, conf.MaxDepth, fmtDot[opt2num(conf.Type)], &output)

	if (conf.Mode == c.PrintSubsysWs) || (conf.Mode == c.PrintTargeted) {
		output = decorate(output, adjm)
	}

	graphOutput += output
	if conf.Mode == c.PrintTargeted {
		for _, i := range conf.TargetSubsys {
			if cache3[conf.Symbol] == i {
				graphOutput += fmt.Sprintf(fmtDotNodeHighlightWSymb, i, conf.Symbol)
			} else {
				graphOutput += fmt.Sprintf(fmtDotNodeHighlightWoSymb, i)
			}
		}
	}
	graphOutput += "}"

	symbdata, err := symbSubsys(db, visited, conf.DBInstance, Cache{cache, cache2, cache3})
	if err != nil {
		return "", err
	}

	switch opt2num(conf.Type) {
	case c.GraphOnly:
		jsonOutput = graphOutput
	case c.JsonOutputPlain:
		jsonOutput = fmt.Sprintf(jsonOutputFMT, graphOutput, conf.Type, symbdata)
	case c.JsonOutputB64:
		b64dot := base64.StdEncoding.EncodeToString([]byte(graphOutput))
		jsonOutput = fmt.Sprintf(jsonOutputFMT, b64dot, conf.Type, symbdata)

	case c.JsonOutputGZB64:
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		if _, err := gz.Write([]byte(graphOutput)); err != nil {
			return "", errors.New("gzip failed")
		}
		if err := gz.Close(); err != nil {
			return "", errors.New("gzip failed")
		}
		b64dot := base64.StdEncoding.EncodeToString(b.Bytes())
		jsonOutput = fmt.Sprintf(jsonOutputFMT, b64dot, conf.Type, symbdata)

	default:
		return "", errors.New("unknown output mode")
	}
	return jsonOutput, nil
}

func main() {
	conf, err := config.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(c.OSExitError)
	}
	if opt2num(conf.ConfValues.Type) == 0 {
		fmt.Printf("Unknown mode %s\n", conf.ConfValues.Type)
		os.Exit(-2)
	}

	t := connectToken{conf.ConfValues.DBDriver, conf.ConfValues.DBDSN}
	db := connectDb(&t)

	output, err := generateOutput(db, conf)
	if err != nil {
		fmt.Println("Internal error", err)
		os.Exit(-3)
	}
	fmt.Println(output)

}
