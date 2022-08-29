package main
import (
	"strconv"
	"fmt"
	"os"
	"errors"
	"encoding/json"
	"io/ioutil"
        )
var conf_fn = "conf.json"
const app_name	string="nav"
const app_descr	string="kernel symbol navigator"

type Arg_func func(*configuration, []string) (error)

type cmd_line_items struct {
	id		int
	Switch		string
	Help_srt	string
	Has_arg		bool
	Needed		bool
	Func		Arg_func
}

type configuration struct {
	DBURL		string
	DBPort		int
	DBUser		string
	DBPassword	string
	DBTargetDB	string
	Symbol		string
	Instance	int
	Mode		int
	Excluded	[]string
}
var	Default_config  configuration = configuration{
	DBURL:		"dbs.hqhome163.com",
	DBPort:		5432,
	DBUser:		"alessandro",
	DBPassword:	"<password>",
	DBTargetDB:	"kernel_bin",
	Symbol:		"",
	Instance:	0,
	Mode:		PRINT_SUBSYS,
	Excluded:	[]string{"rcu_.*"},
	}

func push_cmd_line_item(Switch string, Help_str string, Has_arg bool, Needed bool, Func Arg_func, cmd_line *[]cmd_line_items){
	*cmd_line = append(*cmd_line, cmd_line_items{id: len(*cmd_line)+1, Switch: Switch, Help_srt: Help_str, Has_arg: Has_arg, Needed: Needed, Func: Func})
}

func cmd_line_item_init() ([]cmd_line_items){
	var res	[]cmd_line_items

	push_cmd_line_item("-s", "Specifies symbol",				true,  true,	func_symbol,	&res)
	push_cmd_line_item("-i", "Specifies instance",				true,  true,	func_instance,	&res)
	push_cmd_line_item("-f", "Specifies config file",			true,  false,	func_jconf,	&res)
	push_cmd_line_item("-u", "Forces use specified database userid",	true,  false,	func_DBUser,	&res)
	push_cmd_line_item("-p", "Forecs use specified password",		true,  false,	func_DBPass,	&res)
	push_cmd_line_item("-d", "Forecs use specified DBhost",			true,  false,	func_DBHost,	&res)
	push_cmd_line_item("-p", "Forecs use specified DBPort",			true,  false,	func_DBPort,	&res)
	push_cmd_line_item("-m", "Sets display mode 2=subsystems,1=all",	true,  false,	func_Mode,	&res)
	push_cmd_line_item("-h", "This Help",					false, false,	func_help,	&res)

	return res
}
func func_help          (conf *configuration,fn []string)               (error){
	return errors.New("Dummy")
}
func func_jconf		(conf *configuration,fn []string)		(error){
	jsonFile, err := os.Open(fn[0])
	if err != nil {
                return err
                }
        byteValue, _ := ioutil.ReadAll(jsonFile)
        jsonFile.Close()
	err=json.Unmarshal(byteValue, conf)
	if err != nil {
		return err
		}
	return nil
}

func func_symbol	(conf *configuration, fn []string)	(error){
	(*conf).Symbol=fn[0]
	return nil
}

func func_DBUser	(conf *configuration, user []string)	(error){
	(*conf).DBUser=user[0]
	return nil
}

func func_DBPass	(conf *configuration, pass []string)	(error){
	(*conf).DBPassword=pass[0]
	return nil
}

func func_DBHost	(conf *configuration, host []string)	(error){
	(*conf).DBURL=host[0]
	return nil
}

func func_DBPort	(conf *configuration, port []string)	(error){
	s, err := strconv.Atoi(port[0])
	if err!=nil {
		return err
		}
	(*conf).DBPort=s
	return nil
}

func func_instance        (conf *configuration, instance []string)    (error){
        s, err := strconv.Atoi(instance[0])
        if err!=nil {
                return err
                }
        (*conf).Instance=s
        return nil
}

func func_Mode        (conf *configuration, mode []string)    (error){
        s, err := strconv.Atoi(mode[0])
        if err!=nil {
                return err
                }
	if s<1 || s>2 {
		return errors.New("unsupported mode")
		}
        (*conf).Mode=s
        return nil
}

func print_help(lines []cmd_line_items){

	fmt.Println(app_name)
	fmt.Println(app_descr)
	for _,item := range lines{
		fmt.Printf(
			"\t%s\t%s\t%s\n",
			item.Switch,
			func (a bool)(string){
				if a {
					return "<v>"
					}
				return ""
			}(item.Has_arg),
			item.Help_srt,
			)
		}
}

func args_parse(lines []cmd_line_items)(configuration, error){
	var	skip		bool=false;
	var	conf		configuration=Default_config
	var 	f		Arg_func

	for _, os_arg := range os.Args[1:] {
//		fmt.Printf("osarg=%s\n", os_arg)
		if !skip {
//			fmt.Printf("check if I have it\n")
			for _, arg := range lines{
//				fmt.Printf("consider configured arg=%s\n", arg.Switch)
				if arg.Switch==os_arg {
//					fmt.Printf("have it\n")
					if arg.Has_arg{
//						fmt.Printf("it needs more arguments\n")
						f=arg.Func
						skip=true
						break
						}
					err := arg.Func(&conf, []string{})
					if err != nil {
//						fmt.Println("-----------------------------_")
						return Default_config, err
						}
					}
				}
			continue
			}
		if skip{
//			fmt.Printf("Fetch extra arg (%s)and use it\n", os_arg)
			err := f(&conf,[]string{os_arg})
			if err != nil {
				return Default_config, err
				}
			skip=false
			}

		}
	return	conf, nil
}

