package main

import (
	run "github.com/iwita/simple-sw/pkg/runtime"
)

func main() {
	_ = run.NewRuntime("testfiles/applicationrequest.yaml", run.WithInputFile("testfiles/applicant.json"))

}
