package main
import (
//	"fmt"
	"os"
	"strconv"
	)

func main() {
	var prod = map[string]int{}
        start, err := strconv.Atoi(os.Args[1])
        if err != nil {
                panic(err)
        }
	t:=Connect_token{"dbs.hqhome163.com",5432,"alessandro","<password>","kernel"}
	db:=Connect_db(&t)

//	Navigate(db, 153735, nil, prod)
	Navigate(db, start, nil, prod)
}
