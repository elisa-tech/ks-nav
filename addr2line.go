package main

import (
	"fmt"
	"strings"
	"path/filepath"
	addr2line "github.com/elazarl/addr2line"
)

type workloads struct{
	Addr	uint64
	Name    string
	Query	string
	}

type Addr2line_items struct {
	Addr		uint64
	File_name	string
	}

type ins_f func(string)

var Addr2line_cache []Addr2line_items;

func addr2line_init(fn string) (chan workloads){
	a, err := addr2line.New(fn)
	if err != nil {
		panic( err)
		}
	adresses := make(chan workloads, 16)
	go workload(a, adresses, print_query)
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

func print_query(s string){
	fmt.Println(s)
}


func workload(a *addr2line.Addr2line, addresses chan workloads, insert_func ins_f/*,db*/){  //db needs to be added
	var e	workloads
	var qready string

	for {
		e = <-addresses
//		fmt.Printf("--->%s(0x%08x)\n",  e.Name, e.Addr)
		hit, val := in_cache(e.Addr, Addr2line_cache)
		if hit {
			qready=fmt.Sprintf(e.Query, val)
			insert_func(/*db,*/qready)
			}
		rs, _ := a.Resolve(e.Addr)
		if len(rs)==0 {
			fmt.Printf("no results resolving 0x%08x, giving up!\n", e.Addr)
			continue
			}
		for _, a:=range rs{
			fmt.Println(a.Function)
			if a.Function == strings.ReplaceAll(e.Name, "sym.", "") {
				qready=fmt.Sprintf(e.Query, filepath.Clean(a.File))
				insert_func(/*db,*/ qready)
				}
			}
		}
}


func spawn_query(addr uint64, name string, addresses chan workloads, query string) {
//	fmt.Printf("***>0x%08x\n",addr);
	addresses <- workloads{addr, name, query}
}
