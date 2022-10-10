package main

import "testing"


/*
type kversion struct {
        Version         int64
        Patchlevel      int64
        Sublevel        int64
        Extraversion    string
}
*/

func TestParseFiles(t *testing.T){
/*
	var test_config ={} struct{
		FileName	string
		Expected	string
		}{
		{"linux-4.9.214_autoconf.h", "4.9.214"},
		{"linux-5.13.9_autoconf.h","5.13.9"},
		{"linux-5.15_autoconf.h","5.15"},
		{"linux-5.18.4_autoconf.h","5.18.4"},
		{"linux-5.4.154_autoconf.h","5.4.154"},
		{"linux-6.0-rc2_autoconf.h","6.0-rc2"},
		}
*/
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
			t.Error("Error fetch config {}", item.FileName)
			}
		v, err:= get_version(makefile)
		if err!=nil {
			t.Error("Error parsing config" )
			}
		 if v!=item.Expected {
			t.Error("Error in parsing: got {}, expected {}", v, item.Expected)
			}
		}

}
