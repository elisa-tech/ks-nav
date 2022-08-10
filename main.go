package main
import (
	"fmt"
	"strings"
	r2 "github.com/radareorg/r2pipe-go"
)

var goroutine_nr int = 100

func main(){
	var note	string

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

	for _, a :=range funcs_data{
		note=""
		if strings.Contains(a.Name, "__SCT__tp_func_") {
			note="Tracepoint"
			}
		if strings.Contains(a.Name, "sym.") {
			spawn_query(
				db,
				a.Offset, strings.ReplaceAll(a.Name, "sym.", ""),
				addresses,
				fmt.Sprintf(
//					"begin;"+
//					"lock table files IN exclusive mode;"+
					"insert into files (file_name) Select '%%[1]s' Where not exists (select * from files where file_name='%%[1]s');"+
//					"commit;"+
//					"begin;"+
					"insert into symbols (symbol_name, address, file_ref_id) select '%[1]s', '%[2]s', (select file_id from files where file_name='%%[1]s');"+ // where not exists (select * from symbols where symbol_name ='%[1]s');"+
//					"commit;"+
					"",
					strings.ReplaceAll(a.Name, "sym.", ""),
					fmt.Sprintf("0x%08x",a.Offset)),
				note)
			}
		}
}
