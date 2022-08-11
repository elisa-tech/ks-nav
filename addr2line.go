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
		e = <-addresses
		fmt.Printf("workload %s(0x%08x)\n", e.Name, e.Addr)
		if e.Name!= "None" {
			fmt.Printf("workload - resolve %s(0x%08x)\n", e.Name, e.Addr)
			rs, _ := a.Resolve(e.Addr)
			if len(rs)==0 {
				qready=fmt.Sprintf(e.Query, "NONE")
				}
			fmt.Printf("workload got reolution %s(0x%08x)\n", e.Name, e.Addr)
			fmt.Println(rs)
			for _, a:=range rs{
				qready=fmt.Sprintf(e.Query, filepath.Clean(a.File))
				if a.Function == strings.ReplaceAll(e.Name, "sym.", "") {
					fmt.Println("workload ", qready)
					break
					}
				}
			/*go*/ insert_func(e.DB, qready, false)
			} else {
				insert_func(e.DB, e.Query, false)
				}
	}
}


func spawn_query(db *sql.DB, addr uint64, name string, addresses chan workloads, query string) {
	fmt.Println("spawn_query: ", name)
	addresses <- workloads{addr, name, query, db}
}
