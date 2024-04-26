/*
 * Copyright (c) 2022 Red Hat, Inc.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

package main

import (
	"os/exec"
	"fmt"
)

// Executes the strip operation on a new file to ease the operations Radare 2 performs on the binary image
func strip(executable string, fn string, outfile string) {
	_, err := exec.LookPath(executable)
	if err != nil {
		panic(fmt.Sprintf("exec.LookPath: %s", err.Error()))
	}
	cmd := exec.Command(executable, "--strip-debug", fn, "-o", outfile)
	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("exec.Command(%s %s %s %s %s): %s", executable, "--strip-debug", fn, "-o", outfile, err.Error()))
	}
}
