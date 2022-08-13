package main
import (
	"os/exec"
)

func strip(executable string, fn string, outfile string){


	_, err := exec.LookPath(executable)
	if err != nil {
		panic(err)
		}

	cmd := exec.Command(executable, "--strip-debug", fn, "-o", outfile)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}


