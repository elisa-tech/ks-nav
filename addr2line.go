package main

import (
	"fmt"
	"strings"
	"database/sql"
	"path/filepath"
	addr2line "github.com/elazarl/addr2line"
)

type workloads struct{
	Addr	uint64
	Name    string
	Query	string
	Note	string
	DB	*sql.DB
	}

type Addr2line_items struct {
	Addr		uint64
	File_name	string
	}

type ins_f func(*sql.DB, string, bool)

var Addr2line_cache []Addr2line_items;

func addr2line_init(fn string) (chan workloads){
	a, err := addr2line.New(fn)
	if err != nil {
		panic( err)
		}
	adresses := make(chan workloads, 16)
//	go workload(a, adresses, print_query)
	go workload(a, adresses, Insert_data)
	return adresses
}
func in_cache(Addr uint64, Addr2line_cache []Addr2line_items)(bool, string){
	for _,a := range Addr2line_cache {
		if a.Addr == Addr {
			return true, a.File_name
			}
		}
	return false, ""
}
/*
func print_query(s string){
	fmt.Println(s)
}
*/

func workload(a *addr2line.Addr2line, addresses chan workloads, insert_func ins_f){
	var e	workloads
	var qready string

	for {
		if goroutine_nr>0 {
			e = <-addresses

			rs, _ := a.Resolve(e.Addr)
			if len(rs)==0 {
				if e.Note=="" {
					fmt.Printf("no results resolving %s(0x%08x), giving up!\n", e.Name, e.Addr)
					continue
					}
				qready=fmt.Sprintf(e.Query, e.Note)
				/*go*/ insert_func(e.DB, qready, false)
				}
			for _, a:=range rs{
				if a.Function == strings.ReplaceAll(e.Name, "sym.", "") {
					qready=fmt.Sprintf(e.Query, filepath.Clean(a.File))
					/*go*/ insert_func(e.DB, qready, false)
					break
					}
				}
			}
		}
}


func spawn_query(db *sql.DB, addr uint64, name string, addresses chan workloads, query string, note string) {
	addresses <- workloads{addr, name, query, note, db}
}
