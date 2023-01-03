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
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const Insert_Instance_Q	string =	"insert into instances (version_string, note) values ('%d.%d.%d%s', '%s');"
const Insert_Config_Q	string =	"insert into configs (config_symbol, config_value, config_instance_id_ref) values ('%s', '%s', %d);"
const Insert_Files_Q	string =	"insert into files (file_name, file_instance_id_ref) select 'NoFile',%d;"
const Insert_Symbols_Q	string =	"insert into symbols (symbol_name,symbol_address,symbol_type,symbol_file_ref_id,symbol_instance_id_ref) "+"select (select 'Indirect call'), '0x00000000', 'indirect', (select file_id from files where file_name ='NoFile' and file_instance_id_ref=%[1]d), %[1]d;"
const Insert_Tags_Q	string =	"insert into tags (subsys_name, tag_file_ref_id, tag_instance_id_ref) select (select 'Indirect Calls'), "+"(select file_id from files where file_name='NoFile' and file_instance_id_ref=%[1]d), %[1]d;"
const Insert_Mixed_Q	string =	"insert into files (file_name, file_instance_id_ref) Select '%%[1]s', %[1]d Where not exists "+"(select * from files where file_name='%%[1]s' and file_instance_id_ref=%[1]d);"+"insert into symbols (symbol_name, symbol_address, symbol_type, symbol_file_ref_id, symbol_instance_id_ref) "+"select '%[2]s', '%[3]s', '%[4]s', (select file_id from files where file_name='%%[1]s' and file_instance_id_ref=%[1]d), %[1]d;"
const Insert_Xrefs_Q	string =	"insert into xrefs (caller, callee, ref_addr, source_line, xref_instance_id_ref) "+"select (Select symbol_id from symbols where symbol_address ='0x%08[1]x' and symbol_instance_id_ref=%[3]d), "+"(Select symbol_id from symbols where symbol_address ='0x%08[2]x' and symbol_instance_id_ref=%[3]d limit 1), "+"'0x%08[5]x', "+"'%[4]s', "+"%[3]d;"
const Insert_Tags2_Q	string =	"insert into tags (subsys_name, tag_file_ref_id, tag_instance_id_ref) select '%%[1]s', "+"(select file_id from files where file_name='%[1]s%%[2]s' and file_instance_id_ref=%%[3]d) as fn_id, %%[3]d "+"WHERE EXISTS ( select file_id from files where file_name='%[1]s%%[2]s' and file_instance_id_ref=%%[3]d);"

// Sql connection configuration
type Connect_token struct {
	Host   string
	Port   int
	User   string
	Pass   string
	Dbname string
}

// Connects the target db and returns the handle
func Connect_db(t *Connect_token) *sql.DB {
	fmt.Println("connect")
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", (*t).Host, (*t).Port, (*t).User, (*t).Pass, (*t).Dbname)
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		panic(err)
	}
	fmt.Println("connected")
	return db
}

// Executes insert queries
func Insert_data(db *sql.DB, query string, test bool) {

	if !test {
		_, err := db.Exec(query)
		if err != nil {
			fmt.Println("##################################################")
			fmt.Println(query)
			fmt.Println("##################################################")
			panic(err)
		}
	} else {
		fmt.Println(query)
	}
}

// Executes insert query for instance table and returns the id allocated
func Insert_datawID(db *sql.DB, query string) int {
	var res int

	_, err := db.Exec(query)
	if err != nil {
		fmt.Println("##################################################")
		fmt.Println(query)
		fmt.Println("##################################################")
		panic(err)
	}
	rows, err := db.Query("SELECT currval('instances_instance_id_seq');")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	rows.Next()
	if err := rows.Scan(&res); err != nil {
		panic(err)
	}

	return res
}
