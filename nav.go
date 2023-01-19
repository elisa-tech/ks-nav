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
)

const (
	graphOnly	int	= iota
	jsonOutputPlain
	jsonOutputB64
	jsonOutputGZB64
	outputLast
)

const	jsonOutputFMT	string = "{\"graph\": \"%s\",\"graph_type\":\"%s\",\"symbols\": [%s]}"

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

func generateOutput(db *sql.DB, conf *configuration) (string, error) {
	var graphOutput string
	var jsonOutput string
	var prod = map[string]int{}
	var visited []int
	var entryName string
	var output string

	cache1 := make(map[int][]entry)
	cache2 := make(map[int]entry)
	cache3 := make(map[string]string)

	start, err := sym2num(db, (*conf).Symbol, (*conf).Instance)
	if err != nil {
		fmt.Println("symbol not found")
		return "", err
	}

	graphOutput = fmtDotHeader[opt2num((*conf).Jout)]
	entry, err := getEntryByID(db, start, (*conf).Instance, cache2)
	if err != nil {
		entryName = "unknown"
		return "", err
	} else {
		entryName = entry.symbol
	}

	navigate(db, start, entryName, &visited, prod, (*conf).Instance, cache{cache1, cache2, cache3}, (*conf).Mode, (*conf).Excluded, 0, (*conf).MaxDepth, fmtDot[opt2num((*conf).Jout)], &output)
	graphOutput = graphOutput + output
	graphOutput = graphOutput + "}"

	symbdata, err := symbSubsys(db, visited, (*conf).Instance, cache{cache1, cache2, cache3})
	if err != nil {
		return "", err
	}

	switch opt2num((*conf).Jout) {
	case graphOnly:
		jsonOutput = graphOutput
	case jsonOutputPlain:
		jsonOutput = fmt.Sprintf(jsonOutputFMT, graphOutput, (*conf).Jout, symbdata)
	case jsonOutputB64:
		b64dot := base64.StdEncoding.EncodeToString([]byte(graphOutput))
		jsonOutput = fmt.Sprintf(jsonOutputFMT, b64dot, (*conf).Jout, symbdata)

	case jsonOutputGZB64:
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		if _, err := gz.Write([]byte(graphOutput)); err != nil {
			return "", errors.New("gzip failed")
		}
		if err := gz.Close(); err != nil {
			return "", errors.New("gzip failed")
		}
		b64dot := base64.StdEncoding.EncodeToString(b.Bytes())
		jsonOutput = fmt.Sprintf(jsonOutputFMT, b64dot, (*conf).Jout, symbdata)

	default:
		return "", errors.New("unknown output mode")
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
	if opt2num(conf.Jout) == 0 {
		fmt.Printf("unknown Mode %s\n", conf.Jout)
		os.Exit(-2)
	}
	t := connectToken{conf.DBURL, conf.DBPort, conf.DBUser, conf.DBPassword, conf.DBTargetDB}
	db := connectDB(&t)

	output, err := generateOutput(db, &conf)
	if err != nil {
		fmt.Println("internal error", err)
		os.Exit(-3)
	}
	fmt.Println(output)

}
