package main
import (
	"fmt"
	"strings"
	r2 "github.com/radareorg/r2pipe-go"
	"github.com/cheggaaa/pb/v3"
)

const ENABLE_SYBOLSNFILES	= 1
const ENABLE_XREFS		= 2
const ENABLE_MAINTAINERS	= 4
type configuration struct {
	LinuxWDebug	string
	LinuxWODebug	string
	StripBin	string
	DBURL		string
	DBPort		int
	DBUser		string
	DBPassword	string
	DBTargetDB	string
	Maintainers_fn	string
	Mode		int
}



func main(){
	var cache	[]xref_cache
	var r2p		*r2.Pipe
	var bar		*pb.ProgressBar
	var funcs_data	[]func_data
	var err		error
	var count	int

	conf:=configuration{"vmlinux", "vmlinux.work", "/usr/bin/strip", "dbs.hqhome163.com",5432,"alessandro","<password>","kernel_bin", "MAINTAINERS", ENABLE_SYBOLSNFILES|ENABLE_XREFS|ENABLE_MAINTAINERS}
	fmt.Println("create stripped version")
	strip(conf.StripBin, conf.LinuxWDebug, conf.LinuxWODebug)
	addresses:=addr2line_init(conf.LinuxWDebug)


	if conf.Mode & (ENABLE_SYBOLSNFILES|ENABLE_XREFS) != 0 {
		r2p, err = r2.NewPipe(conf.LinuxWODebug)
		if err != nil {
			panic(err)
			}
		fmt.Println("initialize analysis")

		init_fw(r2p)
		funcs_data = get_all_funcdata(r2p)
		}

	t:=Connect_token{ conf.DBURL, conf.DBPort,  conf.DBUser, conf.DBPassword, conf.DBTargetDB}
	db:=Connect_db(&t)


	if conf.Mode & ENABLE_SYBOLSNFILES != 0 {
		count=len(funcs_data)
		bar = pb.StartNew(count)

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
		}
	if conf.Mode & ENABLE_XREFS != 0 {
		bar = pb.StartNew(count)
		fmt.Println("Collecting xref")
		for _, a :=range funcs_data{
			bar.Increment()
			if strings.Contains(a.Name, "sym.") {
				Move(r2p, a.Offset)
				xrefs:=remove_non_func(removeDuplicate(Getxrefs(r2p, a.Offset, &cache)),funcs_data)
				for _, l :=range xrefs {
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

	if conf.Mode & ENABLE_MAINTAINERS != 0 {
		fmt.Println("Collecting tags")
        	s,err:=get_FromFile(conf.Maintainers_fn)
	        if err!= nil {
	                panic(err)
        	        }
	        ss:=s[seek2data(s):]
        	items:=parse_maintainers(ss)
	        queries:=generate_queries(items, "insert into tags (subsys_name, file_ref_id) select '%[1]s', "+
        	                                "(select file_id from files where file_name='%[2]s') as fn_id "+
                	                        "WHERE EXISTS ( select file_id from files where file_name='%[2]s');")
		bar = pb.StartNew(len(queries))
        	for _,q :=range queries{
//			fmt.Println(q)
			bar.Increment()
	                spawn_query(db, 0, "None", addresses, q, )
        	        }
		bar.Finish()
		}
}
