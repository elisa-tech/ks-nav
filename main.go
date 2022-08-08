package main
import (
	"fmt"
	"strings"
	r2 "github.com/radareorg/r2pipe-go"
)

func main(){
//	var s string



	fmt.Println("create stripped version")
	strip("/usr/bin/strip","./vmlinux.debug")


	addresses:=addr2line_init("vmlinux.debug")

/*	spawn_query(0xffffffff81008b80, addresses, "insert into pippo values ('%s')")
	fmt.Println("----",s)
	time.Sleep(5 * time.Second)
	fmt.Println("----",s)
*/
	r2p, err := r2.NewPipe("./vmlinux")
	if err != nil {
		panic(err)
		}
	fmt.Println("initialize analysis")

	init_fw(r2p)
	funcs_data := get_all_funcdata(r2p)
	//fmt.Println(funcs_data)

	for _, a :=range funcs_data{
//		fmt.Printf("######### %s, 0x%08x\n",a.Name, a.Offset)
		spawn_query(a.Offset, strings.ReplaceAll(a.Name, "sym.", ""), addresses, fmt.Sprintf("insert into file values('%%s');insert into symbols values ('%s',%d)",a.Name, a.Offset))
		}
}
