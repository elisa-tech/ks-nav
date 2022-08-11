package main
import (
	"fmt"
	"strings"
	"time"
	r2 "github.com/radareorg/r2pipe-go"
)

func main(){
	var cache	[]xref_cache

	fmt.Println("create stripped version")
	strip("/usr/bin/strip","./vmlinux.debug")


	addresses:=addr2line_init("vmlinux.debug")

	r2p, err := r2.NewPipe("./vmlinux")
	if err != nil {
		panic(err)
		}
	fmt.Println("initialize analysis")
	fmt.Println(r2p)

	init_fw(r2p)
	funcs_data := get_all_funcdata(r2p)
	t:=Connect_token{ "dbs.hqhome163.com",5432,"alessandro","<password>","kernel_bin"}
	db:=Connect_db(&t)


	//first iteration fills symbols and files tables

	for _, a :=range funcs_data{
		fmt.Printf("main cycle: %s 0x%08x enter\n",a.Name, a.Offset)
		fmt.Printf("main cycle: %s 0x%08x before ifx\n", a.Name, a.Offset)
		if strings.Contains(a.Name, "sym.") {
			fmt.Printf("main cycle: %s 0x%08x in the if\n",a.Name, a.Offset)
			fmtstring:=fmt.Sprintf(
					"insert into files (file_name) Select '%%[1]s' Where not exists (select * from files where file_name='%%[1]s');"+
					"insert into symbols (symbol_name, address, file_ref_id) select '%[1]s', '%[2]s', (select file_id from files where file_name='%%[1]s');"+
					"",
					strings.ReplaceAll(a.Name, "sym.", ""),
					fmt.Sprintf("0x%08x",a.Offset))

			fmt.Printf("main cycle: %s 0x%08x raw query %s\n",a.Name, a.Offset, fmtstring)
			spawn_query(
				db,
				a.Offset, strings.ReplaceAll(a.Name, "sym.", ""),
				addresses,
				fmtstring)
			}
		}

//	fmt.Println("---------------")
//	fmt.Println(funcs_data)
//	fmt.Println("---------------")
	//second iteration fills xrefs table
	time.Sleep(5 * time.Second)
	fmt.Println(r2p)
	for _, a :=range funcs_data{
		if strings.Contains(a.Name, "sym.") {
//			xrefs:=Getxrefs(r2p, a.Offset, &cache)
			Move(r2p, a.Offset)
			xrefs:=remove_non_func(removeDuplicate(Getxrefs(r2p, a.Offset, &cache)),funcs_data)
			fmt.Printf("func_data iteration: 0x%08x, %s", a.Offset, a.Name)
			fmt.Println(xrefs)
			for _, l :=range xrefs {
				fmt.Printf("deps: 0x%08x\n", l)
				spawn_query(
					db,
					0,
					"None",
					addresses,
					fmt.Sprintf(
						"insert into xrefs (caller, callee) select (Select symbol_id from symbols where address ='0x%08x'), (Select symbol_id from symbols where address ='0x%08x');"+
						"",
						a.Offset,
						l))
				}

			}
		}


}

