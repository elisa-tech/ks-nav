/*
 * Copyright (c) 2022 Red Hat, Inc.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/DATA-DOG/go-sqlmock"
)

var _ = Describe("Psql Tests", func() {

	var db *sql.DB
	var mock sqlmock.Sqlmock
	var e entry

	BeforeEach(func() {
		e = entry{
			symbol:     "mysymbol",
			sourceRef:  "config.h",
			addressRef: "0",
			subsys:     []string{},
			symId:      1,
		}

		db, mock, _ = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	})

	AfterEach(func() {
		defer GinkgoRecover()
		defer db.Close()
	})

	When("connectDB", func() {
		// TODO: `psql.connectDB` fn refactor needed
	})

	When("getEntryById", func() {
		testQuery := "select symbol_id, symbol_name, subsys_name, file_name from " +
			"(select * from symbols, files where symbols.symbol_file_ref_id=files.file_id and symbols.symbol_instance_id_ref=0) as dummy " +
			"left outer join tags on dummy.symbol_file_ref_id=tags.tag_file_ref_id where symbol_id=0 and symbol_instance_id_ref=0"
		It("Should return a cached result", func() {
			cache := map[int]entry{
				1: e,
			}
			_entry, err := getEntryById(nil, 1, 1, cache)

			Expect(err).To(BeNil())
			Expect(_entry).To(Equal(e))
		})

		It("Should inform the user of an internal database error", func() {
			mock.ExpectExec(testQuery).
				WithArgs(0, 0).
				WillReturnError(fmt.Errorf("myerror"))
			mock.ExpectRollback()

			_entry, err := getEntryById(db, 0, 0, map[int]entry{})

			Expect(err).ToNot(BeNil())
			Expect(_entry).To(Equal(entry{}))
		})

		It("Should not return an empty entry without errors", func() {
			rows := sqlmock.NewRows([]string{
				"id",
				"symbol",
				"fn",
				"sourceRef",
				"addressRef",
				"subsys",
				"symId",
			})

			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			_, err := getEntryById(db, 0, 0, map[int]entry{})

			Expect(err).To(BeNil())
		})

		It("Should fail for a row scan", func() {
			rows := sqlmock.NewRows([]string{
				"symbol",
				"fn",
				"subsys",
				"symId",
			})
			rows.AddRow(nil, nil, nil, nil)

			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			_, err := getEntryById(db, 0, 0, map[int]entry{})

			Expect(err).ToNot(BeNil())
		})

		It("Should fail for a row error", func() {
			rows := sqlmock.NewRows([]string{
				"symbol",
				"fn",
				"subsys",
				"symId",
			})
			rows.AddRow("0", "1", "2", 3)
			rows.RowError(0, fmt.Errorf("row error"))

			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			_, err := getEntryById(db, 0, 0, map[int]entry{})

			Expect(err).ToNot(BeNil())
		})

		It("Should find and return an entry", func() {
			rows := sqlmock.NewRows([]string{
				"symbol",
				"fn",
				"subsys",
				"symId",
			})
			rows.AddRow("0", "1", "2", 3)

			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			_entry, err := getEntryById(db, 0, 0, map[int]entry{})

			expectedEntry := entry{
				symbol:     "1",
				fn:         "3",
				sourceRef:  "",
				addressRef: "",
				subsys:     []string{"2"},
				symId:      0,
			}
			Expect(err).To(BeNil())
			Expect(_entry).To(Equal(expectedEntry))
		})
	})

	When("getSuccessorsById", func() {

		testQuery := "select caller, callee, source_line, ref_addr from xrefs where caller = 0 and xref_instance_id_ref = 0"

		It("Should return cached sucessors", func() {
			cache := Cache{
				successors: map[int][]entry{
					1: {e},
				},
			}
			entries, err := getSuccessorsById(nil, 1, 1, cache)

			Expect(err).To(BeNil())
			Expect(entries).To(Equal([]entry{e}))
		})

		It("Should panic because of a query error", func() {
			mock.ExpectQuery(testQuery).
				WithArgs(0, 0).
				WillReturnError(fmt.Errorf("myerror"))
			mock.ExpectRollback()

			Expect(func() { getSuccessorsById(db, 0, 0, Cache{}) }).To(Panic())
		})

		It("Should not fail in case no records are found", func() {
			rows := sqlmock.NewRows([]string{
				"symbol",
				"fn",
				"subsys",
				"symId",
			})

			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			cache := Cache{
				successors: map[int][]entry{},
			}
			entries, err := getSuccessorsById(db, 0, 0, cache)

			Expect(err).To(BeNil())
			Expect(entries).To(BeNil())
		})

		It("Should fail for a row scan", func() {
			rows := sqlmock.NewRows([]string{
				"symbol",
				"fn",
				"subsys",
				"symId",
			})
			rows.AddRow(nil, nil, nil, nil)

			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			_, err := getSuccessorsById(db, 0, 0, Cache{})

			Expect(err).ToNot(BeNil())
		})

		It("Should fail for a row error", func() {
			rows := sqlmock.NewRows([]string{
				"symbol",
				"fn",
				"subsys",
				"symId",
			})
			rows.AddRow("0", "1", "2", 3)
			rows.RowError(0, fmt.Errorf("row error"))

			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			entries, err := getSuccessorsById(db, 0, 0, Cache{})

			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(fmt.Errorf("row error")))
			Expect(entries).To(BeNil())
		})

		It("Should return a list of entries", func() {
			rows := sqlmock.NewRows([]string{
				"symbol",
				"fn",
				"subsys",
				"symId",
			})
			rows.AddRow("0", "1", "2", 3)

			mock.
				ExpectQuery(testQuery).
				//WithArgs(0, 0).
				WillReturnRows(rows)
			mock.ExpectCommit()

			expected := entry{symbol: "", fn: "", sourceRef: "2", addressRef: "3", subsys: nil, symId: 0}
			cache := Cache{
				successors: map[int][]entry{},
			}
			entries, err := getSuccessorsById(db, 0, 0, cache)

			Expect(err).To(BeNil())
			Expect(entries).To(Equal([]entry{expected}))
		})
	})

	When("notIn", func() {
		It("Should return false if a value is not in the slice", func() {
			data := []int{
				1,
				3,
				5,
			}

			res := notIn(data, 2)

			Expect(res).To(BeTrue())
		})

		It("Should return true if a value is in the slice", func() {
			data := []int{
				1,
				3,
				5,
			}

			res := notIn(data, 1)

			Expect(res).To(BeFalse())
		})

		It("Should return true if using an empty slice", func() {
			data := []int{}

			res := notIn(data, 1)

			Expect(res).To(BeTrue())
		})
	})

	When("removeDuplicate", func() {
		It("Should return an ordered but unmodified slice", func() {
			entries := []entry{
				{symId: 3},
				{symId: 1},
			}

			res := removeDuplicate(entries)

			Expect(len(res)).To(Equal(2))
			Expect(res).To(Equal([]entry{{symId: 1}, {symId: 3}}))
		})

		It("Should return a new slice with no duplicated values", func() {
			entries := []entry{
				{symId: 1},
				{symId: 3},
				{symId: 1},
			}

			res := removeDuplicate(entries)

			Expect(len(res)).To(Equal(2))
			Expect(res).To(Equal([]entry{{symId: 1}, {symId: 3}}))
		})

		It("Should return an empty slice", func() {
			entries := []entry{}

			res := removeDuplicate(entries)

			Expect(len(res)).To(Equal(0))
			Expect(res).To(BeNil())
		})
	})

	When("getSubsysFromSymbolName", func() {
		testQuery := "select (select symbol_type from symbols where symbol_name='mysym_key' and symbol_instance_id_ref=0) as type, subsys_name from " +
			"(select count(*) as cnt, subsys_name from tags where subsys_name in (select subsys_name from symbols, " +
			"tags where symbols.symbol_file_ref_id=tags.tag_file_ref_id and symbols.symbol_name='mysym_key' and symbols.symbol_instance_id_ref=0) " +
			"group by subsys_name order by cnt desc) as tbl;"

		It("Should return a cached symbol", func() {
			cache := map[string]string{"mysym_key": "mysym_val"}
			sym, err := getSubsysFromSymbolName(nil, "mysym_key", 0, cache)

			Expect(err).To(BeNil())
			Expect(sym).To(Equal("mysym_val"))
		})

		It("Should panic because of a query error", func() {
			mock.ExpectQuery(testQuery).
				WithArgs("mysym_key", 0).
				WillReturnError(fmt.Errorf("myerror"))
			mock.ExpectRollback()
			cache := map[string]string{}

			Expect(func() { getSubsysFromSymbolName(db, "mysym_key", 0, cache) }).To(Panic())
		})

		It("Should return nil in case of no rows", func() {
			rows := sqlmock.NewRows([]string{
				"ty",
				"sub",
			})

			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			sym, err := getSubsysFromSymbolName(db, "mysym_key", 0, map[string]string{})

			Expect(err).To(BeNil())
			Expect(sym).To(Equal(""))

		})

		It("Should return a row scan error", func() {
			rows := sqlmock.NewRows([]string{
				"ty",
				"sub",
			})
			rows.AddRow("direct", nil)
			mock.
				ExpectQuery(testQuery).
				//WithArgs("subsys", 0).
				WillReturnRows(rows)
			mock.ExpectCommit()

			cache := map[string]string{}
			sym, err := getSubsysFromSymbolName(db, "mysym_key", 0, cache)

			Expect(err).ToNot(BeNil())
			Expect(sym).To(Equal(""))
		})

		It("Should return a row error", func() {
			rows := sqlmock.NewRows([]string{
				"ty",
				"sub",
			})
			rows.AddRow("direct", "subsys")
			rows.RowError(0, fmt.Errorf("row error"))

			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			cache := map[string]string{}
			sym, err := getSubsysFromSymbolName(db, "mysym_key", 0, cache)

			Expect(err).ToNot(BeNil())
			Expect(sym).To(Equal(""))
		})

		It("Should find and return a subsystem name", func() {
			rows := sqlmock.NewRows([]string{
				"ty",
				"sub",
			})
			rows.AddRow("direct", "subsys")

			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			cache := map[string]string{}
			sym, err := getSubsysFromSymbolName(db, "mysym_key", 0, cache)

			Expect(err).To(BeNil())
			Expect(sym).To(Equal("subsys"))
			Expect(len(cache)).To(Equal(1))
			Expect(cache["mysym_key"]).To(Equal("subsys"))
		})

		It("Should find and return a subsystem name with indirect type", func() {
			rows := sqlmock.NewRows([]string{
				"ty",
				"sub",
			})
			rows.AddRow("indirect", "subsys")

			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			cache := map[string]string{}
			sym, err := getSubsysFromSymbolName(db, "mysym_key", 0, cache)

			Expect(err).To(BeNil())
			Expect(sym).To(Equal("indirect"))
			Expect(len(cache)).To(Equal(1))
			Expect(cache["mysym_key"]).To(Equal("indirect"))
		})
	})

	When("sym2num", func() {
		testQuery := "select symbol_id from symbols where symbols.symbol_name='mysym_key' and symbols.symbol_instance_id_ref=0"

		It("Should panic because of a query error", func() {
			mock.ExpectQuery(testQuery).
				WithArgs("mysim", 0).
				WillReturnError(fmt.Errorf("myerror"))
			mock.ExpectRollback()

			Expect(func() { sym2num(db, "mysym_key", 0) }).To(Panic())
		})

		It("Should return a row scan error", func() {
			rows := sqlmock.NewRows([]string{
				"res",
			})
			rows.AddRow(nil)
			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			symid, err := sym2num(db, "mysym_key", 0)

			Expect(err).ToNot(BeNil())
			Expect(symid).To(Equal(-1))
		})

		It("Should return a row error", func() {
			rows := sqlmock.NewRows([]string{
				"res",
			})
			rows.AddRow(1)
			rows.RowError(0, fmt.Errorf("sym2num row error"))
			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			symid, err := sym2num(db, "mysym_key", 0)

			Expect(err).ToNot(BeNil())
			Expect(symid).To(Equal(-1))
		})

		It("Should not fail with no rows", func() {
			rows := sqlmock.NewRows([]string{
				"res",
			})
			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			symid, err := sym2num(db, "mysym_key", 0)

			Expect(err).ToNot(BeNil())
			Expect(symid).To(Equal(-1))
		})

		It("Should fail if more than one record using the same id", func() {
			rows := sqlmock.NewRows([]string{
				"res",
			})
			rows.AddRow(1)
			rows.AddRow(1)
			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			symid, err := sym2num(db, "mysym_key", 0)

			Expect(err).ToNot(BeNil())
			Expect(symid).To(Equal(1))
		})

		It("Should return the symbol id", func() {
			rows := sqlmock.NewRows([]string{
				"res",
			})
			rows.AddRow(42)
			mock.
				ExpectQuery(testQuery).
				WillReturnRows(rows)
			mock.ExpectCommit()

			symid, err := sym2num(db, "mysym_key", 0)

			Expect(err).To(BeNil())
			Expect(symid).To(Equal(42))
		})
	})

	When("notExcluded", func() {
		excluded := []string{"sym1", "sym2"}

		It("Should return false if item is in excluded slice", func() {
			res := notExcluded("sym1", excluded)

			Expect(res).To(BeFalse())
		})

		It("Should return true if item is not in excluded slice", func() {
			res := notExcluded("sym3", excluded)

			Expect(res).To(BeTrue())
		})

		It("Should return true if slice is empty", func() {
			res := notExcluded("sym1", []string{})

			Expect(res).To(BeTrue())
		})
	})

	When("navigate", func() {
		// TODO: `psql.navigate` fn refactor needed
	})

	Describe("intargets", func() {
		targets := []string{"n1", "n2"}

		When("There is a target match", func() {
			It("Should return true if n1 is a match", func() {
				res := intargets(targets, "n1", "n3")

				Expect(res).To(BeTrue())
			})

			It("Should return true if n2 is a match", func() {
				res := intargets(targets, "n3", "n2")

				Expect(res).To(BeTrue())
			})

			It("Should return true if both are a match", func() {
				res := intargets(targets, "n1", "n2")

				Expect(res).To(BeTrue())
			})
		})

		When("There no target match", func() {
			It("Should return false if neither is a match", func() {
				res := intargets(targets, "n3", "n5")

				Expect(res).To(BeFalse())
			})

			It("Should return false if the target slice is empty", func() {
				res := intargets([]string{}, "n1", "n2")

				Expect(res).To(BeFalse())
			})
		})
	})

	Describe("symbSubsys", func() {
		var symList []int
		var instance int
		var cache Cache
		entryIdTestQuery := "select symbol_id, symbol_name, subsys_name, file_name from " +
			"(select * from symbols, files where symbols.symbol_file_ref_id=files.file_id and symbols.symbol_instance_id_ref=0) as dummy " +
			"left outer join tags on dummy.symbol_file_ref_id=tags.tag_file_ref_id where symbol_id=0 and symbol_instance_id_ref=0"
		testQuery := "select subsys_name from tags where tag_file_ref_id= (select symbol_file_ref_id from symbols where symbol_id=0);"
		_getEntryById := func(commit bool) {
			entryRows := sqlmock.NewRows([]string{
				"symbol",
				"fn",
				"subsys",
				"symId",
			})
			entryRows.AddRow("0", "1", "2", 0)

			mock.ExpectQuery(entryIdTestQuery).
				WillReturnRows(entryRows)
			if commit {
				mock.ExpectCommit()
			}
		}

		BeforeEach(func() {
			symList = []int{0}
			instance = 0
			cache = Cache{
				entries: map[int]entry{},
			}
		})

		When("Using an empty symbol list", func() {
			It("Should neither fail or return data", func() {
				out, err := symbSubsys(db, []int{}, instance, cache)

				Expect(err).To(BeNil())
				Expect(out).To(Equal(""))
			})
		})

		When("A getEntryById error happens", func() {
			It("Should return an error if getEntryById fails", func() {
				mock.ExpectExec(entryIdTestQuery).
					//WithArgs(0, 0).
					WillReturnError(fmt.Errorf("getEntryById query error"))
				mock.ExpectRollback()

				out, err := symbSubsys(db, symList, instance, cache)
				isErrMatched := strings.Contains(
					fmt.Sprint(err),
					"symbSubsys::getEntryById error")

				Expect(err).ToNot(BeNil())
				Expect(isErrMatched).To(BeTrue())
				Expect(out).To(Equal(""))
			})
		})

		When("A database error happens", func() {
			It("Should fail in case of a db.Query error", func() {
				_getEntryById(true)

				mock.ExpectQuery(testQuery).
					//WithArgs(0).
					WillReturnError(errors.New("symbSubsys db query error"))
				mock.ExpectRollback()

				out, err := symbSubsys(db, symList, instance, cache)

				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprint(err)).To(Equal("symbSubsys: query failed"))
				Expect(out).To(Equal(""))
			})

			It("Should fail in case of a db.Scan error", func() {
				_getEntryById(false)

				rows := sqlmock.NewRows([]string{
					"subsys",
				})
				rows.AddRow(nil)

				mock.
					ExpectQuery(testQuery).
					WillReturnRows(rows)
				mock.ExpectCommit()

				out, err := symbSubsys(db, symList, instance, cache)

				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprint(err)).To(Equal("symbSubsys: error while scan query rows"))
				Expect(out).To(Equal(""))
			})
		})

		When("It succeeds", func() {
			It("Should return the function name and its subsystems", func() {
				_getEntryById(false)

				rows := sqlmock.NewRows([]string{
					"subsys",
				})
				rows.AddRow("mock")

				mock.
					ExpectQuery(testQuery).
					WillReturnRows(rows)
				mock.ExpectCommit()

				out, err := symbSubsys(db, symList, instance, cache)

				Expect(err).To(BeNil())
				Expect(out).To(Equal("{\"FuncName\":\"1\", \"subsystems\":[\"mock\"]}"))
			})
		})
	})
})
