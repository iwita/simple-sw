package runtime

import (
	"fmt"

	"github.com/serverlessworkflow/sdk-go/model"
)

type RuntimeOption func(*Runtime)

func WithInputFile(file string) RuntimeOption {
	return func(r *Runtime) {
		r.InputFile = file
	}
}

func WithUser(user string) RuntimeOption {
	return func(r *Runtime) {
		r.User = user
	}
}

func WithNamespace(namespace string) RuntimeOption {
	return func(r *Runtime) {
		r.Namespace = namespace
	}
}

func NewRuntime(sw string, options ...RuntimeOption) *Runtime {
	if _, err := ParseWorkflow(sw); err != nil {
		fmt.Println(err)
	}

	wf, _ := ParseWorkflow(sw)

	run := &Runtime{Workflow: wf, User: "def", Namespace: "def"}

	for _, o := range options {
		o(run)
	}
	return run
}

type Runtime struct {
	Workflow   *model.Workflow
	User       string
	Namespace  string
	InputFile  string
	lastOutput []byte
}
