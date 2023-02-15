/*
 * Copyright (c) 2022 Red Hat, Inc.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

package main

import (
	"database/sql"
	"errors"
	"fmt"
	c "nav/constants"
	"regexp"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const SUBSYS_UNDEF = "The REST"

// Parent node.
type node struct {
	subsys     string
	symbol     string
	sourceRef  string
	addressRef string
}
type adjM struct {
	l node
	r node
}

// Sql connection configuration.
type connectToken struct {
	DBDriver string
	DBDSN    string
}

type entry struct {
	symbol     string
	fn         string
	sourceRef  string
	addressRef string
	subsys     []string
	symId      int
}

type edge struct {
	sourceRef  string
	addressRef string
	caller     int
	callee     int
}

type Cache struct {
	successors map[int][]entry
	entries    map[int]entry
	subSys     map[string]string
}

// Connects the target db and returns the handle.
func connectDb(t *connectToken) *sql.DB {
	db, err := sql.Open(t.DBDriver, t.DBDSN)
	if err != nil {
		panic(err)
	}
	return db
}

// Returns function details from a given id.
func getEntryById(db *sql.DB, symbolId int, instance int, cache map[int]entry) (entry, error) {
	var e entry
	var s sql.NullString

	if e, ok := cache[symbolId]; ok {
		return e, nil
	}

	query := "select symbol_id, symbol_name, subsys_name, file_name from " +
		"(select * from symbols, files where symbols.symbol_file_ref_id=files.file_id and symbols.symbol_instance_id_ref=%[2]d) as dummy " +
		"left outer join tags on dummy.symbol_file_ref_id=tags.tag_file_ref_id where symbol_id=%[1]d and symbol_instance_id_ref=%[2]d"
	query = fmt.Sprintf(query, symbolId, instance)
	rows, err := db.Query(query)
	if err != nil {
		return entry{}, err
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil {
			err = closeErr
		}
	}()

	for rows.Next() {
		if err := rows.Scan(&e.symId, &e.symbol, &s, &e.fn); err != nil {
			fmt.Println("getEntryById: error while scan query rows")
			fmt.Println(err)
			return e, err
		}
		if s.Valid {
			e.subsys = append(e.subsys, s.String)
		}
	}
	if err = rows.Err(); err != nil {
		fmt.Println("getEntryById: error in access query rows")
		return e, err
	}
	cache[symbolId] = e
	return e, nil
}

// Returns the list of successors (called function) for a given function.
func getSuccessorsById(db *sql.DB, symbolId int, instance int, cache Cache) ([]entry, error) {
	var e edge
	var res []entry

	if res, ok := cache.successors[symbolId]; ok {
		return res, nil
	}

	query := "select caller, callee, source_line, ref_addr from xrefs where caller = %[1]d and xref_instance_id_ref = %[2]d"
	query = fmt.Sprintf(query, symbolId, instance)
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil {
			err = closeErr
		}
	}()

	for rows.Next() {
		if err := rows.Scan(&e.caller, &e.callee, &e.sourceRef, &e.addressRef); err != nil {
			fmt.Println("get_successors_by_id: error while scan query rows", err)
			return nil, err
		}
		successor, _ := getEntryById(db, e.callee, instance, cache.entries)
		successor.sourceRef = e.sourceRef
		successor.addressRef = e.addressRef
		res = append(res, successor)
	}
	if err = rows.Err(); err != nil {
		fmt.Println("get_successors_by_id: error in access query rows")
		return nil, err
	}
	cache.successors[symbolId] = res
	return res, nil
}

// Return id an item is already in the list.
func notIn(list []int, v int) bool {

	for _, a := range list {
		if a == v {
			return false
		}
	}
	return true
}

// Removes duplicates resulting by the exploration of a call tree.
func removeDuplicate(list []entry) []entry {

	sort.SliceStable(list, func(i, j int) bool { return list[i].symId < list[j].symId })
	allKeys := make(map[int]bool)
	var res []entry
	for _, item := range list {
		if _, value := allKeys[item.symId]; !value {
			allKeys[item.symId] = true
			res = append(res, item)
		}
	}
	return res
}

// Given a function returns the lager subsystem it belongs.
func getSubsysFromSymbolName(db *sql.DB, symbol string, instance int, subsytemsCache map[string]string) (string, error) {
	var ty, sub string

	if res, ok := subsytemsCache[symbol]; ok {
		return res, nil
	}
	query := "select (select symbol_type from symbols where symbol_name='%[1]s' and symbol_instance_id_ref=%[2]d) as type, subsys_name from " +
		"(select count(*) as cnt, subsys_name from tags where subsys_name in (select subsys_name from symbols, " +
		"tags where symbols.symbol_file_ref_id=tags.tag_file_ref_id and symbols.symbol_name='%[1]s' and symbols.symbol_instance_id_ref=%[2]d) " +
		"group by subsys_name order by cnt desc) as tbl;"

	query = fmt.Sprintf(query, symbol, instance)
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil {
			err = closeErr
		}
	}()

	for rows.Next() {
		if err := rows.Scan(&ty, &sub); err != nil {
			fmt.Println("get_subsys_from_symbol_name: error while scan query rows")
			return "", err
		}
	}

	if err = rows.Err(); err != nil {
		fmt.Println("getSubsysFromSymbolName: error in access query rows")
		return "", err
	}

	if sub == "" {
		return "", nil
	}

	if ty == "indirect" {
		sub = ty
	}
	subsytemsCache[symbol] = sub
	return sub, nil
}

// Returns the id of a given function name.
func sym2num(db *sql.DB, symb string, instance int) (int, error) {
	var res = -1
	var cnt = 0
	query := "select symbol_id from symbols where symbols.symbol_name='%[1]s' and symbols.symbol_instance_id_ref=%[2]d"
	query = fmt.Sprintf(query, symb, instance)
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil {
			err = closeErr
		}
	}()

	for rows.Next() {
		cnt++
		if err := rows.Scan(&res); err != nil {
			fmt.Println("sym2num: error while scan query rows")
			fmt.Println(err)
			return res, err
		}
	}

	if err = rows.Err(); err != nil {
		fmt.Println("sym2num: error in access query rows")
		return -1, err
	}

	if cnt != 1 {
		return res, errors.New("duplicate ID in the DB")
	}
	return res, nil
}

// Checks if a given function needs to be explored.
func notExcluded(symbol string, excluded []string) bool {
	for _, s := range excluded {
		if match, _ := regexp.MatchString(s, symbol); match {
			return false
		}
	}
	return true
}

// Computes the call tree of a given function name
// TODO: refactory needed:
// What is the problem: too many args.
// suggestion: New version with input and output structs.
func navigate(db *sql.DB, symbolId int, parentDispaly node, targets []string, visited *[]int, AdjMap *[]adjM, prod map[string]int, instance int, cache Cache, mode c.OutMode, excludedAfter []string, excludedBefore []string, depth int, maxdepth int, dotFmt string, output *string) {
	var tmp, s string
	var l, r, ll node

	*visited = append(*visited, symbolId)
	l = parentDispaly
	successors, err := getSuccessorsById(db, symbolId, instance, cache)
	if mode == c.PrintAll {
		successors = removeDuplicate(successors)
	}
	if err == nil {
		for _, curr := range successors {
			depthInc := 0
			if notExcluded(curr.symbol, excludedBefore) {
				r.symbol = curr.symbol
				r.sourceRef = curr.sourceRef
				r.addressRef = curr.addressRef
				tmp, _ = getSubsysFromSymbolName(db, r.symbol, instance, cache.subSys)
				if tmp == "" {
					r.subsys = SUBSYS_UNDEF
				}

				switch mode {
				case c.PrintAll:
					s = fmt.Sprintf(dotFmt, l.symbol, r.symbol)
					ll = r
					depthInc = 1
				case c.PrintSubsys, c.PrintSubsysWs, c.PrintTargeted:
					if tmp, _ = getSubsysFromSymbolName(db, r.symbol, instance, cache.subSys); r.subsys != tmp {
						if tmp != "" {
							r.subsys = tmp
						} else {
							r.subsys = SUBSYS_UNDEF
						}
					}

					if l.subsys != r.subsys {
						s = fmt.Sprintf(dotFmt, l.subsys, r.subsys)
						*AdjMap = append(*AdjMap, adjM{l, r})
						depthInc = 1
					} else {
						s = ""
					}
					ll = r
				default:
					panic(mode)
				}
				if _, ok := prod[s]; ok {
					prod[s]++
				} else {
					prod[s] = 1
					if s != "" {
						if (mode != c.PrintTargeted) || (intargets(targets, l.subsys, r.subsys)) {
							*output = (*output) + s
						}
					}
				}
				if notIn(*visited, curr.symId) {
					if (notExcluded(curr.symbol, excludedAfter) && notExcluded(curr.symbol, excludedBefore) ) && (maxdepth == 0 || ((maxdepth > 0) && (depth+depthInc < maxdepth))) {
						navigate(db, curr.symId, ll, targets, visited, AdjMap, prod, instance, cache, mode, excludedAfter, excludedBefore, depth+depthInc, maxdepth, dotFmt, output)
					}
				}
			}
		}
	}
}

// returns true if one of the nodes n1, n2 is a target node.
func intargets(targets []string, n1 string, n2 string) bool {

	for _, t := range targets {
		if (t == n1) || (t == n2) {
			return true
		}
	}
	return false
}

// Returns the subsystem list associated with a given function name.
func symbSubsys(db *sql.DB, symblist []int, instance int, cache Cache) (string, error) {
	var out string
	var res string

	for _, symbid := range symblist {
		// Resolve symb.
		symb, err := getEntryById(db, symbid, instance, cache.entries)
		if err != nil {
			return "", fmt.Errorf("symbSubsys::getEntryById error: %s", err)
		}
		out += fmt.Sprintf("{\"FuncName\":\"%s\", \"subsystems\":[", symb.symbol)
		query := fmt.Sprintf("select subsys_name from tags where tag_file_ref_id= (select symbol_file_ref_id from symbols where symbol_id=%d);", symbid)
		rows, err := db.Query(query)
		if err != nil {
			return "", errors.New("symbSubsys: query failed")
		}

		defer func() {
			closeErr := rows.Close()
			if err == nil {
				err = closeErr
			}
		}()

		for rows.Next() {
			if err := rows.Scan(&res); err != nil {
				return "", errors.New("symbSubsys: error while scan query rows")
			}
			out += fmt.Sprintf("\"%s\",", res)
		}
		out = strings.TrimSuffix(out, ",") + "]},"
	}
	out = strings.TrimSuffix(out, ",")
	return out, nil
}
