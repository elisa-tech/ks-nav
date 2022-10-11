package main

import (
	"testing"
	"sort"
	"crypto/sha1"
	"fmt"
	"encoding/hex"
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	)

func TestParseFilesMakefile(t *testing.T){
	var test_makefile =[]struct{
		FileName	string
		Expected	kversion
		}{
		{"t_files/linux-4.9.214_Makefile",	kversion{4,9,214,""}	},
		{"t_files/linux-5.13.9_Makefile",	kversion{5,13,9,""}	},
		{"t_files/linux-5.15_Makefile",		kversion{5,15,0,""}	},
		{"t_files/linux-5.18.4_Makefile",	kversion{5,18,4,""}	},
		{"t_files/linux-5.4.154_Makefile",	kversion{5,4,154,""}	},
		{"t_files/linux-6.0-rc2_Makefile",	kversion{6,0,0,"-rc2"}	},
		}

	for _, item := range test_makefile {
		makefile, err := get_FromFile(item.FileName);
		if err!=nil {
			t.Error("Error fetch makefile {}", item.FileName)
			}
		v, err:= get_version(makefile)
		if err!=nil {
			t.Error("Error parsing makefile")
			}
		 if v!=item.Expected {
			t.Error("Error in validating the result: got {}, expected {}", v, item.Expected)
			}
		}

}
/**/
func TestParseFilesConfig(t *testing.T){
	var test_config =[]struct{
		FileName	string
		Expected	string
		}{
		{"t_files/linux-4.9.214_autoconf.h",	"7e3619ddf81d683c15e5cb55c57dd16386b359aa"},
		{"t_files/linux-5.13.9_autoconf.h",	"99fd41f9da13c43f880ec71500e56b719db4308f"},
		{"t_files/linux-5.15_autoconf.h",	"eaa565eaedbbd1b9aaf7bbceb51804cec3dcca53"},
		{"t_files/linux-5.18.4_autoconf.h",	"fec6afca6f92e093433727c3c6d1fd07ffbe5f12"},
		{"t_files/linux-5.4.154_autoconf.h",	"d1471ae2dbf261ae65089db3b012676834fceae8"},
		{"t_files/linux-6.0-rc2_autoconf.h",	"2a6d6426a81c2f84771c00a286f0a592f4cc6a24"},
		}

	for _, item := range test_config  {
		config, err := get_FromFile(item.FileName)
		if err!=nil {
                        t.Error("Error fetch config {}", item.FileName)
                        }
		kconfig:=parse_config(config)
		tconf:=""

		keys := make([]string, 0, len(tconf))
		for k := range kconfig{
			keys = append(keys, k)
			}
		sort.Strings(keys)
		for _, k := range keys {
				tconf=tconf+fmt.Sprintf("CONFIG_%s=%s\n", k, kconfig[k])
			}
		hasher := sha1.New()
		hasher.Write([]byte(tconf))
		sha := hex.EncodeToString(hasher.Sum(nil))
		if sha!=item.Expected {
			t.Error("Error in validating the result: got {}, expected {}", sha, item.Expected)
			}
		}
}

func Untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
		}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
			} else if err != nil {
				return err
				}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
				}
			continue
			}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
			}
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
			}
		 file.Close()
		}
	return nil
}

func cp(source string, destination string) error{

	bytesRead, err := ioutil.ReadFile(source)
	if err != nil {
		return err
		}
	err = ioutil.WriteFile(destination, bytesRead, 0644)
	if err != nil {
		return err
		}
	return nil
}


func TestMaintainer(t *testing.T){
	var FakeLinuxTreeTest string =	"t_files/linux-fake.tar"
	var Fakedir string =		"/tmp/linux-fake"
	var testData = []struct {
		filename	string
		subs		int
		files		int
		}{
			{"t_files/linux-4.9.214_MAINTAINERS",	1627,113029},
			{"t_files/linux-5.10.57_MAINTAINERS",	2251,135588},
			{"t_files/linux-5.13.9_MAINTAINERS",	2336,136637},
			{"t_files/linux-5.8.2_MAINTAINERS",	2208,133981},
			{"t_files/linux-6.0_MAINTAINERS",	2585,142382},
		}


	defer os.RemoveAll(Fakedir)

	err := Untar(FakeLinuxTreeTest, Fakedir)
	if err!=nil {
		t.Error("Error cant initialize fake linux directory", err)
		}

	current, _ := os.Getwd()
	err = os.Chdir(Fakedir)
	if err != nil {
		t.Error("Error cant initialize fake linux directory", err)
		}

	for _,f := range testData {
		err=cp(current+"/"+f.filename, "MAINTAINERS")
		if err!=nil {
			t.Error("Error cant use maintainer file", f.filename)
			}
		s, err := get_FromFile("MAINTAINERS")
		if err!= nil {
			t.Error("Error cant read maintainers", err)
			}
		ss:=s[seek2data(s):]
		items:=parse_maintainers(ss)
		queries:=generate_queries(items, "Subsystem: %[1]s' FileName='%[2]s",0)
		if (f.subs!=len(items)) && (f.files!=len(queries)){
			t.Error("Error validating number of files and subsystems in ", f.filename)
			}
		}

}
