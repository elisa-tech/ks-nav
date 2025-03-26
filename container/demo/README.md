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
#### Podman 
---
```bash 
    # Move to the demo folder (should be the current directory: ks-nav/container/demo)
    cd ./demo
    # Optional step: Delete the ksbuild folder to start with a new environment
    rm -rf ksbuild
    # Create an artifacts folder
    mkdir ksbuild
    # Build Linux image with debug symbols enabled 
    podman build -t linux-img -f Dockerfile_linux_app .
    podman run -it --name linux-cont linux-img sh
    podman cp linux-cont:/app $(pwd)/ksbuild/app
```
#### Docker
---
```bash 
    # Move to the demo folder (should be the current directory: ks-nav/container/demo)
    cd ./demo
    # Optional step: Delete the ksbuild folder to start with a new environment
    rm -rf ksbuild
    # Create an artifacts folder
    mkdir ksbuild
    # Build Linux image with debug symbols enabled 
    docker build -t linux-img -f Dockerfile_linux_app .
    docker run -it --name linux-cont linux-img sh
    docker cp linux-cont:/app $(pwd)/ksbuild/app
```

At the end of this step, the **./demo/ksbuild/app** folder should be populated with the Linux source code and build artifacts. 

### 2.2 Run ks-nav on the built image

This step assumes that the **./demo/ksbuild/app** folder contains both Linux source and build artifacts.
The artifacts generated from the Linux build in the previous step are now passed to another container - the ksnav container.

```bash
    # Move to the container folder (ks-nav/container)
    cd container
    # Create an output folder for ksnav images
    mkdir ./demo/ksbuild/out_img
    # Copy the ksnav config files to the /app folder 
    cp ksnav_kdb_local.json ./demo/ksbuild/app/.
    cp ksnav_nav_local.json ./demo/ksbuild/app/.
```
ks-nav can be run through a CLI, or through a web interface. We'll go through both options in this demo. 

#### 2.2.1 Run ks-nav from CLI

```bash
    # OPTION 1: RUN KS-NAV FROM CLI 
    # These steps should start from the ks-nav/container directory
    # Comment out the ENTRYPOINT in the Dockerfile to disable the launch of the web server
    cat Dockerfile | \
        sed -r "s|ENTRYPOINT|#ENTRYPOINT |g" | tee Dockerfile_clidemo
    # Build the ksnav container in demo mode
    podman build -t ksnav-img -f Dockerfile_clidemo . --build-arg _EN_DEMO=true
    # Run the container in interractive mode, mapping the linux source code and image output folders
    podman run -it --name ksnav-cont -p 5432:5432 -p 8080:8080 \
     -v $(pwd)/demo/ksbuild/app:/app:z \
     -v $(pwd)/demo/ksbuild/out_img:/out_img:z \
     ksnav-img sh
    # Start the database service (if it wasn't already running)
    su postgres 
    pg_ctl start -D /var/lib/pgsql/data/only/
    exit 
    # Build the database for ks-nav - this could take some time
    cd /app
    nav-db-filler -f ksnav_kdb_local.json    
    # Run the Source code Navigator which generates call tree graphs. 
    # The graphs generated are redirected to a file in the ./demo/ksbuild/out_img folder
    nav -f ksnav_nav_local.json > /out_img/run_init_process
```
Exit the container. 
Folder **demo/ksbuild/out_img** is now populated with a JPEG image *run_init_process*.

#### 2.2.2 Run ks-nav from Web User Interface

We will generate the same graph from the CLI section; this time, from the web user interface. 
```bash
    # Move to the container directory (ks-nav/container)
    cd container
    podman stop ksnav-cont; podman rm ksnav-cont
    # Build the Docker image for ksnav as shown below
    podman build -t ksnav-img -f Dockerfile . --build-arg _EN_DEMO=true
    # Run it in interractive mode, mapping the linux image and the output folder
    podman run -d --name ksnav-cont -p 5432:5432 -p 8080:8080 \
     -v $(pwd)/demo/ksbuild/app:/app:z \
     -v $(pwd)/demo/ksbuild/out_img:/out_img:z \
     ksnav-img sh
    # The web server should now be running
```
Open a web browser at: http://localhost:8080 and click on `Acquire`
The script will build the kern-db database. This could take a while.
Once the command is finished executing, go back to http://localhost:8080 and click on `Explore`
Fill out the form as follow: 

```bash
    start_symbol = run_init_process
    instance = 1
    display mode = Functions Mode
    depth = 2
    Click on 'Generate Image'
```

At the end of this step, the same image from the CLI step is displayed on the web interface.

## 3 Detailed Tools Configurations

### 3.1 Configuring kern_bin_db and the local database
We configured **kern_bin_db** using a json file *ksnav_kdb_local.json*. However, it may be out of date; in which case the [Sample Configuration](../../kern_bin_db/README.md#sample-configuration) section in the **kern_bin_db** [README](../../kern_bin_db/README.md) has additional details on parameter/value updates. 

File *ksnav_kdb_local.json* is shown below: 
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
	}
```
In this demo, the DBDSN points to the container-hosted database. 
In Yocto environments, the Linux source and Linux build folders usually tend to be in different locations. The user should take care to update the paths accordingly. 

File *'./demo/ksnav_app_linux/postgres_conf_template.conf'* used in this demo is shown below. 
It may also be out of date. When in doubt, refer to its latest version available in the [container](../../container/postgres_conf_template.conf) folder. 

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

To generate the database for ks-nav, run the *nav-db-filler* command with the *ksnav_kdb_local.json* configuration file as shown below.  

```bash
  nav-db-filler -f <path_to_ksnav_kdb.json_file>
```

It is strongly encouraged to double-check that all paths mentioned in the json and postgres_conf_template.conf files are valid  before running the tool.

### 3.2 Configuring nav and navweb for caller graph generation

We configured the **nav** and **navweb** tools using json file *ksnav_nav_local.json*. However, it may be out of date; in which case, the [Sample Configuration](../../nav/README.md#sample-configuration) section in the **nav** [README](../../nav/README.md) has additional details on parameter/value updates. 

File *ksnav_nav_local.json* is shown below. It generates a JPG image of a symbol caller graph. 
To generate a DOT file instead, *out_type* would need to be set to 1.
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
If the user plans to use the web interface, it is recommended to use the same *ksnav_nav_local.json* file for both **navweb** and **nav** tools. Save the **navweb**'s config as: *./ks-nav/navweb/data/configs/container.json* and re-build both tools as shown in the Dockerfile. 
This step will propagate into the **navweb** tool, all updates made in the **nav** tool, that are not editable via the web interface.

After building the **nav** tool, run the command line shown below to generate a caller graph in an output filename. In this demo, the output image is saved in the *./ks-nav/demo/ksbuild/out_img* folder. 

```bash
nav -f </build/>ksnav_nav.json > <path_to_outputfile>
```

The dot file generated when *out_type=1* in the *ksnav_nav_local.json* file is shown below.

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


