# nav - Kernel Source Code Navigator

## Motivation
`nav` is a powerful tool designed to assist with kernel source code analysis
by generating call tree graphs. These graphs provide valuable insights into
feature-level analysis and showcase functions call trees or subsystems call
trees.

## Prerequisites
Before using `nav`, ensure that you have already built and configured the 
`kern_bin_db` tool to acquire data from the kernel image. 
`nav` utilizes the pre-constituted database from `kern_bin_db` for generating
diagrams.

## Build
`nav` is implemented in Golang, and the build process is straightforward. 
A `Makefile` is provided to ease the build process.

To build the native executable, run:

```bash
$ make
```

In addition to the default target, `amd64`, `arm64`, and `upx` targets exist. 
For example, to build an aarch64 UPX-compressed executable:

```bash
$ make arm64
$ make upx
```

## Usage Example

To use `nav`, you need to provide the configuration to query the backend
database. The easiest way to do this is by specifying a configuration file.
While `nav` has internal defaults for configuration parameters, these defaults
can be overridden by both the configuration file or command line switches.

The configuration file is a plain JSON object and can be passed using the 
command line switch `-f`. The order in which the configuration is evaluated is
as follows:

```sql
+-------------------+    +--------------+   +----------+
|Nav builtin default|--->|conf json file|-->|CLI switch|
+-------------------+    +--------------+   +----------+
```

For example, to start the navigation from the symbol `start_kernel` using the
configuration in `conf.json`:

```bash
$ ./nav -f conf.json -s start_kernel
```

## Command Line Switches

The following command line switches are available in the nav tool:

```
$ ./nav -h
command help
App Name: nav
Descr: kernel symbol navigator
	-j	<v>	Force Json output with subsystems data
	-s	<v>	Specifies symbol
	-i	<v>	Specifies instance
	-f	<v>	Specifies config file
	-e	<v>	Forces to use a specified DB Driver (i.e. postgres, mysql or sqlite3)
	-d	<v>	Forces to use a specified DB DSN
	-m	<v>	Sets display mode 2=subsystems,1=all
	-x	<v>	Specify Max depth in call flow exploration
	-g	<v>	if -j option is graphOnly (default) output PNG in place of dot
	-h		This help
```

## Sample Configuration
The `conf.json` file contains a JSON serialized configuration object. 
Here is an example configuration:

```json
{
  "db_driver": "postgres",
  "DBDSN": "host=dbs.hqhome163.com port=5432 user=alessandro password=<password> dbname=kern_bin_new sslmode=disable",
  "db_instance": 1,
  "symbol": "__arm64_sys_getppid",
  "mode": 4,
  "excluded_before": [],
  "excluded_after": [".*rcu.*"],
  "max_depth": 1,
  "output_type": "graphOnly",
  "target_subsys": []
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

