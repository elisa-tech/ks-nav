/*
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *
 *   Name: kern_bin_db - Kernel source code analysis tool database creator
 *   Description: Parses kernel source tree and binary images and builds the DB
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
	"database/sql"
	"fmt"
	addr2line "github.com/elazarl/addr2line"
	"path/filepath"
	"strings"
	"sync"
)

type workloads struct {
	Addr  uint64
	Name  string
	Query string
	DB    *sql.DB
}

// Context type
type Context struct {
	instance    *addr2line.Addr2line
	ch_workload chan workloads
}

// Caches item elements
type Addr2line_items struct {
	Addr      uint64
	File_name string
}

// Commandline handle functions prototype
type ins_f func(*sql.DB, string, bool)

var mu sync.Mutex

func addr2line_init(fn string) (*addr2line.Addr2line, chan workloads) {
	a, err := addr2line.New(fn)
	if err != nil {
		panic(err)
	}
	adresses := make(chan workloads, 16)
	go workload(a, adresses, Insert_data)
	return a, adresses
}

func resolve_addr(a *addr2line.Addr2line, address uint64) string {
	var res string = ""
	mu.Lock()
	rs, _ := a.Resolve(address)
	mu.Unlock()
	if len(rs) == 0 {
		res = "NONE"
	}
	for _, a := range rs {
		res = fmt.Sprintf("%s:%d", filepath.Clean(a.File), a.Line)
	}
	return res
}

func workload(a *addr2line.Addr2line, addresses chan workloads, insert_func ins_f) {
	var e workloads
	var qready string

	for {
		e = <-addresses
		switch e.Name {
		case "None":
			insert_func(e.DB, e.Query, false)
			break
		default:
			mu.Lock()
			rs, _ := a.Resolve(e.Addr)
			mu.Unlock()
			if len(rs) == 0 {
				qready = fmt.Sprintf(e.Query, "NONE")
			}
			for _, a := range rs {
				qready = fmt.Sprintf(e.Query, filepath.Clean(a.File))
				if a.Function == strings.ReplaceAll(e.Name, "sym.", "") {
					break
				}
			}
			insert_func(e.DB, qready, false)
			break
		}
	}
}

func spawn_query(db *sql.DB, addr uint64, name string, addresses chan workloads, query string) {
	addresses <- workloads{addr, name, query, db}
}
