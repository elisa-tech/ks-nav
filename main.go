package main
import (
	"fmt"
	"time"
)

func main(){
	var s string

	addresses:=addr2line_init("vmlinux.debug")
	get_fn(0xffffffff81008b80, addresses, &s)
	fmt.Println("----",s)
	time.Sleep(5 * time.Second)
	fmt.Println("----",s)
}
