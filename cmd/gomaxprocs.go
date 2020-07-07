// +build multicore

package main

import "runtime"

func init() {
  numCPU := runtime.NumCPU()
  runtime.GOMAXPROCS(numCPU)
}