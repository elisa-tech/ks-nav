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
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"regexp"
	"sort"
	"strings"
)

// Const values for configuration Mode field.
const (
	printAll    int = 1
	printSubsys     = 2
)

// Sql connection configuration
type connectToken struct {
	host   string
	port   int
	user   string
	pass   string
	DBName string
}

type entry struct {
	symID     int
	symbol    string
	exported  bool
	entryType string
	subSys    []string
	fn        string
}

type edge struct {
	caller int
	callee int
}

type cache struct {
	successors map[int][]entry
	entries    map[int]entry
	subSys     map[string]string
}

// Connects the target db and returns the handle
func connectDB(t *connectToken) *sql.DB {
	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", (*t).host, (*t).port, (*t).user, (*t).pass, (*t).DBName)
	db, err := sql.Open("postgres", psqlConn)
	if err != nil {
		panic(err)
	}
	return db
}

// Returns function details from a given id
func getEntryByID(db *sql.DB, symbol_id int, instance int, cache map[int]entry) (entry, error) {
	var e entry
	var s sql.NullString

	if e, ok := cache[symbol_id]; ok {
		return e, nil
	}

	query := "select symbol_id, symbol_name, subsys_name, file_name from " +
		"(select * from symbols, files where symbols.symbol_file_ref_id=files.file_id and symbols.symbol_instance_id_ref=$2) as dummy " +
		"left outer join tags on dummy.symbol_file_ref_id=tags.tag_file_ref_id where symbol_id=$1 and symbol_instance_id_ref=$2"
	rows, err := db.Query(query, symbol_id, instance)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&e.symID, &e.symbol, &s, &e.fn); err != nil {
			fmt.Println("this error hit3")
			fmt.Println(err)
			return e, err
		}
		if s.Valid {
			e.subSys = append(e.subSys, s.String)
		}
	}
	if err = rows.Err(); err != nil {
		fmt.Println("this error hit")
		return e, err
	}
	cache[symbol_id] = e
	return e, nil
}

// Returns the list of successors (called function) for a given function
func getSuccessorsByID(db *sql.DB, symbol_id int, instance int, cache cache) ([]entry, error) {
	var e edge
	var res []entry

	if res, ok := cache.successors[symbol_id]; ok {
		return res, nil
	}

	query := "select caller, callee from xrefs where caller =$1 and xref_instance_id_ref=$2"
	rows, err := db.Query(query, symbol_id, instance)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&e.caller, &e.callee); err != nil {
			fmt.Println("this error hit1 ")
			return nil, err
		}
		successors, _ := getEntryByID(db, e.callee, instance, cache.entries)
		res = append(res, successors)
	}
	if err = rows.Err(); err != nil {
		fmt.Println("this error hit2 ")
		return nil, err
	}
	cache.successors[symbol_id] = res
	return res, nil
}

// Return id an item is already in the list
func notIn(list []int, v int) bool {

	for _, a := range list {
		if a == v {
			return false
		}
	}
	return true
}

// Removes duplicates resulting by the exploration of a call tree
func removeDuplicate(list []entry) []entry {

	sort.SliceStable(list, func(i, j int) bool { return list[i].symID < list[j].symID })
	allKeys := make(map[int]bool)
	res := []entry{}
	for _, item := range list {
		if _, value := allKeys[item.symID]; !value {
			allKeys[item.symID] = true
			res = append(res, item)
		}
	}
	return res
}

// Given a function returns the lager subsystem it belongs
func getSubsysFromSymbolName(db *sql.DB, symbol string, instance int, subsystemsCache map[string]string) (string, error) {
	var res string

	if res, ok := subsystemsCache[symbol]; ok {
		return res, nil
	}
	query := "select subsys_name from (select count(*)as cnt, subsys_name from tags where subsys_name in (select subsys_name from symbols, " +
		"tags where symbols.symbol_file_ref_id=tags.tag_file_ref_id and symbols.symbol_name=$1 and symbols.symbol_instance_id_ref=$2) " +
		"group by subsys_name order by cnt desc) as tbl;"

	rows, err := db.Query(query, symbol, instance)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&res); err != nil {
			fmt.Println("this error hit1 ")
			return "", err
		}
	}
	subsystemsCache[symbol] = res
	return res, nil
}

