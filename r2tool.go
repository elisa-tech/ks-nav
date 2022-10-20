	/*
	 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	 *
	 *   Name: nav - Kernel source code analysis tool
	 *   Description: Extract call trees for kernel API
	 *
	 *   Author: Alessandro Carminati <acarmina@redhat.com>
	 *   Author: Maurizio Papini <mpapini@redhat.com>
	 *
	 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	 *
	 *   Copyright (c) 2008-2010 Red Hat, Inc. All rights reserved.
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
	"errors"
	"log"
	"encoding/json"
	"sort"
	"strings"
	"strconv"
	r2 "github.com/radareorg/r2pipe-go"
)

type sysc struct {
	Addr		uint64
	Name		string
}

type res struct{
	Syscall		sysc
	Path		[]uint64
}

type reloc_data struct {
	Name		string		`json: "name"`
	Demname		string		`json: "demname"`
	Type		string		`json: "type"`
	Vaddr		uint64		`json: "vaddr"`
	Paddr		uint32		`json: "paddr"`
	Sym_va		uint64		`json: "sym_va"`
	is_ifunc	bool		`json: "is_ifunc"`
}

type func_data struct {
	Offset		uint64		`json:"offset"`
	Name		string		`json: "name"`
	Size		uint64		`json: "size"`
	Is_pure		string		`json: "is-pure"`
	Realsz		uint64		`json: "realsz"`
	Noreturn	bool		`json: "noreturn"`
	Stackframe	uint16		`json: "stackframe"`
	Calltype	string		`json: "calltype"`
	Cost		uint16		`json: "cost"`
	Cc		uint16		`json: "cc"`
	Bits		uint16		`json: "bits"`
	Type		string		`json: "type"`
	Nbbs		uint16		`json: "nbbs"`
	Is_lineal	bool		`json: "is-lineal"`
	Ninstrs		uint16		`json: "ninstrs"`
	Edges		uint16		`json: "edges"`
	Ebbs		uint16		`json: "ebbs"`
	Signature	string		`json: "signature"`
	Minbound	uint64		`json: "minbound"`
	Maxbound	uint64		`json: "maxbound"`
	Callrefs	[]ref_		`json: "callrefs"`
	Datarefs	[]uint64	`json: "datarefs"`
	Codexrefs	[]ref_		`json: "codexrefs"`
	Dataxrefs	[]uint64	`json: "dataxrefs"`
	Indegree	uint16		`json: "indegree"`
	Outdegree	uint16		`json: "outdegree"`
	Nlocals		uint16		`json: "nlocals"`
	Nargs		uint16		`json: "nargs"`
	Bpvars		[]stack_var_	`json: "bpvars"`
	Spvars		[]stack_var_	`json: "spvars"`
	Regvars		[]reg_var_	`json: "regvars"`
	Difftype	string		`json: "difftype"`
	Indirect	bool
}
type ref_ struct{
	Addr		uint64		`json: "addr"`
	Type		string		`json: "type"`
	At		uint64		`json: "at"`
}
type stack_var_ struct{
	Name		string		`json: "name"`
	Kind		string		`json: "kind"`
	Type		string		`json: "type"`
	Ref		vars_ref	`json: "ref"`
}
type vars_ref struct{
	Base		string		`json: "base"`
	Offset		int32		`json: "offset"`
}
type reg_var_ struct{
	Name		string		`json: "name"`
	Kind		string		`json: "kind"`
	Type		string		`json: "type"`
	Ref		string		`json: "ref"`
}

type xref struct{
	Type		string		`json: "type"`
	From		uint64		`json: "from"`
	To		uint64		`json: "to"`
}
type xref_cache struct{
	Addr		uint64
	Xr		[]uint64
}

type fref struct{
	Addr		uint64
	Name		string
}
type results struct{
	Addr		uint64
	Name		string
	Path		[]fref
}
type rad_bloc struct {
	Jump		uint64		`json: "jump"`
	Fail		uint64		`json: "fail"`
	Opaddr		uint64		`json: "opaddr"`
	Addr		uint64		`json: "addr"`
	Size		uint64		`json: "size"`
	Inputs		uint8		`json: "inputs"`
	Outputs		uint8		`json: "outputs"`
	ninstr		uint16		`json: "ninstr"`
	traced		bool		`json: "traced"`
}
type bloc struct {
	Start		uint64
	End		uint64
}

type bin_info struct{
	Arch		string		`json: "arch"`
	Bits		int		`json: "bits"`
	Compiler	string		`json: "compiler"`
	Endian		string		`json: "endian"`
	Machine		string		`json: "machine"`
}
type core_info struct{
	Type		string		`json: "type"`
	Format		string		`json: "format"`
}
type file_info struct{
	Core	core_info
	Bin	bin_info
}
type symb_data struct{
	Name		string		`json: "name"`
	Realname	string		`json: "realname"`
	Size		int		`json: "size"`
	Offset		uint64		`json: "offset"`
}
func get_function_by_addr(addr uint64, all_funcs []func_data)(*func_data){

	for i, f := range all_funcs{
		if addr == f.Offset {
			return &(all_funcs[i])
			}
		}
	return nil
}

func get_all_relocdata(r2p *r2.Pipe)([]reloc_data){

	var relocs   []reloc_data

	buf, err := r2p.Cmd("irj")
	if err != nil {
		panic(err)
		}
	error := json.Unmarshal( []byte(buf), &relocs)
	if(error != nil){
		fmt.Printf("Error while parsing data: %s", error)
		}
	return relocs
}

func removeSDup(intSlice []string) []string {

	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range intSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
			}
		}
	return list
}

func get_f_relocs(sym string, all_relocs []reloc_data, all_funcs []func_data) ([]string, error){
	var fun func_data
	var  res []string
	for _,f := range all_funcs {
		if f.Name==sym {
			fun=f
			break
			}
		}
	if fun.Name == "" {
		return  nil, errors.New("symbol not found")
		}
	for _,r := range all_relocs{
		if (r.Sym_va ==0) && (r.Vaddr>=fun.Offset) && (r.Vaddr<=fun.Offset+fun.Size) {
			res=append(res, r.Name)
			}
		}
	return removeSDup(res), nil
}

func Move(r2p *r2.Pipe,current uint64){
	_, err := r2p.Cmd("s "+ strconv.FormatUint(current,10))
	if err != nil {
		panic(err)
		}
}

func Getxrefs(r2p *r2.Pipe, current uint64, indcall []uint64, funcs []func_data, cache *[]xref_cache) ([]uint64){
	var xrefs		[]xref
	var res			[]uint64;

	for _, item := range *cache  {
		if item.Addr==current {
			return item.Xr
				}
		}
	buf, err := r2p.Cmd("afxj")
	if err != nil {
		panic(err)
		}
	error := json.Unmarshal( []byte(buf), &xrefs)
	if(error != nil){
		fmt.Printf("Error while parsing data: %s", error)
		}
	for _, item := range xrefs  {
		res=append(res,item.To)
		}
	if func_has_indirects(r2p, indcall, current, funcs){
		res=append(res,0)	//zero is null and it is used to indicate indirect calls.
		}
	*cache=append(*cache,xref_cache{current,res})
	return  res
}

func Symb2Addr_r(s string, r2p *r2.Pipe) (uint64){
	var f  []func_data
	buf, err := r2p.Cmd("afij "+ s)
	if err != nil {
		panic(err)
		}
	error := json.Unmarshal( []byte(buf), &f)
	if(error != nil){
		fmt.Printf("Error while parsing data: %s", error)
		}
	if len(f)>0 {
		return f[0].Offset
		}
	return 0
}

func removeDuplicate(intSlice []uint64) []uint64 {

	allKeys := make(map[uint64]bool)
	list := []uint64{}
	for _, item := range intSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
			}
		}
	return list
}

func remove_non_func(list []uint64, functions []func_data) []uint64 {

	res := []uint64{}
	for _, item := range list {
		if is_func(item, functions) || item==0 {
			res = append(res, item)
			}
		}
	return res
}

func init_fw(r2p *r2.Pipe){
	l := log.New(os.Stderr, "", 0)

	l.Println("Initializing Radare framework")
	_, err := r2p.Cmd("e anal.nopskip=false")
	if err != nil {
		panic(err)
		}
	_, err = r2p.Cmd("aa")
	if err != nil {
		panic(err)
		}
	l.Println("analisys")



}

func is_func(addr uint64, list []func_data) (bool){
	i := sort.Search(len(list), func(i int) bool { return list[i].Offset >= addr })
	if i < len(list) && list[i].Offset == addr && strings.Contains(list[i].Name, "sym."){
		return true;
		}
	return false
}

func get_all_funcdata(r2p *r2.Pipe)([]func_data){

	var functions	[]func_data
	var symbols	[]symb_data
	var info	file_info

	buf, err := r2p.Cmd("aflj")
	if err != nil {
		panic(err)
		}
	error := json.Unmarshal( []byte(buf), &functions)
	if(error != nil){
		fmt.Printf("Error while parsing data: %s", error)
		}
	buf, err = r2p.Cmd("ij")
	if err != nil {
		panic(err)
		}
	error = json.Unmarshal( []byte(buf), &info)
	if(error != nil){
		fmt.Printf("Error while parsing data: %s", error)
		}
	if info.Bin.Arch=="x86" {
		buf, err = r2p.Cmd("fs symbols")
		if err != nil {
			panic(err)
			}
		buf, err = r2p.Cmd("fj")
		if err != nil {
			panic(err)
			}
		error = json.Unmarshal( []byte(buf), &symbols)
		if(error != nil){
			fmt.Printf("Error while parsing data: %s", error)
			}
		for _,s :=range symbols{
			if strings.Contains(s.Name, "loc.__x86_indirect_thunk_") {
				functions=append(functions,func_data{Offset: s.Offset, Name: s.Name, Indirect: true})
				}
			}
		}
	sort.SliceStable(functions, func(i, j int) bool {return functions[i].Offset < functions[j].Offset})
	return functions
}

func Addr2Sym(addr uint64, list []func_data) (string){
	i := sort.Search(len(list), func(i int) bool { return list[i].Offset >= addr })
	if i < len(list) && list[i].Offset == addr {
		return list[i].Name;
		}
	return "Unknown"
}

func get_indirect_calls(r2p *r2.Pipe, funcs []func_data) ([]uint64){
	var smap	[]uint64

	buf, err := r2p.Cmd("/at rcall")
	if err != nil {
		panic(err)
		}
	temp := strings.Split(buf,"\n")
	for _, line := range temp {
		temp2 := strings.Split(line," ")
		num, err := strconv.ParseUint(strings.Replace(temp2[0], "0x", "", -1) , 16, 64)
		if err != nil {
			panic(err)
			}
			smap = append(smap, uint64(num))
		}
	buf, err = r2p.Cmd("/at ucall")
	if err != nil {
		panic(err)
		}
	if len(buf)>10 {
		temp = strings.Split(buf,"\n")
		for _, line := range temp {
			temp2 := strings.Split(line," ")
			num, err := strconv.ParseUint(strings.Replace(temp2[0], "0x", "", -1) , 16, 64)
			if err != nil {
				panic(err)
				}
			smap = append(smap, uint64(num))
			}
		}
	sort.SliceStable(smap, func(i, j int) bool {return smap[i] < smap[j]})
	return smap
}

func get_func_space(r2p *r2.Pipe, addr uint64, funcs []func_data)([]bloc){
	var blocs	[]bloc
	var rad_blocs	[]rad_bloc

	for _, f := range funcs {
		if f.Offset == addr {
			if f.Size==f.Realsz {
				blocs=append(blocs,bloc{f.Offset, f.Offset+f.Size})
				} else {
				buf, err := r2p.Cmd("afbj")
				if err != nil {
					panic(err)
					 }
				error := json.Unmarshal( []byte(buf), &rad_blocs)
				if(error != nil){
					fmt.Printf("Error while parsing data: %s", error)
					}
				for _,b := range rad_blocs {
					blocs=append(blocs, bloc{b.Addr,b.Addr+b.Size})
					}
				}
		break
		}
	}
	return blocs
}

func is_in_func(r2p *r2.Pipe, addr uint64, funcs []func_data) (uint64){

	for _, f := range funcs {
		blocs:=get_func_space(r2p, f.Offset, funcs)
		for _, b := range blocs {
			 if addr >= b.Start && addr <= b.End {
				return f.Offset
				}
			}
		}
	return 0

}

func func_has_indirects(r2p *r2.Pipe, indcall []uint64, faddr uint64, funcs []func_data) (bool){

	blocs:=get_func_space(r2p, faddr, funcs)
	for _, ic:= range indcall{
		for _, b := range blocs {
			 if ic >= b.Start && ic <= b.End {
				return true
				}
			}
		}
	return false
}
