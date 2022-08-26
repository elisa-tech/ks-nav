package main
import (
	"fmt"
	"os"
//	"strconv"
	)

func main() {
	var prod = map[string]int{}
	var visited []int
	var entry_name string

	cache := make(map[int][]Entry)
	cache2 := make(map[int]Entry)
	cache3 := make(map[string]string)


        conf, err := args_parse(cmd_line_item_init())
        if err!=nil {
                fmt.Println("Kernel symbol fetcher")
                print_help(cmd_line_item_init());
                os.Exit(-1)
                }

	t:=Connect_token{ conf.DBURL, conf.DBPort,  conf.DBUser, conf.DBPassword, conf.DBTargetDB}
	db:=Connect_db(&t)

	start, err:=sym2num(db, conf.Symbol, conf.Instance)
	if err!=nil{
		fmt.Println("symbol not found")
		os.Exit(-2)
		}

	fmt.Println("digraph G {")
	entry, err := get_entry_by_id(db, start, conf.Instance, cache2)
		if err!=nil {
			entry_name="Unknown";
			} else {
				entry_name=entry.Symbol
				}

	Navigate(db, start, entry_name, &visited, prod, conf.Instance, Cache{cache, cache2, cache3}, conf.Mode)
	fmt.Println("}")
}
