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
	"strings"
	"regexp"
	"strconv"
)

type kversion struct {
	Version		int64
	Patchlevel	int64
	Sublevel	int64
	Extraversion	string
}

func parse_config(kconfig []string) (map[string]string){

	res := make(map[string]string)
	for _, line := range kconfig {
			if strings.Contains(line, "#define CONFIG_") {
				tmp := strings.ReplaceAll(line, "#define CONFIG_", "")
				items:=strings.Split(tmp, " ")
				res[items[0]]=items[1]
				}
		}
	return res
}
func get_version(makefile []string) (kversion, error){
	var state	int=0
	var v		kversion

	for _, line := range makefile {
			if match, _ := regexp.MatchString("VERSION[ \t]*=[ \t]*[0-9]+", line); match  && state==0 {
				re:=regexp.MustCompile(`VERSION[ \t]*=[ \t]*([0-9]+)`)
				tmp, err := strconv.ParseInt(re.ReplaceAllString(line, "$1"), 10, 64)
				if err!=nil{
					panic(err)
					}
				v.Version=tmp
				state=1
				}
			if match, _ := regexp.MatchString("PATCHLEVEL[ \t]*=[ \t]*[0-9]+", line); match && state==1 {
				re:=regexp.MustCompile(`PATCHLEVEL[ \t]*=[ \t]*([0-9]+)`)
				tmp, err := strconv.ParseInt(re.ReplaceAllString(line, "$1"), 10, 64)
				if err!=nil{
					panic(err)
					}
				v.Patchlevel=tmp
				state=2
				}
			if match, _ := regexp.MatchString("SUBLEVEL[ \t]*=[ \t]*[0-9]+", line); match  && state==2 {
				re:=regexp.MustCompile(`SUBLEVEL[ \t]*=[ \t]*([0-9]+)`)
				tmp, err := strconv.ParseInt(re.ReplaceAllString(line, "$1"), 10, 64)
				if err!=nil{
					panic(err)
					}
				v.Sublevel=tmp
				state=3
				}
			if match, _ := regexp.MatchString("EXTRAVERSION[ \t]*=[ \t]*.*", line); match  && state==3 {
				re:=regexp.MustCompile(`EXTRAVERSION[ \t]*=[ \t]*(.*)`)
				v.Extraversion=re.ReplaceAllString(line, "$1")
				state=4
				break
				}
			}
	if state==4 {
		return v, nil
		}
	return v, fmt.Errorf("can't parse makefile (%d)", state)

}
