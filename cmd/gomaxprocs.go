// +build multicore

/*
The default runtime.GOMAXPROCS in go >= 1.5 will be the number of cpu cores (runtime.NumCPU())
This is just in case that someone builds the package in go < 1.5
Use multicore tag (go build -tags multicore) while building if using go < 1.5 for multi-core support
 */
package main

import "runtime"

func init() {
  numCPU := runtime.NumCPU()
  runtime.GOMAXPROCS(numCPU)
}