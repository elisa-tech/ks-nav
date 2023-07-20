# Kernel Static-Analysis Navigator - ks-nav

## Overview

This project combines three powerful tools for performing static analysis on
Linux kernels. The toolset is designed to assist developers and engineers in
analyzing kernel source code, understanding function call trees, and
generating informative diagrams.

The three tools included in this project are:

1. **kern_bin_db** - Kernel Source Symbols Extractor and DB Builder: This 
   tool acquires data from a kernel image and builds a structured kernel
   symbol and cross-references database in SQL format.

2. **nav** - Kernel Source Code Navigator: Utilizing the pre-constituted
   database from `kern_bin_db`, `nav` emits call tree graphs Diagrams that can
   be used as standalone representations or integrated into graph display 
   systems.

3. **navweb** - Frontend for Kernel Tools: The navweb frontend acts as a 
   user-friendly web interface for interacting with the nav tool. It
   seamlessly integrates with the nav tool, allowing users to perform
   static analysis through a browser-based interface. The frontend
   provides interactive diagrams to visualize data acquired from kernel
   images and enhances the overall user experience.

## Building and Usage


## Contributing

We welcome contributions to improve the Kernel Static-Analysis Navigator. If
you are interested in contributing, please refer to the project's
contribution guidelines for detailed instructions on how to get started.

Let's build a powerful toolset together for better kernel analysis and 
development!

