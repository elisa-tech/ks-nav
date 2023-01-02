/*
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *
 *   Name: kern_bin_db - Kernel source code analysis tool database creator
 *   Description: Parses kernel source tree and binary images and builds the DB
 *
 *   Author: Alessandro Carminati <acarmina@redhat.com>
 *   Author: Maurizio Papini <mpapini@redhat.com>
 *
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *
 *   Copyright (c) 2022 Red Hat, Inc. All rights reserved.
 *
 *   This copyrighted material is made available to anyone wishing
 *   to use, modify, copy, or redistribute it subject to the terms
 *   and conditions of the GNU General Public License version 2.
 *
 *   This program is distributed in the hope that it will be
 *   useful, but WITHOUT ANY WARRANTY; without even the implied
 *   warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
 *   PURPOSE. See the GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public
 *   License along with this program; if not, write to the Free
 *   Software Foundation, Inc., 51 Franklin Street, Fifth Floor,
 *   Boston, MA 02110-1301, USA.
 *
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 */
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cheggaaa/pb/v3"
	r2 "github.com/radareorg/r2pipe-go"
)

// Bitfield configuration mode constants
const (
	ENABLE_SYBOLSNFILES   = 1
	ENABLE_XREFS          = 2
	ENABLE_MAINTAINERS    = 4
	ENABLE_VERSION_CONFIG = 8
)

