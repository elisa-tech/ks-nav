package main
 
import (
	"database/sql"
	"fmt"
//	"strings"
	_ "github.com/lib/pq"
)
type Connect_token struct{
	Host	string
	Port	int
	User	string
	Pass	string
	Dbname	string
}

type Entry struct {
	Sym_id		int
	Symbol		string
	Exported	bool
	Type		string
	Subsys		[]string
	Fn		string
}

type Edge struct {
	Caller	int
	Callee	int
}

func Connect_db(t *Connect_token) (*sql.DB){
	fmt.Println("connect")
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", (*t).Host, (*t).Port, (*t).User, (*t).Pass, (*t).Dbname)
	db, err := sql.Open("postgres", psqlconn)
	if err!= nil {
		panic(err)
		}
	fmt.Println("connected")
	return db
}

func get_entry_by_id(db *sql.DB, symbol_id int)(Entry, error){
	var e		Entry
	var s		sql.NullString

//	fmt.Println("query")
	query:="select sym_id, symbol, exported, type, subsys, fn from (select * from symbols, kernel_file where symbols.fn_id=kernel_file.id) as dummy left outer join tags on dummy.fn_id=tags.fn_id where sym_id=$1"
	rows, err := db.Query(query, symbol_id)
	if err!= nil {
		panic(err)
		}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&e.Sym_id, &e.Symbol, &e.Exported, &e.Type, &s, &e.Fn); err != nil {
			fmt.Println("this error hit3")
			fmt.Println(err)
			return e, err
			}
//		fmt.Println(e)
		if s.Valid {
			e.Subsys=append(e.Subsys, s.String)
			}
		}
	if err = rows.Err(); err != nil {
		fmt.Println("this error hit")
        	return e, err
		}
	return e, nil
}

func get_successors_by_id(db *sql.DB, symbol_id int)([]Entry, error){
	var e		Edge
	var res		[]Entry

//	fmt.Println("query")
	query:="select caller, callee from calls where caller =$1"
	rows, err := db.Query(query, symbol_id)
	if err!= nil {
		panic(err)
		}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&e.Caller, &e.Callee); err != nil {
			fmt.Println("this error hit1 ")
			return nil, err
			}
		successors,_ := get_entry_by_id(db, e.Callee)
		res=append(res, successors )
		}
	if err = rows.Err(); err != nil {
		fmt.Println("this error hit2 ")
        	return nil, err
		}
	return res, nil
}

func Not_in(list []int, v int) bool {

        for _, a := range list {
                if a == v {
                        return false
                }
        }
        return true
}

func Navigate(db *sql.DB, symbol_id int, visited []int) {
	var rname, lname, lpname	string

	visited=append(visited, symbol_id)
	lname=""
	entry, err := get_entry_by_id(db, symbol_id)
	if err!=nil {
		rname="Unknown";
		}
	rname=entry.Symbol
	successors, err:=get_successors_by_id(db, symbol_id);
	if err==nil {
		for _, curr := range successors{
			entry, err := get_entry_by_id(db, curr.Sym_id)
		        if err!=nil {
 		               lname="Unknown";
                		}
		        lname=entry.Symbol
//			fmt.Printf("%d -> %d\n", symbol_id, curr.Sym_id)
			if lpname!=lname {
				fmt.Printf("%s -> %s\n", rname, lname)
				}
			lpname=lname
			if Not_in(visited, curr.Sym_id){
				Navigate(db, curr.Sym_id, visited)
				}
			}
//		fmt.Println("--------------")
		}
}
