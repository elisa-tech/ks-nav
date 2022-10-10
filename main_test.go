package main

import (
	"testing"
	"sort"
	"crypto/sha1"
	"fmt"
	"encoding/hex"
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

