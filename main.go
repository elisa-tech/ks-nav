package main
import (
	"fmt"
	"strings"
	r2 "github.com/radareorg/r2pipe-go"
	"github.com/cheggaaa/pb/v3"
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

	init_fw(r2p)
	funcs_data := get_all_funcdata(r2p)
	t:=Connect_token{ "dbs.hqhome163.com",5432,"alessandro","<password>","kernel_bin"}
	db:=Connect_db(&t)

	count:=len(funcs_data)
	bar := pb.StartNew(count)

	//first iteration fills symbols and files tables
	fmt.Println("collecting symbols & files")
	for _, a :=range funcs_data{
		bar.Increment()
		if strings.Contains(a.Name, "sym.") {
			fmtstring:=fmt.Sprintf(
					"insert into files (file_name) Select '%%[1]s' Where not exists (select * from files where file_name='%%[1]s');"+
					"insert into symbols (symbol_name, address, file_ref_id) select '%[1]s', '%[2]s', (select file_id from files where file_name='%%[1]s');"+
					"",
					strings.ReplaceAll(a.Name, "sym.", ""),
					fmt.Sprintf("0x%08x",a.Offset))

			spawn_query(
				db,
				a.Offset, strings.ReplaceAll(a.Name, "sym.", ""),
				addresses,
				fmtstring)
			}
		}
	bar.Finish()
	bar = pb.StartNew(count*4)	//assuming 4 as callee/caller ratio
	fmt.Println("Collecting xref")
	for _, a :=range funcs_data{
		if strings.Contains(a.Name, "sym.") {
			Move(r2p, a.Offset)
			xrefs:=remove_non_func(removeDuplicate(Getxrefs(r2p, a.Offset, &cache)),funcs_data)
			for _, l :=range xrefs {
				bar.Increment()
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
bar.Finish()

}

