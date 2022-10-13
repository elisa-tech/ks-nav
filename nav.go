package main
import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"os"
	"database/sql"
	"encoding/base64"
	)
const (
	GraphOnly int		= 1
	JsonOutputPlain int	= 2
	JsonOutputB64 int	= 3
	JsonOutputGZB64 int	= 4
	)
var fmt_dot = []string {
		"",
		"\"%s\"->\"%s\" \n",
		"\\\"%s\\\"->\\\"%s\\\" \\\n",
		"\"%s\"->\"%s\" \n",
		"\"%s\"->\"%s\" \n",
		}
func generate_output(db *sql.DB, conf *configuration) (string, error){
	var	GraphOutput	string
	var 	JsonOutput	string
	var	prod =		map[string]int{}
	var	visited		[]int
	var	entry_name	string
	var	output		string

	cache := make(map[int][]Entry)
	cache2 := make(map[int]Entry)
	cache3 := make(map[string]string)

	start, err:=sym2num(db, (*conf).Symbol, (*conf).Instance)
	if err!=nil{
		fmt.Println("symbol not found")
		return "", err
		}

	GraphOutput="digraph G {\n"
	entry, err := get_entry_by_id(db, start, (*conf).Instance, cache2)
		if err!=nil {
			entry_name="Unknown";
			return "",err
			} else {
				entry_name=entry.Symbol
				}

	Navigate(db, start, entry_name, &visited, prod, (*conf).Instance, Cache{cache, cache2, cache3}, (*conf).Mode, (*conf).Excluded, 0, (*conf).MaxDepth, fmt_dot[(*conf).Jout], &output)
	GraphOutput=GraphOutput+output
	GraphOutput=GraphOutput+"}\n"

	switch (*conf).Jout {
		case GraphOnly:
			JsonOutput=GraphOutput
		case JsonOutputPlain:
			JsonOutput=fmt.Sprintf("{\"graph\": \"%s\",\"subsystems\": [%s],}", GraphOutput, "")
		case JsonOutputB64:
			b64dot:=base64.StdEncoding.EncodeToString([]byte(GraphOutput))
			JsonOutput=fmt.Sprintf("{\"graph\": \"%s\",\"subsystems\": [%s],}", b64dot, "")

		case JsonOutputGZB64:
			var b bytes.Buffer
			gz := gzip.NewWriter(&b)
			if _, err := gz.Write([]byte(GraphOutput)); err != nil {
				return "", errors.New("gzip failed")
				}
			if err := gz.Close(); err != nil {
				return "", errors.New("gzip failed")
				}
			b64dot:=base64.StdEncoding.EncodeToString(b.Bytes())
			JsonOutput=fmt.Sprintf("{\"graph\": \"%s\",\"subsystems\": [%s],}", b64dot, "")

		default:
			return "", errors.New("Unknown output mode")
	}
	return JsonOutput, nil
}




func main() {

        conf, err := args_parse(cmd_line_item_init())
        if err!=nil {
		if err.Error() != "dummy"{
			fmt.Println(err.Error())
			}
                print_help(cmd_line_item_init());
                os.Exit(-1)
                }

	t:=Connect_token{ conf.DBURL, conf.DBPort,  conf.DBUser, conf.DBPassword, conf.DBTargetDB}
	db:=Connect_db(&t)

	output, err := generate_output(db, &conf)
	if err!=nil{
		fmt.Println("internal error", err)
		os.Exit(-3)
		}
	fmt.Println(output)

}

