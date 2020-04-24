package main

import "fmt"

var (
	release    = "not set"
	buildstamp = "not set"
	githash    = "not set"
)

func showVersion() {
	fmt.Println("Release: ", release)
	fmt.Println("Git hash: ", githash)
	fmt.Println("Build time: ", buildstamp)
}
