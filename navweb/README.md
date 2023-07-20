# navweb Frontend for Kernel tools

## Project Overview

This application frontend is part of a larger project that aims to provide 
a static analysis tool for the kernel. The toolset consists of three 
applications, one of which is this frontend. 
The other two applications are:

* `kern_bin_db`: A tool designed to acquire data from a kernel image.
* `nav`: A tool that utilizes the acquired data to generate diagrams.

The purpose of this frontend application is to provide a web interface for 
interacting with the nav tool. It acts as a wrapper around nav and enables 
users to conveniently access and utilize the functionality of the 
underlying analysis tool through a browser-based interface.

## Functionality and Purpose

The `navweb` frontend serves as an interface to the nav tool, allowing users 
to perform static analysis on kernel images and generate diagrams. 
The key features and objectives of this tool include:

1. Web Interface: Provides a user-friendly web-based interface for 
   interacting with the static analysis tool.
2. Seamless Integration: Wraps the nav tool, ensuring smooth integration 
   with the underlying analysis functionalities.
3. Interactive Diagrams: Enables users to visualize the acquired data from 
   kernel images through intuitive and interactive diagrams.
4. User-Focused Design: Prioritizes ease of use, accessibility, and efficient 
   analysis workflows to enhance the user experience.

## Building and Usage

To build and use the `navweb` frontend, follow the steps outlined below:

* Clone the repository: `git clone <repository_url>`
* Navigate to the project directory: cd `<project_directory>`
* Modify the necessary configuration files to match your environment and 
  requirements.
* Build the application:
    * Run `make` to build the frontend application.
    * Start the application:
        * Execute `./navweb` to start the frontend application, it listen 
          at `8080`.
    * Access the application: Open your web browser and visit the running 
      machine URL or localhost address where the application is running at 
      `8080`.
    * Verify `nav` app to work properly.

Please note that these instructions provide a basic outline, and further 
details specific to your environment may be required. Refer to the project 
documentation or consult the project team for any additional information or 
troubleshooting steps.
