# GETTING STARTED WITH KS-NAV - A BASIC TUTORIAL

## Table of Contents

- [GETTING STARTED WITH KS-NAV - A BASIC TUTORIAL](#getting-started-with-ks-nav---a-basic-tutorial)
  - [Table of Contents](#table-of-contents)
  - [1 Overview](#1-overview)
  - [2 A quick tutorial](#2-a-quick-tutorial)
    - [2.1 Build Linux for ks-nav](#21-build-linux-for-ks-nav)
    - [2.2 Run ks-nav on the built image](#22-run-ks-nav-on-the-built-image)
  - [3 Detailed Tools Configurations](#3-detailed-tools-configurations)
    - [3.1 Configuring kern\_bin\_db and the local database](#31-configuring-kern_bin_db-and-the-local-database)
    - [3.2 Configuring nav and navweb for caller graph generation](#32-configuring-nav-and-navweb-for-caller-graph-generation)

## 1 Overview

To start using ks-nav, a Linux (or Yocto) build with debug symbols enabled is required. The tool then extracts symbols, functions and any useful data into a PostgreSQL database. 

In this tutorial, you will:

- Build a default branch from Linux - tag v6.6 is used in this tutorial
- Extract symbols into a local database using ksnav-kdb
- Run the navigator using ks-nav

## 2 A quick tutorial

### 2.1 Build Linux for ks-nav

Skip this step if you already have a valid Linux build.
The **demo/linux_app** folder includes a pre-configured Linux Configuration file and a Dockerfile.
Use the commands below to create the container that will build the Linux image as defined in the configuration file **linux_config**.

```bash 

    # Move to the demo folder (should be the current directory) 
    cd ./demo
    # delete the ksbuild folder if you want to start with a new environment
    rm -rf ksbuild
    # Run script setup_linux_app to build a vmlinux image for ks-nav
    # The built image should be available under ./demo/ksbuild/app
    ./setup_linux_app
```

At the end of this step, the **ksbuild/app** folder should be populated with the Linux source and build artifacts. 

### 2.2 Run ks-nav on the built image

This step assumes that the **ksbuild/app** folder contains both Linux source and build artifacts.
The artifacts generated from the Linux build in the previous step are now passed to another container - the ksnav container.

```bash

    # Move to the demo folder (should be the current directory)
    cd demo
    # Run the run_ksnav script to build a ks-nav container
    ./run_ksnav
    # Access the container's shell (assuming $LINUX_CONT='linux-cont' in setup_linux_app)
    podman exec -it ksnav-cont sh
    # From the interractive shell, run the following command to:
    # - Extract symbols into a local database using kern_bin_db - the Kernel Source Symbols Extractor and DB Builder
    # - This takes a little while
    cd /app
    nav-db-filler /build/ksnav_kdb.json    
    # - Run the Source code Navigator using nav - which generates call tree graphs. 
    # The graphs are generated via script in the ksbuild/out_img folder
    nav -f /build/ksnav_nav.json > /out_img/run_init_process

```

At the end of this step, folder **./demo/ksbuild/out_img** should be populated with at least one image file of a call tree graph.

## 3 Detailed Tools Configurations

### 3.1 Configuring kern_bin_db and the local database

ks-nav is comprised of three different tools. The first tool to run  **kern_bin_db**. It acquires data from a kernel image and builds a structured kernel symbol and cross-reference database in SQL format. 
In this section, we dive into the configuration of **kern_bin_db** for this demo. 

The kern_bin_db main source code is in *ks-nav/kern_bin_db/main.go*. Once successfully built, it generates the ***nav-db-filler*** executable. 
The Makefile under folder *ks-nav/kern_bin_db/* provides other build options. Once the executable is generated, the tool requires the following inputs to run successfully: 
- a json configuration file to run (we'll use file '*./demo/ksnav_app_linux/ksnav_kdb.json*' for this example)
- a Linux image with debug symbols (located in the /app folder in the demo container)

File *ksnav_kdb.json* is shown below: 
```bash
{
"LinuxWDebug":		"/app/vmlinux", 
"LinuxWODebug":		"/app/vmlinux.work",
"StripBin":		"/usr/bin/strip",
"DBDriver":		"postgres",
"DBDSN":		"host=localhost port=5432 user=postgres password=my123 dbname=kernel_bin sslmode=disable",
"Maintainers_fn":	"/app/MAINTAINERS",
"KConfig_fn":		"include/generated/autoconf.h",
"KMakefile":		"/app/Makefile",
"Mode":			15,
"Note":			"upstream"
}LinuxWDebug: The path to the linux binary unstripped of debug symbols


```
- *LinuxWDebug*: The path to the linux binary unstripped of debug symbols
- *LinuxWODebug*: The path to the Linux binary once stripped of debug symbols. In this example, the default strip command in the container is used. When using a cross-compiled image, the appropriate strip library must be used. The path to the strip command to use is defined in the next config parameter
- *StripBin*: The path to the strip command used to generate 'LinuxWODebug'
- *DBDriver*: The specified Database to use (can be either postgres, mysql or sqlite3)
- *DBDSN*: The details required to connect to the database. In this demo, the database is local to the container. 
- *Maintainers_fn*: ks-nav can organize diagrams by sub-systems. The Maintainers file is parsed for that purpose. This config param is the path to Maintainers file. 
- *KConfig_fn*: The path to the build configuration 
- *KMakefile*: This should point to the Makefile
- *Mode*: This is a mode of operation for debug purposes - linked to the database installed (which in this case is postgres:15). Should be defaulted to 15 in this demo. 
- *Note*: This is the string copied to the database - a way to tag the dataset

In Yocto environments, the Linux source and Linux build folders usually tend to be different paths. The user should take care to point to the appropriate path. 

File *'./demo/ksnav_app_linux/postgres_conf_template.conf'* is the file used to configure the database in the container. It is copied into the container's /tmp folder. 
 
```bash
data_directory = '%POSTGRES_DATA_DIR%/%POSTGRES_NAME%'
hba_file = '%POSTGRES_DATA_DIR%/%POSTGRES_NAME%/pg_hba.conf'
ident_file = '%POSTGRES_DATA_DIR%/%POSTGRES_NAME%/pg_ident.conf'
external_pid_file = '/tmp/%POSTGRES_NAME%.pid'
listen_addresses = '*'
port = 5432
max_connections = 100
unix_socket_directories = '/tmp'
ssl = off
shared_buffers = 128MB
dynamic_shared_memory_type = posix
max_wal_size = 1GB
min_wal_size = 80MB
log_line_prefix = '%m [%p] %q%u@%d '
cluster_name = 'only'
datestyle = 'iso, mdy'
lc_messages = 'en_US.UTF-8'
lc_monetary = 'en_US.UTF-8'
lc_numeric = 'en_US.UTF-8'
lc_time = 'en_US.UTF-8'
default_text_search_config = 'pg_catalog.english'
```

To generate the database for ks-nav, run the *nav-db-filler* with the *ksnav_kdb.json* file as shown below. The command should work as described. However, if there are any running the tool successfully, consider running the command from the /app folder where the Linux source and build artifacts are located. 

```bash
  nav-db-filler <path_to_ksnav_kdb.json_file>
```

It is strongly encouraged to double-check that all paths mentioned in the json and postgres_conf_template.conf files are valid  before running the tool.

### 3.2 Configuring nav and navweb for caller graph generation

The nav tool is used to generate different diagrams using the symbols and cross-references from the database generated in the previous section. The ***nav*** tool is a command-line tool that has a configurable user interface: ***navweb*** tool. 

The ***nav*** main source code is in *ks-nav/nav/nav.go*. Once successfully built, it generates the ***nav*** executable. 
The Makefile under folder *ks-nav/nav/* provides other build options. Once the executable is generated, the tool requires the following inputs to run successfully: 
- a json configuration file to run (we'll use file '*./demo/ksnav_app_linux/ksnav_nav.json*' for this example)
- a database generated through the ***kern_bin_db*** tool. This database is locally hosted in the demo container generated by *./demo/ksnav_app_linux/Dockerfile_ksnav*

File *ksnav_nav.json* is shown below: 

```bash
{
  "mode":1,
  "excluded_before": [],
  "excluded_after": [".*rcu.*"],
  "target_subsys":[],
  "max_depth":2,
  "output_type": "graphOnly",
  "out_type":3,
  "symbol": "run_init_process",
  "db_driver": "postgres",
  "DBDSN": "host=localhost port=5432 user=postgres password=my123 dbname=kernel_bin sslmode=disable",
  "db_instance": 1
}
```

The field options are detailed in [this README](../nav/README.md#sample-configuration). Here are the main options that were used in this demo to generate a JPG image of a symbol caller graph via ***command line***. 

- *mode*: It is set to 1 to visualize symbols
- *max_depth*: The Max number of levels to explore
- *output_type*: Type of output to generate. The *generateOutput()* function in *./ks-nav/nav/nav.go* provides more insight into the meaning of these options.
- *out_type*: When set to 1, 2 and 3, this field generates  DOT (ASCII), PNG and JPG files, respectively.
- *symbol*: The symbol to start the navigation from 
- *db_driver*: The specified Database to use (can be either postgres, mysql or sqlite3). This field should match the entry in the kern_kdb json file.
- *DBDSN*: The details required to connect to the database. In this demo, the database is local to the container. This field should match the entry in the kern_kdb json file.
- *db_instance*: Default to 1. The Database instance.

If the user plans to use the web interface, it is recommended to copy the *./demo/ksnav_app_linux/ksnav_nav.json* file into the **navweb** tool's json file (*./ks-nav/navweb/data/configs/container.json*) and re-build both **nav** and **navweb** tools as shown in the Dockerfile. This step will propagate into the **navweb** tool, all updates made in the **nav** tool, that are not editable via the web interface.
 
Once the **nav** and **navweb** tools are successfully built, run the command line shown below to generate an image of the caller graph in an output filename. In this demo, it is mapped to the *./ks-nav/demo/ksbuild/out_img* folder. 

```bash
nav -f </build/>ksnav_nav.json > <path_to_outputfile>
```

The **navweb** tool yields similar images from its user interface through http://localhost:8080 (the default web interface port in the current configuration). Click on Explore to try the tool out with the same entries as the nav json file above. 
Note that the Dockerfile calls the *navweb* command to start the webserver.

The dot file generated when *out_type=1* in the *ksnav_nav.json* file is shown below.

```bash
digraph G {
rankdir=LR; node [style=filled fillcolor=yellow]
"run_init_process"->"_printk" [ edgeid = "1"]; 
"_printk"->"vprintk" [ edgeid = "2"]; 
"vprintk" [style=filled; fillcolor=red];
"_printk"->"__stack_chk_fail" [ edgeid = "3"]; 
"__stack_chk_fail" [style=filled; fillcolor=red];
"run_init_process"->"kernel_execve" [ edgeid = "4"]; 
"kernel_execve"->"count_strings_kernel.part.0" [ edgeid = "5"]; 
"count_strings_kernel.part.0" [style=filled; fillcolor=red];
"kernel_execve"->"bprm_execve" [ edgeid = "6"]; 
"bprm_execve" [style=filled; fillcolor=red];
"kernel_execve"->"free_bprm" [ edgeid = "7"]; 
"free_bprm" [style=filled; fillcolor=red];
"kernel_execve"->"alloc_bprm" [ edgeid = "8"]; 
"alloc_bprm" [style=filled; fillcolor=red];
"kernel_execve"->"copy_string_kernel" [ edgeid = "9"]; 
"copy_string_kernel" [style=filled; fillcolor=red];
"kernel_execve"->"copy_strings_kernel" [ edgeid = "10"]; 
"copy_strings_kernel" [style=filled; fillcolor=red];
"kernel_execve"->"putname" [ edgeid = "11"]; 
"putname" [style=filled; fillcolor=red];
"kernel_execve"->"getname_kernel" [ edgeid = "12"]; 
"getname_kernel" [style=filled; fillcolor=red];
}
```
