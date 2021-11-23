package main

import (
	"fmt"
	"time"

	run "github.com/iwita/simple-sw/pkg/runtime"
)

func main() {
	start := time.Now()
	r := run.NewRuntime("testfiles/workflow7c.yaml", run.WithInputFile("testfiles/workflow7file.json"))
	r.Start()
	elapsed := time.Since(start)
	fmt.Printf("EXECUTION TIME: %s\n", elapsed) //calculating workflow's execution time
}
