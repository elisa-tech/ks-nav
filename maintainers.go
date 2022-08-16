package main

import (
	"net/http"
	"io/ioutil"
	"strings"
	"bufio"
	"os"
	"path/filepath"
	"fmt"
)

type m_item struct{
	subsystem_name	string
	wildcards	[]string
}

func get_FromHttp(url string) ([]string, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
		}
	return strings.Split(string(buf), "\n") , nil
}

func get_FromFile(path string) ([]string, error) {
	var lines []string

	file, err := os.Open(path)
	if err != nil {
		return nil, err
		}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
 		lines = append(lines, scanner.Text())
		}
	return lines, scanner.Err()
}

func seek2data(s []string) int{
	var state int = 0
	var i	int
	var line string

	res:=0
	for i,line = range s{
		searchpattern:=""
		if len(line)>=2 {
			searchpattern=line[0:2]
			}
//		fmt.Printf("%d [%s] - %s\n", state, searchpattern, line)
		if len(line)>=2 && searchpattern ==".." {
			state=1
			}
		if state == 1 && line=="" {
			state = 2
			break
			}
		}
	if state==2 {
		res=i+1
		}
	return res
}


func parse_maintainers(lines []string) []m_item{
	var res []m_item
	var it m_item
	var state int = 0
//	lines := strings.Split(s, "\n")
//	fmt.Println(lines)

	for _, line := range lines {
//		fmt.Printf("%d - %s\n", state, line)
		if state == 0 {
			it.subsystem_name = line
			state = 1
			continue
		}
		if state == 1 && len(line) > 2 && line[0:2] == "F:" {
			tmp := strings.Split(line, ":")
			it.wildcards = append(it.wildcards,  strings.TrimLeftFunc(tmp[1], func(c rune) bool {
//							fmt.Println(c)
							if c == ' ' || c == '\t' {
								return true
								} else {
									return false
									}
							}))
			continue
		}
		if len(line) < 2 {
			if it.subsystem_name!="THE REST"{
				res = append(res, it)
				it.subsystem_name=""
				it.wildcards = nil
				}
			state = 0
		}
	}
//	fmt.Println(res)
	return res
}

func isdir(f string) bool {
	file, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	if fileInfo.IsDir() {
		return true
	}
	return false
}

func expand_file(f string) []string {
	var res []string
	if isdir(f) {
		filesInDir, err := ioutil.ReadDir(f)
		if err != nil {
			panic(err)
		}
		for _, file := range filesInDir {
			res = append(res, f+"/"+file.Name())
		}
		return res
	}
	return []string{f}
}

func navigate(root string) []string {
	var res []string
	if isdir(root) {
		for _, current := range expand_file(root) {
			res = append(res, navigate(current)...)
		}
	} else {
		res = []string{root}
	}
	return res
}



func generate_queries(items []m_item, template_query string, id int) []string{
	var res []string

	for _, item := range items{
		for _, wildcard_item := range item.wildcards {
			files, err := filepath.Glob(wildcard_item)
			if err != nil {
				panic(err)
				}
			for _, f := range files{
				if isdir(f) {
					for _,x:= range navigate(f){
						res=append(res,fmt.Sprintf(template_query, strings.ReplaceAll(item.subsystem_name, "'", "`"), filepath.Clean(x), id))
						}
					} else {
						res=append(res,fmt.Sprintf(template_query, strings.ReplaceAll(item.subsystem_name, "'", "`"), filepath.Clean(f), id))
						}
				}
			}
		}

	return res
}
