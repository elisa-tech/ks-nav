package main
import (
	"fmt"
	"os"
	"strconv"
	)

func main() {
	var prod = map[string]int{}
	var visited []int
	var entry_name string

	cache := make(map[int][]Entry)
	cache2 := make(map[int]Entry)
	cache3 := make(map[string]string)


        start, err := strconv.Atoi(os.Args[1])
        if err != nil {
                panic(err)
        }
	t:=Connect_token{"dbs.hqhome163.com",5432,"alessandro","<password>","kernel_bin"}
	db:=Connect_db(&t)

//	Navigate(db, 153735, nil, prod)
	fmt.Println("digraph G {")
	entry, err := get_entry_by_id(db, start, cache2)
		if err!=nil {
			entry_name="Unknown";
			} else {
				entry_name=entry.Symbol
				}

	Navigate(db, start, entry_name, &visited, prod, Cache{cache, cache2, cache3}, PRINT_SUBSYS)
	fmt.Println("}")
}
