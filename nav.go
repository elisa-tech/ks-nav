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
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	graphOnly int = iota
	jsonOutputPlain
	jsonOutputB64
	jsonOutputGZB64
	outputLast
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
	"digraph G {\n",
	"digraph G {\\\\\\n",
	"digraph G {\n",
	"digraph G {\n",
}

var fmtDotNodeHighlightWSymb = "\"%[1]s\" [shape=record style=\"rounded,filled,bold\" fillcolor=yellow label=\"%[1]s|%[2]s\"]\n"
var fmtDotNodeHighlightWoSymb = "\"%[1]s\" [shape=record style=\"rounded,filled,bold\" fillcolor=yellow label=\"%[1]s\"]\n"

var fmtNodeDefault = "node [shape=\"box\"];"

func opt2num(s string) int {
	var opt = map[string]int{
		"graphOnly":       1,
		"jsonOutputPlain": 2,
		"jsonOutputB64":   3,
		"jsonOutputGZB64": 4,
	}
	val, ok := opt[s]
	if !ok {
		return 0
	}
	return val
}

func decorateLine(l string, r string, adjm []adjM) string {
	var res string = " [label=\""

	for _, item := range adjm {
		if (item.l.subsys == l) && (item.r.subsys == r) {
			tmp := fmt.Sprintf("%s([%s]%s),\\n", item.r.symbol, item.r.addressRef, item.r.sourceRef)
			if !strings.Contains(res, tmp) {
				res = res + fmt.Sprintf("%s([%s]%s),\\n", item.r.symbol, item.r.addressRef, item.r.sourceRef)
			}
		}
	}
	res = res + "\"]"
	return res
}

func decorate(dot_str string, adjm []adjM) string {
	var res string

	dot_body := strings.Split(dot_str, "\n")
	for i, line := range dot_body {
		split := strings.Split(line, "->")
		if len(split) == 2 {
			res = res + dot_body[i] + decorateLine(strings.TrimSpace(strings.Replace(split[0], "\"", "", -1)), strings.TrimSpace(strings.Replace(split[1], "\"", "", -1)), adjm) + "\n"
		}
	}
	return res
}

func generateOutput(db *sql.DB, conf *configuration) (string, error) {
	var graphOutput string
	var jsonOutput string
	var prod = map[string]int{}
	var visited []int
	var entryName string
	var output string
	var adjm []adjM

	cache := make(map[int][]entry)
	cache2 := make(map[int]entry)
	cache3 := make(map[string]string)

	start, err := sym2num(db, (*conf).symbol, (*conf).instance)
	if err != nil {
		fmt.Println("Symbol not found")
		return "", err
	}

	graphOutput = fmtDotHeader[opt2num((*conf).jout)]
	entry, err := getEntryById(db, start, (*conf).instance, cache2)
	if err != nil {
		entryName = "Unknown"
		return "", err
	} else {
		entryName = entry.symbol
	}
	start_subsys, _ := getSubsysFromSymbolName(db, entryName, (*conf).instance, cache3)
	if start_subsys == "" {
		start_subsys = SUBSYS_UNDEF
	}

	if ((*conf).mode == printTargeted) && len((*conf).targetSybsys) == 0 {
		targ_subsys_tmp, err := getSubsysFromSymbolName(db, (*conf).symbol, (*conf).instance, cache3)
		if err != nil {
			panic(err)
		}
		(*conf).targetSybsys = append((*conf).targetSybsys, targ_subsys_tmp)
	}

	navigate(db, start, node{start_subsys, entryName, "enty point", "0x0"}, (*conf).targetSybsys, &visited, &adjm, prod, (*conf).instance, Cache{cache, cache2, cache3}, (*conf).mode, (*conf).excludedAfter, (*conf).excludedBefore, 0, (*conf).maxDepth, fmtDot[opt2num((*conf).jout)], &output)

	if ((*conf).mode == printSubsysWs) || ((*conf).mode == printTargeted) {
		output = decorate(output, adjm)
	}

	graphOutput = graphOutput + output
	if (*conf).mode == printTargeted {
		for _, i := range (*conf).targetSybsys {
			if cache3[(*conf).symbol] == i {
				graphOutput = graphOutput + fmt.Sprintf(fmtDotNodeHighlightWSymb, i, (*conf).symbol)
			} else {
				graphOutput = graphOutput + fmt.Sprintf(fmtDotNodeHighlightWoSymb, i)
			}
		}
	}
	graphOutput = graphOutput + "}"

	symbdata, err := symbSubsys(db, visited, (*conf).instance, Cache{cache, cache2, cache3})
	if err != nil {
		return "", err
	}

	switch opt2num((*conf).jout) {
	case graphOnly:
		jsonOutput = graphOutput
	case jsonOutputPlain:
		jsonOutput = fmt.Sprintf(jsonOutputFMT, graphOutput, (*conf).jout, symbdata)
	case jsonOutputB64:
		b64dot := base64.StdEncoding.EncodeToString([]byte(graphOutput))
		jsonOutput = fmt.Sprintf(jsonOutputFMT, b64dot, (*conf).jout, symbdata)

	case jsonOutputGZB64:
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		if _, err := gz.Write([]byte(graphOutput)); err != nil {
			return "", errors.New("Gzip failed")
		}
		if err := gz.Close(); err != nil {
			return "", errors.New("Gzip failed")
		}
		b64dot := base64.StdEncoding.EncodeToString(b.Bytes())
		jsonOutput = fmt.Sprintf(jsonOutputFMT, b64dot, (*conf).jout, symbdata)

	default:
		return "", errors.New("Unknown output mode")
	}
	return jsonOutput, nil
}

func main() {

	conf, err := argsParse(cmdLineItemInit())
	if err != nil {
		if err.Error() != "dummy" {
			fmt.Println(err.Error())
		}
		printHelp(cmdLineItemInit())
		os.Exit(-1)
	}
	if opt2num(conf.jout) == 0 {
		fmt.Printf("Unknown mode %s\n", conf.jout)
		os.Exit(-2)
	}
	t := connectToken{conf.dbUrl, conf.dbPort, conf.dbUser, conf.dbPassword, conf.dbTargetDB}
	db := connectDb(&t)

	output, err := generateOutput(db, &conf)
	if err != nil {
		fmt.Println("Internal error", err)
		os.Exit(-3)
	}
	fmt.Println(output)

}
