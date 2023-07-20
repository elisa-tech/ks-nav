# kern_bin_db - Kernel Source Symbols Extractor and DB Builder

## Motivation
`kern_bin_db` has been developed to produce a structured kernel symbol and
cross-references database (SQL). This database can aid manual source code
analysis, simplify the source code, and enable automated checks, such as
scanning for recursive functions and other queries.

## Prerequisites
To develop this project, you need to install the `radare2` development package:

### Fedora

```bash
$ sudo dnf install radare2-devel
```

### Arch Linux

```bash
$ sudo pacman -S radare2
```


### Debian

```bash
$ sudo apt-get install radare2
```
#### Disclaimer

Since the radare2 dependency is very volatile, and the presented interface can
change frequently, it is suggested to check the radare2 version before issuing
any command to the tool. Last known version:

## Build

`kern_bin_db` is implemented in Golang, and Golang applications are typically
easy to build with few dependencies other than the standard Golang packages.
A Makefile is provided to ease the build process.

To build the native executable, simply run:

```bash
$ make
```

In addition to the default target, `amd64`, `arm64`, and `upx` targets exist.
For example, to build an aarch64 UPX-compressed executable:

```bash
$ make arm64
$ make upx
```

## Usage example

The following command line switches are available in the `kern_bin_db` tool:

```bash
$ ./kern_bin_db -h
Kernel symbol fetcher
	-f	<v>	specifies json configuration file
	-s	<v>	Forces use specified strip binary
	-e	<v>	Forces to use a specified DB Driver (i.e. postgres, mysql or sqlite3)
	-d	<v>	Forces to use a specified DB DSN
	-n	<v>	Forecs use specified note (default 'upstream')
	-c		Checks dependencies
	-h		This Help
```

After compiling with the default PostgreSQL database backend specified, you
can start collecting symbols by issuing the following command:

```bash
$ ./kern_bin_db -f conf.json -n "Custom kernel from NXP bsp"
```

## Sample Configuration

The conf.json file contains a JSON serialized configuration object.
Here is an example configuration:

```json
{
    "LinuxWDebug": "vmlinux",
    "LinuxWODebug": "vmlinux.work",
    "StripBin": "/usr/bin/aarch64-linux-gnu-strip",
    "DBDriver": "postgres",
    "DBDSN": "host=dbs.hqhome163.com port=5432 user=alessandro password=<password> dbname=kernel_bin sslmode=disable",
    "Maintainers_fn": "MAINTAINERS",
    "KConfig_fn": "include/generated/autoconf.h",
    "KMakefile": "Makefile",
    "Mode": 15,
    "Note": "upstream"
}
```
Configuration is a file containing a JSON serialized conf object

|Field         |description                                                                         |type    |Default value               |
|--------------|------------------------------------------------------------------------------------|--------|----------------------------|
|LinuxWDebug   |Linux image built with the debug symbols, input for the operation                   |string  |vmlinux                     |
|LinuxWODebug  |File created after the strip operation, and on which the R" tool operates on        |string  |vmlinux.work                |
|StripBin      |Executable that performs the strip operation to the selected architecture.          |string  |/usr/bin/strip              |
|DBDriver      |Name of DB engine driver, i.e. postgres, mysql or sqlite3                           |string  |postgres                    |
|DBDSN         |DSN in the engine specific format                                                   |string  |See Note                    |
|Maintainers_fn|The path to MAINTAINERS file, typically in the kernel source tree                   |string  |MAINTAINERS                 |
|KConfig_fn    |The path to autoconf file containing the current build configuration                |string  |include/generated/autoconf.h|
|KMakefile     |The path to main kernel sourcecode Makefile, typically sitting on the kernel tree / |string  |Makefile                    |
|Mode          |Mode of operation, use only for debug purpose. Defaults to 15                       |integer |15                          |
|Note          |The string gets copied to the database. Consider a sort of tag for the data set     |string  |upstream                    |

### DSN Examples

| DBMS          | Example                                                                                                |
|---------------|--------------------------------------------------------------------------------------------------------|
| MySQL/MariaDB | alessandro:<password>@tcp(dbs.hqhome163.com:3306)/kernel_bin?multiStatements=true                      |
| Postgresql    | host=dbs.hqhome163.com port=5432 user=alessandro password=<password> dbname=kernel_bin sslmode=disable |

**Note**: Defaults are designed to make the tool work out of the box if the
executable is placed in the Linux kernel source code root directory. Before
starting `kern_bin_db`, the DB schema needs to be created manually with one of
the provided SQL files.
