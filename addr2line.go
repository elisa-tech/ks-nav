package main

import (
	"fmt"
	addr2line "github.com/elazarl/addr2line"
)

type workloads struct{
	Addr	uint64
	Dest	*string
	}



func addr2line_init(fn string) (chan workloads){
	a, err := addr2line.New(fn)
	if err != nil {
		panic( err)
		}
	adresses := make(chan workloads, 16)
	go workload(a, adresses)
	return adresses
}

func workload(a *addr2line.Addr2line, addresses chan workloads){
	var e	workloads
	for {
		e = <-addresses
		rs, _ := a.Resolve(e.Addr)
		if len(rs)!=1 {
			fmt.Print("multiple or no results resolving 0x%08x\n", e.Addr)
			if len(rs)==0 {
				continue
				}
			}
		*e.Dest=rs[0].File
		fmt.Println(rs[0].Function, "@", rs[0].File, rs[0].Line)
		}
}


func get_fn(addr uint64, addresses chan workloads, destination *string) {
	addresses <- workloads{addr, destination}
}
