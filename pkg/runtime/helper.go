package runtime

import (
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
