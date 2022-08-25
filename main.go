package main
import (
	"fmt"
	"strings"
	r2 "github.com/radareorg/r2pipe-go"
	"github.com/cheggaaa/pb/v3"
)

const (
	ENABLE_SYBOLSNFILES	= 1
	ENABLE_XREFS		= 2
	ENABLE_MAINTAINERS	= 4
	ENABLE_VERSION_CONFIG	= 8
	)

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
	KConfig_fn	string
	KMakefile	string
	Mode		int
	Note		string
}



func main(){
	var cache	[]xref_cache
	var r2p		*r2.Pipe
	var bar		*pb.ProgressBar
	var funcs_data	[]func_data
	var err		error
	var count	int
	var id		int

/*
	conf:=configuration{
				"vmlinux",
				"vmlinux.work",
				"/usr/bin/strip",
				"dbs.hqhome163.com",
				5432,
				"alessandro",
				"<password>",
				"kernel_bin",
				"MAINTAINERS",
				"./include/generated/autoconf.h",
				"Makefile",
				ENABLE_SYBOLSNFILES|ENABLE_XREFS|ENABLE_MAINTAINERS|ENABLE_VERSION_CONFIG,
				"upstream"
				}
*/
	conf:=configuration{
				"vmlinux",
				"vmlinux.work",
				"/usr/bin/aarch64-linux-gnu-strip",
				"dbs.hqhome163.com",
				5432,
				"alessandro",
				"<password>",
				"kernel_bin",
				"MAINTAINERS",
				"./include/generated/autoconf.h",
				"Makefile",
				ENABLE_SYBOLSNFILES|ENABLE_XREFS|ENABLE_MAINTAINERS|ENABLE_VERSION_CONFIG,
				"upstream",
				}
	fmt.Println("create stripped version")
	strip(conf.StripBin, conf.LinuxWDebug, conf.LinuxWODebug)
	addresses:=addr2line_init(conf.LinuxWDebug)

	t:=Connect_token{ conf.DBURL, conf.DBPort,  conf.DBUser, conf.DBPassword, conf.DBTargetDB}
	db:=Connect_db(&t)

	if conf.Mode & (ENABLE_VERSION_CONFIG) != 0 {
	        config, _ := get_FromFile(conf.KConfig_fn)
	        makefile, _ := get_FromFile(conf.KMakefile)
	        v, err:= get_version(makefile)
	        if err!=nil {
        	        panic(err)
                	}
	        fmt.Println(v)
		q:=fmt.Sprintf("insert into instances (version_string, note) values ('%d.%d.%d%s', '%s');", v.Version, v.Patchlevel, v.Sublevel, v.Extraversion, conf.Note)
		id=Insert_datawID(db, q)
        	kconfig:=parse_config(config)
		fmt.Println("store config")
                bar = pb.StartNew(len(kconfig))
                for key,value :=range kconfig{
			q:=fmt.Sprintf("insert into configs (config_symbol, config_value, instance_id_ref) values ('%s', '%s', %d);", key, value, id)
                        bar.Increment()
                        spawn_query(db, 0, "None", addresses, q )
                        }
                bar.Finish()

		}


	if conf.Mode & (ENABLE_SYBOLSNFILES|ENABLE_XREFS) != 0 {
		r2p, err = r2.NewPipe(conf.LinuxWODebug)
		if err != nil {
			panic(err)
			}
		q:=fmt.Sprintf("insert into files (file_name, instance_id_ref) select 'NoFile',%d;", id)
		spawn_query(db, 0, "None", addresses, q, )
		q=fmt.Sprintf("insert into symbols (symbol_name,address,file_ref_id,instance_id_ref) select (select 'Indirect call'), '0x00000000', (select file_id from files where file_name ='NoFile'), %d;", id)
		spawn_query(db, 0, "None", addresses, q, )
		q=fmt.Sprintf("insert into tags (subsys_name, file_ref_id, instance_id_ref) select (select 'Indirect Calls'), (select file_id from files where file_name='NoFile'), %d;", id)
		spawn_query(db, 0, "None", addresses, q, )

		fmt.Println("initialize analysis")

		init_fw(r2p)
		funcs_data = get_all_funcdata(r2p)
		}

	if conf.Mode & ENABLE_SYBOLSNFILES != 0 {
		count=len(funcs_data)
		bar = pb.StartNew(count)

		//first iteration fills symbols and files tables
		fmt.Println("collecting symbols & files")
		for _, a :=range funcs_data{
			bar.Increment()
			if strings.Contains(a.Name, "sym.") {
				fmtstring:=fmt.Sprintf(
						"insert into files (file_name, instance_id_ref) Select '%%[1]s', %[1]d Where not exists (select * from files where file_name='%%[1]s');"+
						"insert into symbols (symbol_name, address, file_ref_id, instance_id_ref) select '%[2]s', '%[3]s', (select file_id from files where file_name='%%[1]s'), %[1]d;"+
						"",
						id,
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
		fmt.Println("Collecting indrcalls")
		indcl:=get_indirect_calls(r2p, funcs_data)
		fmt.Println("Collecting xref")
		bar = pb.StartNew(count)
		for _, a :=range funcs_data{
			bar.Increment()
			if strings.Contains(a.Name, "sym.") {
				Move(r2p, a.Offset)
				xrefs:=remove_non_func(removeDuplicate(Getxrefs(r2p, a.Offset, indcl, funcs_data, &cache)),funcs_data)
				for _, l :=range xrefs {
/*
					if l==0 {
						s:=fmt.Sprintf(
							"insert into xrefs (caller, callee, instance_id_ref) select (Select symbol_id from symbols where address ='0x%08x'), (Select symbol_id from symbols where address ='0x%08x'), %d;"+
							"",
							a.Offset,
							l,
							id)
						fmt.Println(s)
						}
*/
					spawn_query(
						db,
						0,
						"None",
						addresses,
						fmt.Sprintf(
							"insert into xrefs (caller, callee, instance_id_ref) select (Select symbol_id from symbols where address ='0x%08x'), (Select symbol_id from symbols where address ='0x%08x'), %d;"+
							"",
							a.Offset,
							l,
							id))
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
	        queries:=generate_queries(items, "insert into tags (subsys_name, file_ref_id, instance_id_ref) select '%[1]s', "+
        	                                "(select file_id from files where file_name='%[2]s') as fn_id, %d "+
                	                        "WHERE EXISTS ( select file_id from files where file_name='%[2]s');", id)
		bar = pb.StartNew(len(queries))
        	for _,q :=range queries{
//			fmt.Println(q)
			bar.Increment()
	                spawn_query(db, 0, "None", addresses, q, )
        	        }
		bar.Finish()
		}
}
