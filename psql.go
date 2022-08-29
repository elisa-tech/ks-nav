package main
 
import (
	"database/sql"
	"fmt"
//	"strings"
	"regexp"
	"errors"
	"sort"
	_ "github.com/lib/pq"
)
const (
	PRINT_ALL int	= 1
	PRINT_SUBSYS	= 2
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
type Cache struct {
	Successors	map[int][]Entry
	Entries		map[int]Entry
	SubSys		map[string]string
}

var check int = 0
var chached int = 0





func Connect_db(t *Connect_token) (*sql.DB){
//	fmt.Println("connect")
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", (*t).Host, (*t).Port, (*t).User, (*t).Pass, (*t).Dbname)
	db, err := sql.Open("postgres", psqlconn)
	if err!= nil {
		panic(err)
		}
//	fmt.Println("connected")
	return db
}

func get_entry_by_id(db *sql.DB, symbol_id int, instance int,cache map[int]Entry)(Entry, error){
	var e		Entry
	var s		sql.NullString
	
//	fmt.Println("query")
	if e, ok := cache[symbol_id]; ok {
//		fmt.Println("--------------------------------------------cache2 hit!---------------------------------------------------")
		return e, nil
		}



	query:="select symbol_id, symbol_name, subsys_name, file_name from (select * from symbols, files where symbols.symbol_file_ref_id=files.file_id and symbols.symbol_instance_id_ref=$2) as dummy left outer join tags on dummy.symbol_file_ref_id=tags.tag_file_ref_id where symbol_id=$1 and symbol_instance_id_ref=$2"
//	fmt.Println(query,symbol_id,instance)
	rows, err := db.Query(query, symbol_id, instance)
	if err!= nil {
		panic(err)
		}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&e.Sym_id, &e.Symbol, &s, &e.Fn); err != nil {
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
	cache[symbol_id]=e
	return e, nil
}

func get_successors_by_id(db *sql.DB, symbol_id int, instance int, cache Cache )([]Entry, error){
	var e		Edge
	var res		[]Entry

//	fmt.Println("query")

	if res, ok := cache.Successors[symbol_id]; ok {
//		fmt.Println("--------------------------------------------cache hit!---------------------------------------------------")
		return res, nil
		}

	query:="select caller, callee from xrefs where caller =$1 and xref_instance_id_ref=$2"
	rows, err := db.Query(query, symbol_id, instance)
	if err!= nil {
		panic(err)
		}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&e.Caller, &e.Callee); err != nil {
			fmt.Println("this error hit1 ")
			return nil, err
			}
		successors,_ := get_entry_by_id(db, e.Callee, instance, cache.Entries)
		res=append(res, successors )
		}
	if err = rows.Err(); err != nil {
		fmt.Println("this error hit2 ")
        	return nil, err
		}
	cache.Successors[symbol_id]=res
//	fmt.Println(cache)
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


func removeDuplicate(list []Entry) []Entry {

	sort.SliceStable(list, func(i, j int) bool { return list[i].Sym_id < list[j].Sym_id })
        allKeys := make(map[int]bool)
        res := []Entry{}
        for _, item := range list {
                if _, value := allKeys[item.Sym_id]; !value {
                        allKeys[item.Sym_id] = true
                        res = append(res, item)
                        }
                }
        return res
}
func get_subsys_from_symbol_name(db *sql.DB, symbol string, instance int, subsytems_cache map[string]string)(string, error){
	var res string

	if res, ok := subsytems_cache[symbol]; ok {
		return res, nil
		}
	query:="select subsys_name from symbols, tags where symbols.symbol_file_ref_id=tags.tag_file_ref_id and symbols.symbol_name=$1 and symbols.symbol_instance_id_ref=$2;"
        rows, err := db.Query(query, symbol, instance)
        if err!= nil {
                panic(err)
                }
        defer rows.Close()

        for rows.Next() {
                if err := rows.Scan(&res); err != nil {
                        fmt.Println("this error hit1 ")
                        return "", err
                        }
		}
        subsytems_cache[symbol]=res
        return res, nil



}

func sym2num(db *sql.DB, symb string, instance int)(int, error){
	var 	res	int=0
	var	cnt 	int=0
	query:="select symbol_id from symbols where symbols.symbol_name=$1 and symbols.symbol_instance_id_ref=$2"
//	fmt.Println(query, symb, instance)
        rows, err := db.Query(query, symb, instance)
        if err!= nil {
                panic(err)
                }
        defer rows.Close()

        for rows.Next() {
		cnt++
                if err := rows.Scan(&res,); err != nil {
                        fmt.Println("this error hit7")
                        fmt.Println(err)
                        return res, err
                        }
		}
	if cnt!=1 {
		return res, errors.New("id is not unique")
		}
	return res, nil
}

func not_exluded(symbol string, excluded []string)bool{

	for _,s:=range excluded{
		if match,_ :=regexp.MatchString(s, symbol); match{
			return false
			}
		}
	return true
}


func Navigate(db *sql.DB, symbol_id int, parent_dispaly string, visited *[]int, prod map[string]int, instance int, cache Cache, mode int, excluded []string) {
	var tmp,s,l,ll,r	string


	*visited=append(*visited, symbol_id)
//	entry, err := get_entry_by_id(db, symbol_id, cache2)
	l=parent_dispaly
	successors, err:=get_successors_by_id(db, symbol_id, instance, cache);
	successors=removeDuplicate(successors)
	if err==nil {
		for _, curr := range successors{
			entry, err := get_entry_by_id(db, curr.Sym_id, instance, cache.Entries)
		        if err!=nil {
 		               r="Unknown";
                		} else {
					r=entry.Symbol
					}
			switch mode {
				case PRINT_ALL:
					s=fmt.Sprintf("\"%s\"->\"%s\"", l, r)
					ll=r
					break
				case PRINT_SUBSYS:
					if tmp, err=get_subsys_from_symbol_name(db,r, instance, cache.SubSys); r!=tmp {
						if tmp != "" {
							r=tmp
							} else {
								r="UNDEFINED SUBSYSTEM"
								}
						}
					if l!=r {
						s=fmt.Sprintf("\"%s\"->\"%s\"", l, r)
						} else {
							s="";
							}
					ll=r
					break

				}
//			fmt.Println(s)
			if _, ok := prod[s]; ok {
				prod[s]++
				} else {
					prod[s]=1
					if s!="" {
						fmt.Println(s)
						}
					}

			if Not_in(*visited, curr.Sym_id){
				if not_exluded(entry.Symbol, excluded){
					Navigate(db, curr.Sym_id, ll, visited, prod, instance, cache, mode, excluded)
					}
				}
			}
		}
}
