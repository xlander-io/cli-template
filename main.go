package main

import (
	"os"

	"github.com/coreservice-io/cli-template/cmd"
)

func main() {

	//config app to run
	errRun := cmd.ConfigCmd().Run(os.Args)
	if errRun != nil {
		panic(errRun)
	}
}
