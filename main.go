package main

import (
	"fmt"

	st "github.com/iwita/simple-sw/pkg/states"
	"github.com/serverlessworkflow/sdk-go/model"
	"github.com/serverlessworkflow/sdk-go/parser"
)

func ParseWorkflow(filePath string) (*model.Workflow, error) {
	workflow, err := parser.FromFile(filePath)
	if err != nil {
		return nil, err
	}
	return workflow, nil
}

func main() {
	if _, err := ParseWorkflow("testfiles/test2.yaml"); err != nil {
		fmt.Println(err)
	}
	wf, _ := ParseWorkflow("testfiles/test2.yaml")

	// Start from the start state
	startState := wf.States[0]
	if err := st.InferType(startState); err != nil {
		fmt.Println(err)
	}
}
