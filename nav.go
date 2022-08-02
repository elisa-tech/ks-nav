package main
import (
//	"fmt"
	"os"
	"strconv"
	)

func main() {
	var prod = map[string]int{}
	var visited []int

	cache := make(map[int][]Entry)
	cache2 := make(map[int]Entry)

        start, err := strconv.Atoi(os.Args[1])
        if err != nil {
                panic(err)
        }
	t:=Connect_token{"dbs.hqhome163.com",5432,"alessandro","<password>","kernel"}
	db:=Connect_db(&t)

//	Navigate(db, 153735, nil, prod)
	Navigate(db, start, &visited, prod, cache, cache2)
}