// Returns the id of a given function name
func sym2num(db *sql.DB, symb string, instance int) (int, error) {
	var res int = 0
	var cnt int = 0
	query := "select symbol_id from symbols where symbols.symbol_name=$1 and symbols.symbol_instance_id_ref=$2"
	rows, err := db.Query(query, symb, instance)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		cnt++
		if err := rows.Scan(&res); err != nil {
			fmt.Println("this error hit7")
			fmt.Println(err)
			return res, err
		}
	}
	if cnt != 1 {
		return res, errors.New("id is not unique")
	}
	return res, nil
}

// Checks if a given function needs to be explored
func notExcluded(symbol string, excluded []string) bool {

	for _, s := range excluded {
		if match, _ := regexp.MatchString(s, symbol); match {
			return false
		}
	}
	return true
}

// Computes the call tree of a given function name
func navigate(db *sql.DB, symbol_id int, parentDisplay string, visited *[]int, prod map[string]int, instance int, cache cache, mode int, excluded []string, depth uint, maxDepth uint, dotFmt string, output *string) {
	var tmp, s, l, ll, r string
	var depthInc uint = 0

	*visited = append(*visited, symbol_id)
	l = parentDisplay
	successors, err := getSuccessorsByID(db, symbol_id, instance, cache)
	successors = removeDuplicate(successors)
	if err == nil {
		for _, curr := range successors {
			entry, err := getEntryByID(db, curr.symID, instance, cache.entries)
			if err != nil {
				r = "Unknown"
			} else {
				r = entry.symbol
			}
			switch mode {
			case printAll:
				s = fmt.Sprintf(dotFmt, l, r)
				ll = r
				depthInc = 1
				break
			case printSubsys:
				if tmp, err = getSubsysFromSymbolName(db, r, instance, cache.subSys); r != tmp {
					if tmp != "" {
						r = tmp
					} else {
						r = "UNDEFINED SUBSYSTEM"
					}
				}
				if l != r {
					s = fmt.Sprintf(dotFmt, l, r)
					depthInc = 1
				} else {
					s = ""
				}
				ll = r
				break

			}
			if _, ok := prod[s]; ok {
				prod[s]++
			} else {
				prod[s] = 1
				if s != "" {
					(*output) = (*output) + s
				}
			}

			if notIn(*visited, curr.symID) {
				if notExcluded(entry.symbol, excluded) && (maxDepth == 0 || (maxDepth > 0 && depth < maxDepth)) {
					navigate(db, curr.symID, ll, visited, prod, instance, cache, mode, excluded, depth+depthInc, maxDepth, dotFmt, output)
				}
			}
		}
	}
}

// Returns the subsystem list associated with a given function name
func symbSubsys(db *sql.DB, symbList []int, instance int, cache cache) (string, error) {
	var out string
	var res string

	for _, symbID := range symbList {
		//resolve symb
		symb, _ := getEntryByID(db, symbID, instance, cache.entries)
		out = out + fmt.Sprintf("{\"FuncName\":\"%s\", \"subsystems\":[", symb.symbol)
		query := fmt.Sprintf("select subsys_name from tags where tag_file_ref_id= (select symbol_file_ref_id from symbols where symbol_id=%d);", symbID)
		rows, err := db.Query(query)
		if err != nil {
			return "", errors.New("symbSubsys query failed")
		}
		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&res); err != nil {
				return "", errors.New("symbSubsys query browsing failed")
			}
			out = out + fmt.Sprintf("\"%s\",", res)
		}
		out = strings.TrimSuffix(out, ",") + "]},"
	}
	out = strings.TrimSuffix(out, ",")
	return out, nil
}
