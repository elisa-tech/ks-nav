# Nav - kernel source code navigator

Nav is a tool that uses a pre-constituted database to emit call trees graphs that can be used stand alone or feed into a graph display system to help engineers do static analysis.

## Motivation
Although other similar tool do exist, the motivation for this tool is to be developed, is to solve a specific need: do a kernel source code analysis aimed at feature level and  shows it as functions call tree or subsystems call tree. 

## Build
Nav is implemented in Golang. Golang applications are usually easy to build. In addition to this, it has very few dependencies other than the standard Golang packages.
A `Makefile` is provided to ease the build process. 
Just typing 
```
$ make
```
the native build is made.
In addition to the default target,  `amd64`, `arm64`, and `upx`, targets exists.
|target |function                                                                   |
|-------|---------------------------------------------------------------------------|
|amd64  |forces the build to amd64 aka x86_64 regardless the underlying architecture|
|arm64  |forces the build to arm64 aka aarc64 regardless the underlying architecture|
|upx    |triggers compress previously generated executable using UPX                |
As example, this builds aarch64 upx compressed executable:
```
$ make arm64
$ make upx
```
## Usage example
As the nav compiled executable is available, it is essential to provide the configuration to query the backend database. The easiest way to provide the configuration to nav is to specify a configuration file.
Although the nav tool has an internal default for all the configuration parameters, that are used if not otherwise specified, this default can be overridden by both configuration file or command line switches.
The configuration file is a plain json object, and it can be passed by using the command line switch `-f`.
The order on which the  configuration is evaluated is as depicted here:
```
+-------------------+    +--------------+   +----------+
|Nav builtin default|--->|conf json file|-->|CLI switch|
+-------------------+    +--------------+   +----------+
```
So the following example the built-in default is overridden with the conf.json and the arguments in the command line in the end override the final configuration.

```
$ ./nav -f conf.json -s kernel_init
```
The following is the command line switches list from the nav help.
```
$ ./nav -h
Nav - kernel symbol navigator

Usage:
  nav [FLAGS]

Flags:
  -f, --config config              path to config file
  -s, --symbol symbol              name of the symbol to start the navigation from
  -j, --output-type type           type of output: graphOnly, jsonOutputPlain, jsonOutputB64 or jsonOutputGZB64 (default "graphOnly")
  -x, --max-depth number           max number of levels in call flow exploration (0=No limit)
  -m, --mode mode                  mode of plotting: 1=Symbols, 2=Subsystems, 3=Subsystems with labels, 4=Target subsystem isolation (default 2)
  -b, --excluded-before symbols    list of symbols to exclude before the target symbol
  -a, --excluded-after symbols     list of symbols to exclude after the target symbol
  -t, --target-subsys subsystems   list of subsystems to include in the output
                                   
  -e, --db-driver driver           database driver: mysql, postgres or sqlite3 (default "postgres")
  -i, --db-instance instance       database instance (default 1)
  -d, --DBDSN DSN                  database DSN in the engine specific format
                                   postgres: "host=dbhost.com port=5432 user=username password=<password> dbname=kernel_bin sslmode=disable"
                                   mysql: "username:@tcp(dbhost.com:3306)/dbname?multiStatements=true"
                                   sqlite3: "file:db_file.db"
                                   
  -h, --help                       show this help message and exit
```

## Sample configuration:
```
{
  "db_driver": "postgres",
  "DBDSN": "host=dbs.hqhome163.com port=5432 user=alessandro password=<password> dbname=kern_bin_new sslmode=disable",
  "db_instance": 1,
  "symbol": "__arm64_sys_getppid",
  "mode":4,
  "excluded_before": [],
  "excluded_after": [".*rcu.*"],
  "max_depth": 1,
  "output_type": "graphOnly",
  "target_subsys": [],
}
```
Configuration is a file containing a JSON serialized conf object

| Field           | description                                                                                               | type     | Default value |
|-----------------|-----------------------------------------------------------------------------------------------------------|----------|---------------|
| db_driver       | Name of DB engine driver, i.e. postgres, mysql or sqlite3                                                 | string   | postgres      |
| DBDSN           | DSN in the engine specific format                                                                         | string   | See Note      |
| db_instance     | Database instance                                                                                         | int      | 1             |
| symbol          | Name of the symbol to start the navigation from                                                           | string   | NULL          |
| mode            | Mode of plotting: 1 symbols, 2 subsystems, 3 subsystems with labels, 4 target subsystem isolation         | integer  | 2             |
| excluded_before | List of symbols to exclude before the target symbol                                                       | string[] | nil           |
| excluded_after  | List of symbols to exclude after the target symbol                                                        | string[] | nil           |
| max_depth       | Max number of levels to explore (0=no limit)                                                              | integer  | 0             |
| output_type     | Type of output: graphOnly, jsonOutputPlain, jsonOutputB64, jsonOutputGZB64                                | enum     | graphOnly     |
| target_subsys   | List of subsys that need to be highlighted. if empty, only the subs that contain the start is highlighted | string[] | nil           | 

