# Containerized Project Integration with Multiple Applications

## Overview
This repository contains a containerized project comprising three applications:
`navweb`, `nav`, and `kern_bin_db`. 
The containerization process makes it easy to build and run the project as a 
container, ensuring seamless integration and deployment.

## Key Facts

* `Dockerfile`: We have included a Dockerfile to facilitate containerization, 
  allowing you to build the project as a container effortlessly.
* SQL Initialization: The SQL initial creation statements have been separated 
  into two parts for improved organization, ensuring a smoother setup process.
* Database Defaults: Database defaults have been adjusted to align perfectly 
  with container requirements, simplifying the overall deployment.
* Kernel Acquisition Tool: A link to the kernel acquisition tool has been 
  incorporated, making it easier to fetch the required kernel during runtime.
* Navweb Functionality: The landing page and fetch page functionalities for 
  the navweb application have been implemented, providing essential features 
  for the project.
* Makefile Changes: Necessary adjustments have been made to the makefiles to 
  ensure a successful container build, reducing potential build issues.
* go-bind Tool: We have introduced the go-bind tool, which significantly 
  reduces the number of files transferred, improving the overall efficiency 
  of the application.
* Default Sample Database: For ELISA Tellteal use case, a default sample 
  database has been included, allowing you to get started quickly.
* Nav Regression Issue: We resolved a nav regression issue introduced by a 
  recent config pull request, ensuring a stable and bug-free experience.

## Usage Guide

A step-by-step example describing how to run ks-nav from a container with a locally hosted database, is available in the `./demo` [folder](./demo/README.md). The steps below represent a generic usage guide when the database is hosted outside of the ks-nav container. 

### Build the ks-nav Container
To build the container, follow these steps:

- Navigate to the container directory. Type the following command:

```bash
        podman build -v <local directory for postgres data>:/var/lib/postgresql/data:z -t ks-nav .
```
In the command above, `<local directory for postgres data>` is the path to the directory where you 
  are storing PostgreSQL data on your host or remote system, outside of the ks-nav container. 
  This parameter is not needed if you intend to run the database within the ks-nav container. 

The build phase initializes the basic database. Two options are provided:
  * An empty database, which is used as the default option.
  * An already initialized database containing the ELISA Tell Tale use case.

### Run the Application

After successfully building the ks-nav container, you can run the ks-nav application. This application requires the user to have a Linux image built with debug symbols enabled. One example on how to build a Linux image for ks-nav is available in the [demo folder](./demo/README.md#21-build-linux-for-ks-nav). Once the image is successfully built, you can run the ks-nav application, using the following steps:

- Type the following command in the container directory: 

```bash
        podman run -it -p 5432:5432 -p 8080:8080 \
          -v <linux kernel build directory>:/app:z \ 
          -v <local directory for postgres data>:/var/lib/postgresql/data:z \
          localhost/ks-nav:latest
```

In the command above, 
    * `<linux kernel build directory>` is the path to the directory containing the Linux kernel build.
    * `<local directory for postgres data>` is the same directory used during the container build or any other directory containing PostgreSQL data suitable for the application. This argument is not needed if the database is locally hosted in the ks-nav container. 
    * If you do not intend to fetch a new database during runtime, you can set the 
  `<linux kernel build directory>` to /tmp.

By following these steps, you can easily set up and run the containerized project 
with its multiple applications. Should you encounter any issues or require further 
assistance, feel free to reach out for support. 
