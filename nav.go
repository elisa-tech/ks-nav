package main
//import (
//	"fmt"
//	)


func main() {
	t:=Connect_token{"dbs.hqhome163.com",5432,"alessandro","<password>","kernel"}
	db:=Connect_db(&t)
/*é
	succ, err :=get_successors_by_id(db, 153735)
	if err!= nil {
		panic(err)
		}
	for _, e := range succ {
		fmt.Println(e, ",,,,,,,,,,,;;")
		}
*/
	Navigate(db, 153735)
}

