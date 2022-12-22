package main

import (
	"fmt"
)

type xref struct {
	Type string `json: "type"`
	From uint64 `json: "from"`
	To   uint64 `json: "to"`
}

func removeDuplicate(intSlice []xref) []xref {
	var key uint64

	allKeys := make(map[uint64]bool)
	list := []xref{}
	for _, item := range intSlice {
		if item.To != 0 {
			key = item.To
		} else {
			key = item.From //key is used to distinguish the itmes, since indirectcall has always To==0 From is used
		}
		if _, value := allKeys[key]; !value {
			allKeys[key] = true
			list = append(list, item)
		}
	}
	return list
}

func main() {
	xref_test1 := []xref{ //should find no duplicates
		xref{"indirect", 0xffffffff826a98a7, 0}, xref{"indirect", 0xffffffff826a98ef, 0}, xref{"indirect", 0xffffffff826ab514, 0}, xref{"indirect", 0xffffffff826ab5f3, 0},
		xref{"direct", 0xffffffff82681b9a, 0xffffffff826831c4}, xref{"direct", 0xffffffff8268214a, 0xffffffff82685b2a}, xref{"direct", 0xffffffff8268215a, 0xffffffff826878ab}, xref{"direct", 0xffffffff8268226a, 0xffffffff82687591},
		xref{"direct", 0xffffffff826822aa, 0xffffffff82685dc1}, xref{"direct", 0xffffffff8268245b, 0xffffffff826826aa}, xref{"direct", 0xffffffff8268255b, 0xffffffff826833a5}, xref{"direct", 0xffffffff8268259b, 0xffffffff82680329},
		xref{"direct", 0xffffffff8268269b, 0xffffffff82687778}, xref{"direct", 0xffffffff8268275b, 0xffffffff82681f54}, xref{"direct", 0xffffffff8268295b, 0xffffffff82680948}, xref{"direct", 0xffffffff82682d7b, 0xffffffff8268062e},
		xref{"direct", 0xffffffff82682dcb, 0xffffffff82680a21}, xref{"direct", 0xffffffff82682f2b, 0xffffffff82680e85}, xref{"direct", 0xffffffff8268305b, 0xffffffff82682533}, xref{"direct", 0xffffffff8268315b, 0xffffffff82681357},
		xref{"direct", 0xffffffff8268335b, 0xffffffff82684240}, xref{"direct", 0xffffffff8268377b, 0xffffffff826801af}, xref{"indirect", 0xffffffff826ab60b, 0}, xref{"direct", 0xffffffff8268385b, 0xffffffff826838f8},
		xref{"direct", 0xffffffff82683b5b, 0xffffffff826873c6}, xref{"indirect", 0xffffffff826ab62f, 0}, xref{"indirect", 0xffffffff826ac49f, 0}, xref{"indirect", 0xffffffff826acecb, 0},
		xref{"indirect", 0xffffffff826ad6f3, 0}, xref{"indirect", 0xffffffff826ad993, 0}, xref{"indirect", 0xffffffff826af05b, 0}, xref{"indirect", 0xffffffff826aff13, 0},
		xref{"indirect", 0xffffffff826b091b, 0}, xref{"indirect", 0xffffffff826b0ab3, 0}, xref{"indirect", 0xffffffff826b0c57, 0}, xref{"indirect", 0xffffffff826b180f, 0},
		xref{"indirect", 0xffffffff826b2306, 0}, xref{"indirect", 0xffffffff826b30db, 0}, xref{"indirect", 0xffffffff826b34a7, 0}, xref{"direct", 0xffffffff82683c5b, 0xffffffff82682750},
		xref{"direct", 0xffffffff82683d3b, 0xffffffff82681e45}, xref{"direct", 0xffffffff82683e3b, 0xffffffff82683ce2}, xref{"direct", 0xffffffff82692001, 0xffffffff82683dc1}, xref{"direct", 0xffffffff82692243, 0xffffffff82684737},
		xref{"direct", 0xffffffff8269264a, 0xffffffff82684816}, xref{"direct", 0xffffffff82692a47, 0xffffffff82681f18}, xref{"indirect", 0xffffffff826b3b43, 0}, xref{"indirect", 0xffffffff826b43a7, 0},
		xref{"indirect", 0xffffffff826b448a, 0}, xref{"indirect", 0xffffffff826b4593, 0}, xref{"indirect", 0xffffffff826b47a2, 0}, xref{"indirect", 0xffffffff826ee65f, 0}, xref{"direct", 0xffffffff8269551e, 0xffffffff82686005},
		xref{"direct", 0xffffffff8269564a, 0xffffffff82682056}, xref{"direct", 0xffffffff826956ff, 0xffffffff82684fb0}, xref{"indirect", 0xffffffff826eeabf, 0}, xref{"indirect", 0xffffffff826ef91f, 0},
		xref{"indirect", 0xffffffff826efa87, 0}, xref{"indirect", 0xffffffff826efc17, 0}, xref{"direct", 0xffffffff8269aa6f, 0xffffffff826847a9}, xref{"indirect", 0xffffffff826efd7f, 0},
		xref{"direct", 0xffffffff8269ac97, 0xffffffff82687579}, xref{"indirect", 0xffffffff826f04af, 0}, xref{"direct", 0xffffffff8269aecb, 0xffffffff82686991}, xref{"indirect", 0xffffffff826f05c7, 0},
		xref{"indirect", 0xffffffff826f1157, 0}, xref{"indirect", 0xffffffff826f1427, 0}, xref{"indirect", 0xffffffff826f17e7, 0}, xref{"indirect", 0xffffffff826f18ff, 0},
		xref{"indirect", 0xffffffff826f1c47, 0}, xref{"indirect", 0xffffffff826f1cbf, 0}, xref{"indirect", 0xffffffff826f1e27, 0}, xref{"indirect", 0xffffffff826f27d7, 0},
		xref{"indirect", 0xffffffff826f2a2f, 0},
	}
	xref_test2 := []xref{ //should find 2 duplicates
		xref{"indirect", 0xffffffff826a98a7, 0}, xref{"indirect", 0xffffffff826a98ef, 0}, xref{"indirect", 0xffffffff826ab514, 0}, xref{"indirect", 0xffffffff826ab5f3, 0},
		xref{"direct", 0xffffffff82681b9a, 0xffffffff826831c4}, xref{"direct", 0xffffffff8268214a, 0xffffffff82685b2a}, xref{"direct", 0xffffffff8268215a, 0xffffffff826878ab}, xref{"direct", 0xffffffff8268226a, 0xffffffff82687591},
		xref{"direct", 0xffffffff826822aa, 0xffffffff82685dc1}, xref{"direct", 0xffffffff8268245b, 0xffffffff826826aa}, xref{"direct", 0xffffffff8268255b, 0xffffffff826833a5}, xref{"direct", 0xffffffff8268259b, 0xffffffff82680329},
		xref{"direct", 0xffffffff8268269b, 0xffffffff82687778}, xref{"direct", 0xffffffff8268275b, 0xffffffff82681f54}, xref{"direct", 0xffffffff8268295b, 0xffffffff82680948}, xref{"direct", 0xffffffff82682d7b, 0xffffffff8268062e},
		xref{"direct", 0xffffffff82682dcb, 0xffffffff82680a21}, xref{"direct", 0xffffffff82682f2b, 0xffffffff82680e85}, xref{"direct", 0xffffffff8268305b, 0xffffffff82682533}, xref{"direct", 0xffffffff8268315b, 0xffffffff82681357},
		xref{"direct", 0xffffffff8268335b, 0xffffffff82684240}, xref{"direct", 0xffffffff8268377b, 0xffffffff826801af}, xref{"indirect", 0xffffffff826ab60b, 0}, xref{"direct", 0xffffffff8268385b, 0xffffffff826838f8},
		xref{"direct", 0xffffffff82683b5b, 0xffffffff826873c6}, xref{"indirect", 0xffffffff826ab62f, 0}, xref{"indirect", 0xffffffff826ac49f, 0}, xref{"indirect", 0xffffffff826acecb, 0},
		xref{"indirect", 0xffffffff826ad6f3, 0}, xref{"indirect", 0xffffffff826ad993, 0}, xref{"indirect", 0xffffffff826af05b, 0}, xref{"indirect", 0xffffffff826aff13, 0},
		xref{"indirect", 0xffffffff826b091b, 0}, xref{"indirect", 0xffffffff826b0ab3, 0}, xref{"indirect", 0xffffffff826b0c57, 0}, xref{"indirect", 0xffffffff826b180f, 0},
		xref{"indirect", 0xffffffff826b2306, 0}, xref{"indirect", 0xffffffff826b30db, 0}, xref{"indirect", 0xffffffff826b34a7, 0}, xref{"direct", 0xffffffff82683c5b, 0xffffffff82682750},
		xref{"direct", 0xffffffff82683d3b, 0xffffffff82681e45}, xref{"direct", 0xffffffff82683e3b, 0xffffffff82683ce2}, xref{"direct", 0xffffffff82692001, 0xffffffff82683dc1}, xref{"direct", 0xffffffff82692243, 0xffffffff82684737},
		xref{"direct", 0xffffffff8269264a, 0xffffffff82684816}, xref{"direct", 0xffffffff82692a47, 0xffffffff82681f18}, xref{"indirect", 0xffffffff826b3b43, 0}, xref{"indirect", 0xffffffff826b43a7, 0},
		xref{"indirect", 0xffffffff826b448a, 0}, xref{"indirect", 0xffffffff826b4593, 0}, xref{"indirect", 0xffffffff826b47a2, 0}, xref{"indirect", 0xffffffff826ee65f, 0}, xref{"direct", 0xffffffff8269551e, 0xffffffff82686005},
		xref{"direct", 0xffffffff8269564a, 0xffffffff82682056}, xref{"direct", 0xffffffff826956ff, 0xffffffff82684fb0}, xref{"indirect", 0xffffffff826eeabf, 0}, xref{"indirect", 0xffffffff826ef91f, 0},
		xref{"indirect", 0xffffffff826efa87, 0}, xref{"indirect", 0xffffffff826efc17, 0}, xref{"direct", 0xffffffff8269aa6f, 0xffffffff826847a9}, xref{"indirect", 0xffffffff826efd7f, 0},
		xref{"direct", 0xffffffff8269ac97, 0xffffffff82687579}, xref{"indirect", 0xffffffff826f04af, 0}, xref{"direct", 0xffffffff8269aecb, 0xffffffff82686991}, xref{"indirect", 0xffffffff826f05c7, 0},
		xref{"indirect", 0xffffffff826f1157, 0}, xref{"indirect", 0xffffffff826f1427, 0}, xref{"indirect", 0xffffffff826f17e7, 0}, xref{"indirect", 0xffffffff826f18ff, 0},
		xref{"indirect", 0xffffffff826f1c47, 0}, xref{"indirect", 0xffffffff826f1cbf, 0}, xref{"indirect", 0xffffffff826f1e27, 0}, xref{"indirect", 0xffffffff826f27d7, 0},
		xref{"indirect", 0xffffffff826f2a2f, 0}, xref{"indirect", 0xffffffff826a98a7, 0}, xref{"direct", 0xffffffff82682766, 0xffffffff82681f54},
	}

	fmt.Printf("Test1 size=%d\n", len(xref_test1))
	res := removeDuplicate(xref_test1)
	fmt.Printf("removeDuplicate applied to xref_test1 new size =%d\n", len(res))
	fmt.Printf("Test2 size=%d\n", len(xref_test2))
	res = removeDuplicate(xref_test2)
	fmt.Printf("removeDuplicate applied to xref_test2 new size =%d\n", len(res))
}
