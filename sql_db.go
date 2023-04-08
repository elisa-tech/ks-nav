package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Sql connection configuration.
type connectToken struct {
	DBDriver string
	DBDSN    string
}

type Cache struct {
	successors map[int][]entry
	entries    map[int]entry
	subSys     map[string]string
}

type SqlDB struct {
	db *sql.DB
	cache Cache
}
/*
func (d SqlDB) check_operative() bool {
	fmt.Printf("check_operative %p\n", d.db)
	return d.db==nil
}
*/

// Connects the target db and returns the handle.
func (d *SqlDB) init(arg interface{}) (err error){
	t, ok := arg.(*connectToken)
	if !ok {
		var ok1 bool
		d.db, ok1 = arg.(*sql.DB)
		if !ok1 {
			return errors.New("invalid type")
		}
	}
	if ok {
		d.db, err = sql.Open(t.DBDriver, t.DBDSN)
	}
	if err==nil {
		d.cache.successors = make(map[int][]entry)
		d.cache.entries = make(map[int]entry)
		d.cache.subSys = make(map[string]string)
	}
	return err
}


func (d *SqlDB) GetExploredSubsystemByName(subs string) (string) {
	return d.cache.subSys[subs]
}

// Returns function details from a given id.
func (d *SqlDB) getEntryById(symbolId int, instance int) (entry, error) {
	var e entry
	var s sql.NullString

	if e, ok := d.cache.entries[symbolId]; ok {
		return e, nil
	}

	query := "select symbol_id, symbol_name, subsys_name, file_name from " +
		"(select * from symbols, files where symbols.symbol_file_ref_id=files.file_id and symbols.symbol_instance_id_ref=%[2]d) as dummy " +
		"left outer join tags on dummy.symbol_file_ref_id=tags.tag_file_ref_id where symbol_id=%[1]d and symbol_instance_id_ref=%[2]d"
	query = fmt.Sprintf(query, symbolId, instance)
	rows, err := d.db.Query(query)
	if err != nil {
		return entry{}, err
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil {
			err = closeErr
		}
	}()

	for rows.Next() {
		if err := rows.Scan(&e.symId, &e.symbol, &s, &e.fn); err != nil {
			fmt.Println("getEntryById: error while scan query rows")
			fmt.Println(err)
			return e, err
		}
		if s.Valid {
			e.subsys = append(e.subsys, s.String)
		}
	}
	if err = rows.Err(); err != nil {
		fmt.Println("getEntryById: error in access query rows")
		return e, err
	}
	d.cache.entries[symbolId] = e
	return e, nil
}

// Returns the list of successors (called function) for a given function.
func (d *SqlDB) getSuccessorsById(symbolId int, instance int) ([]entry, error) {
	var e edge
	var res []entry

	if res, ok := d.cache.successors[symbolId]; ok {
		return res, nil
	}

	query := "select caller, callee, source_line, ref_addr from xrefs where caller = %[1]d and xref_instance_id_ref = %[2]d"
	query = fmt.Sprintf(query, symbolId, instance)
	rows, err := d.db.Query(query)
	if err != nil {
		panic(err)
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil {
			err = closeErr
		}
	}()

	for rows.Next() {
		if err := rows.Scan(&e.caller, &e.callee, &e.sourceRef, &e.addressRef); err != nil {
			fmt.Println("get_successors_by_id: error while scan query rows", err)
			return nil, err
		}
		successor, _ := d.getEntryById(e.callee, instance)
		successor.sourceRef = e.sourceRef
		successor.addressRef = e.addressRef
		res = append(res, successor)
	}
	if err = rows.Err(); err != nil {
		fmt.Println("get_successors_by_id: error in access query rows")
		return nil, err
	}
	d.cache.successors[symbolId] = res
	return res, nil
}

// Given a function returns the lager subsystem it belongs.
func (d *SqlDB) getSubsysFromSymbolName(symbol string, instance int) (string, error) {
	var ty, sub string

	if res, ok := d.cache.subSys[symbol]; ok {
		return res, nil
	}
	query := "select (select symbol_type from symbols where symbol_name='%[1]s' and symbol_instance_id_ref=%[2]d) as type, subsys_name from " +
		"(select count(*) as cnt, subsys_name from tags where subsys_name in (select subsys_name from symbols, " +
		"tags where symbols.symbol_file_ref_id=tags.tag_file_ref_id and symbols.symbol_name='%[1]s' and symbols.symbol_instance_id_ref=%[2]d) " +
		"group by subsys_name order by cnt desc) as tbl;"

	query = fmt.Sprintf(query, symbol, instance)
	rows, err := d.db.Query(query)
	if err != nil {
		panic(err)
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil {
			err = closeErr
		}
	}()

	for rows.Next() {
		if err := rows.Scan(&ty, &sub); err != nil {
			fmt.Println("get_subsys_from_symbol_name: error while scan query rows")
			return "", err
		}
	}

	if err = rows.Err(); err != nil {
		fmt.Println("getSubsysFromSymbolName: error in access query rows")
		return "", err
	}

	if sub == "" {
		return "", nil
	}

	if ty == "indirect" {
		sub = ty
	}
	d.cache.subSys[symbol] = sub
	return sub, nil
}

// Returns the id of a given function name.
func (d *SqlDB) sym2num(symb string, instance int) (int, error) {
	var res = -1
	var cnt = 0
	query := "select symbol_id from symbols where symbols.symbol_name='%[1]s' and symbols.symbol_instance_id_ref=%[2]d"
	query = fmt.Sprintf(query, symb, instance)
	rows, err := d.db.Query(query)
	if err != nil {
		panic(err)
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil {
			err = closeErr
		}
	}()

	for rows.Next() {
		cnt++
		if err := rows.Scan(&res); err != nil {
			fmt.Println("sym2num: error while scan query rows")
			fmt.Println(err)
			return res, err
		}
	}

	if err = rows.Err(); err != nil {
		fmt.Println("sym2num: error in access query rows")
		return -1, err
	}

	if cnt != 1 {
		return res, errors.New("duplicate ID in the DB")
	}
	return res, nil
}

// Returns the subsystem list associated with a given function name.
func (d *SqlDB) symbSubsys(symblist []int, instance int) (string, error) { 
	var out string
	var res string

	for _, symbid := range symblist {
		// Resolve symb.
		symb, err := d.getEntryById(symbid, instance)
		if err != nil {
			return "", fmt.Errorf("symbSubsys::getEntryById error: %s", err)
		}
		out += fmt.Sprintf("{\"FuncName\":\"%s\", \"subsystems\":[", symb.symbol)
		query := fmt.Sprintf("select subsys_name from tags where tag_file_ref_id= (select symbol_file_ref_id from symbols where symbol_id=%d);", symbid)
		rows, err := d.db.Query(query)
		if err != nil {
			return "", errors.New("symbSubsys: query failed")
		}

		defer func() {
			closeErr := rows.Close()
			if err == nil {
				err = closeErr
			}
		}()

		for rows.Next() {
			if err := rows.Scan(&res); err != nil {
				return "", errors.New("symbSubsys: error while scan query rows")
			}
			out += fmt.Sprintf("\"%s\",", res)
		}
		out = strings.TrimSuffix(out, ",") + "]},"
	}
	out = strings.TrimSuffix(out, ",")
	return out, nil
}
