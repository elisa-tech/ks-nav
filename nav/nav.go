/*
 * Copyright (c) 2022 Red Hat, Inc.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"nav/config"
	c "nav/constants"
	"github.com/goccy/go-graphviz"
	"os"
	"strings"
)

const jsonOutputFMT string = "{\"graph\": \"%s\",\"graph_type\":\"%s\",\"symbols\": [edge%s]}"

var fmtDot = []string{
	"",
	"\"%s\"->\"%s\" [ edgeid = \"%d\"]; \n",
	"\\\"%s\\\"->\\\"%s\\\"; \\\\\\n",
	"\"%s\"->\"%s\"; \n",
	"\"%s\"->\"%s\"; \n",
	"",
}

var fmtDotHeader = []string{
	"",
	"digraph G {\nrankdir=LR; node [style=filled fillcolor=yellow]\n",
	"digraph G {\\\\\\nrankdir=\"LR\"\\\\\\n",
	"digraph G {\nrankdir=\"LR\"\n",
	"digraph G {\nrankdir=\"LR\"\n",
	"digraph G {\nlayout=\"fdp\"\noverlap=\"1:scalexy\"\nnode [shape=\"box\";style=filled;color=green];\n",
	"digraph G {\nlayout=\"fdp\"\noverlap=\"1:scalexy\"\nnode [shape=\"box\";style=filled;color=green];\n",
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

func do_graphviz(dot string, output_type c.OutIMode) error{
	var buf bytes.Buffer
	var format graphviz.Format

	switch output_type {
	case c.OPNG:
		format=graphviz.PNG
	case c.OJPG:
		format=graphviz.JPG
	case c.OSVG:
		format=graphviz.SVG
	default:
		return errors.New("Unknown format")
	}

	graph, _ := graphviz.ParseBytes([]byte(dot))
	g := graphviz.New()
	defer func() {
		if err := graph.Close(); err != nil {
			panic(err)
		}
		g.Close()
	}()
	g.Render(graph, format, &buf)
	binary.Write(os.Stdout, binary.LittleEndian, buf.Bytes())
	return nil
}

func generateOutput(d Datasource, cfg *config.Config) (string, error) {
	var graphOutput string
	var jsonOutput string
	var prod = map[string]int{}
	var visited []int
	var entryName string
	var output string
	var adjm []adjM

	archnum := 0
	conf := cfg.ConfValues

	start, err := d.sym2num(conf.Symbol, conf.DBInstance)
	if err != nil {
		fmt.Println("Symbol not found")
		return "", err
	}

	graphOutput = fmtDotHeader[conf.Mode]
	if conf.Mode <= c.PrintTargeted {
		entry, err := d.getEntryById(start, conf.DBInstance)
		if err != nil {
			return "", err
		} else {
			entryName = entry.symbol
		}
		startSubsys, _ := d.getSubsysFromSymbolName(entryName, conf.DBInstance)
		if startSubsys == "" {
			startSubsys = SUBSYS_UNDEF
		}

		if (conf.Mode == c.PrintTargeted) && len(conf.TargetSubsys) == 0 {
			targSubsysTmp, err := d.getSubsysFromSymbolName(conf.Symbol, conf.DBInstance)
			if err != nil {
				panic(err)
			}
			conf.TargetSubsys = append(conf.TargetSubsys, targSubsysTmp)
		}

		navigate(d, start, node{startSubsys, entryName, "entry point", "0x0"}, conf.TargetSubsys, &visited, &adjm, prod, conf.DBInstance, conf.Mode, conf.ExcludedAfter, conf.ExcludedBefore, 0, conf.MaxDepth, fmtDot[opt2num(conf.Type)], &output, &archnum)

		if (conf.Mode == c.PrintSubsysWs) || (conf.Mode == c.PrintTargeted) {
			output = decorate(output, adjm)
		}

		graphOutput += output
		if conf.Mode == c.PrintTargeted {
			for _, i := range conf.TargetSubsys {
				if d.GetExploredSubsystemByName(conf.Symbol) == i {
					graphOutput += fmt.Sprintf(fmtDotNodeHighlightWSymb, i, conf.Symbol)
				} else {
					graphOutput += fmt.Sprintf(fmtDotNodeHighlightWoSymb, i)
				}
			}
		}
	} else {
/*
		print " ##symb## [shape=house;style=filled;color=cyan;];"
		set = select * from nm_symbol where nm_sym_id in (select data_sym_id from data_xrefs where func_id in (select symbol_id from symbols where symbol_name ='##symb##' and symbol_instance_id_ref=##i##));
		for x in set {
		        print " "##x##" [shape=box;style=filled;color=green]; "
		}
		for x in set {
		        arch = select subsys_name, '##x##'  from tags where tag_file_ref_id in (select file_id from files where file_id in (select symbol_file_ref_id from symbols where symbol_id in (select func_id from data_xrefs where data_sym_id in (s>
		        for y in arch {
		                print " "##y1##" -> "##y2##"
		        }
		}
*/
		graphOutput += fmt.Sprintf(" \"%s\" [shape=house;style=filled;color=cyan;width=5, height=2, fixedsize=true];\n", conf.Symbol)
		gdata, err := d.symbGData(conf.Symbol, conf.DBInstance)
		if err!= nil {
			panic(err)
		}
		for _, i := range gdata {
			graphOutput += fmt.Sprintf("\"%s\" [shape=\"ellipse\";style=filled;color=orange;width=5, height=2, fixedsize=true];\n", i)
			tmp := d.symbGDataFuncOf(i, conf.DBInstance)
			for _, j := range tmp {
				graphOutput += fmt.Sprintf("%s\n", j)
			}
		}
	}
	graphOutput += "}"
	switch opt2num(conf.Type) {
	case c.GraphOnly:
		jsonOutput = graphOutput
	case c.JsonOutputPlain:
		symbdata, err := d.symbSubsys(visited, conf.DBInstance)
		if err != nil {
			return "", err
		}
		jsonOutput = fmt.Sprintf(jsonOutputFMT, graphOutput, conf.Type, symbdata)
	case c.JsonOutputB64:
		symbdata, err := d.symbSubsys(visited, conf.DBInstance)
		if err != nil {
			return "", err
		}
		b64dot := base64.StdEncoding.EncodeToString([]byte(graphOutput))
		jsonOutput = fmt.Sprintf(jsonOutputFMT, b64dot, conf.Type, symbdata)

	case c.JsonOutputGZB64:
		var b bytes.Buffer
		symbdata, err := d.symbSubsys(visited, conf.DBInstance)
		if err != nil {
			return "", err
		}
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
	d := &SqlDB{}
	err = d.init(&t)
	if err != nil {
		panic(err)
	}

	output, err := generateOutput(d, conf)
	if err != nil {
		fmt.Println("Internal error", err)
		os.Exit(-3)
	}
	if conf.ConfValues.Graphviz != c.OText {
		err = do_graphviz(output, conf.ConfValues.Graphviz);
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println(output)
	}
}