func main() {
	var cache []xref_cache
	var r2p *r2.Pipe
	var bar *pb.ProgressBar
	var funcs_data []func_data
	var err error
	var count int
	var id int
	var addr2line_prefix string = ""

	conf, err := args_parse(cmd_line_item_init())
	if err != nil {
		fmt.Println("Kernel symbol fetcher")
		print_help(cmd_line_item_init())
		os.Exit(-1)
	}
	fmt.Println("create stripped version")
	strip(conf.StripBin, conf.LinuxWDebug, conf.LinuxWODebug)
	context := addr2line_init(conf.LinuxWDebug)
	t := Connect_token{conf.DBURL, conf.DBPort, conf.DBUser, conf.DBPassword, conf.DBTargetDB}
	db := Connect_db(&t)
	if conf.Mode&(ENABLE_VERSION_CONFIG) != 0 {
		config, _ := get_FromFile(conf.KConfig_fn)
		makefile, _ := get_FromFile(conf.KMakefile)
		v, err := get_version(makefile)
		if err != nil {
			panic(err)
		}
		fmt.Println(v)
		q := fmt.Sprintf("insert into instances (version_string, note) values ('%d.%d.%d%s', '%s');", v.Version, v.Patchlevel, v.Sublevel, v.Extraversion, conf.Note)
		id = Insert_datawID(db, q)
		kconfig := parse_config(config)
		fmt.Println("store config")
		bar = pb.StartNew(len(kconfig))
		for key, value := range kconfig {
			q := fmt.Sprintf("insert into configs (config_symbol, config_value, config_instance_id_ref) values ('%s', '%s', %d);", key, value, id)
			bar.Increment()
			spawn_query(db, 0, "None", context.ch_workload, q)
		}
		bar.Finish()
	}

	if conf.Mode&(ENABLE_SYBOLSNFILES|ENABLE_XREFS) != 0 {
		r2p, err = r2.NewPipe(conf.LinuxWODebug)
		if err != nil {
			panic(err)
		}
		q := fmt.Sprintf("insert into files (file_name, file_instance_id_ref) select 'NoFile',%d;", id)
		spawn_query(db, 0, "None", context.ch_workload, q)
		q = fmt.Sprintf("insert into symbols (symbol_name,symbol_address,symbol_type,symbol_file_ref_id,symbol_instance_id_ref) "+
			"select (select 'Indirect call'), '0x00000000', 'indirect', (select file_id from files where file_name ='NoFile' and file_instance_id_ref=%[1]d), %[1]d;", id)
		spawn_query(db, 0, "None", context.ch_workload, q)
		q = fmt.Sprintf("insert into tags (subsys_name, tag_file_ref_id, tag_instance_id_ref) select (select 'Indirect Calls'), "+
			"(select file_id from files where file_name='NoFile' and file_instance_id_ref=%[1]d), %[1]d;", id)
		spawn_query(db, 0, "None", context.ch_workload, q)
		fmt.Println("initialize analysis")
		init_fw(r2p)
		funcs_data = get_all_funcdata(r2p)
	}

	if conf.Mode&ENABLE_SYBOLSNFILES != 0 {
		count = len(funcs_data)
		bar = pb.StartNew(count)

		fmt.Println("collecting symbols & files")
		for _, a := range funcs_data {
			bar.Increment()
			symbtype := "direct"
			if a.Indirect {
				symbtype = "indirect"
			}
			if strings.Contains(a.Name, "sym.") || a.Indirect {
				fmtstring := fmt.Sprintf(
					"insert into files (file_name, file_instance_id_ref) Select '%%[1]s', %[1]d Where not exists "+
						"(select * from files where file_name='%%[1]s' and file_instance_id_ref=%[1]d);"+
						"insert into symbols (symbol_name, symbol_address, symbol_type, symbol_file_ref_id, symbol_instance_id_ref) "+
						"select '%[2]s', '%[3]s', '%[4]s', (select file_id from files where file_name='%%[1]s' and file_instance_id_ref=%[1]d), %[1]d;"+
						"",
					id,
					strings.ReplaceAll(a.Name, "sym.", ""),
					fmt.Sprintf("0x%08x", a.Offset),
					symbtype,
				)
				spawn_query(
					db,
					a.Offset,
					strings.ReplaceAll(a.Name, "sym.", ""),
					context.ch_workload,
					fmtstring)
			}

			// query for addr2line file prefix
			if a.Name == "sym.start_kernel" {
				var start_kernel_file_tail string = "init/main.c"
				start_kernel_file := strings.Split(resolve_addr(context, a.Offset), ":")
				if start_kernel_file[0] == "NONE" {
					panic("Error resolving start_kernel!")
				}
				if len(start_kernel_file[0]) > len(start_kernel_file_tail) {
					addr2line_prefix = start_kernel_file[0][:len(start_kernel_file)-len(start_kernel_file_tail)]
				}
			}
		}
		bar.Finish()
	}
	if conf.Mode&ENABLE_XREFS != 0 {
		fmt.Println("Collecting indrcalls")
		indcl := get_indirect_calls(r2p, funcs_data)
		fmt.Println("Collecting xref")
		bar = pb.StartNew(count)
		for _, a := range funcs_data {
			bar.Increment()
			if strings.Contains(a.Name, "sym.") {
				Move(r2p, a.Offset)
				xrefs := remove_non_func(Getxrefs(r2p, a.Offset, indcl, funcs_data, &cache), funcs_data)
				for _, l := range xrefs {
					source_ref := resolve_addr(context, l.From)
					spawn_query(
						db,
						0,
						"None",
						context.ch_workload,
						fmt.Sprintf(
							"insert into xrefs (caller, callee, ref_addr, source_line, xref_instance_id_ref) "+
								"select (Select symbol_id from symbols where symbol_address ='0x%08[1]x' and symbol_instance_id_ref=%[3]d), "+
								"(Select symbol_id from symbols where symbol_address ='0x%08[2]x' and symbol_instance_id_ref=%[3]d limit 1), "+
								"'0x%08[5]x', "+
								"'%[4]s', "+
								"%[3]d;"+
								"",
							a.Offset,
							l.To,
							id,
							source_ref,
							l.From))
				}
			}
		}
		bar.Finish()
	}
	tags_query := fmt.Sprintf("insert into tags (subsys_name, tag_file_ref_id, tag_instance_id_ref) select '%%[1]s', "+
		"(select file_id from files where file_name='%[1]s%%[2]s' and file_instance_id_ref=%%[3]d) as fn_id, %%[3]d "+
		"WHERE EXISTS ( select file_id from files where file_name='%[1]s%%[2]s' and file_instance_id_ref=%%[3]d);", addr2line_prefix)
	if conf.Mode&ENABLE_MAINTAINERS != 0 {
		fmt.Println("Collecting tags")
		s, err := get_FromFile(conf.Maintainers_fn)
		if err != nil {
			panic(err)
		}
		ss := s[seek2data(s):]
		items := parse_maintainers(ss)
		queries := generate_queries(conf.Maintainers_fn, items, tags_query, id)
		bar = pb.StartNew(len(queries))
		for _, q := range queries {
			bar.Increment()
			spawn_query(db, 0, "None", context.ch_workload, q)
		}
		bar.Finish()
	}
}
