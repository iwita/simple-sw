package runtime

import (
	"fmt"

	"github.com/serverlessworkflow/sdk-go/model"
)

func (r *Runtime) Start() {
	r.exec()
}

func (r *Runtime) exec() {
	initState := r.Workflow.States[0]
	r.begin(initState)

}

func (r *Runtime) begin(st model.State) error {
	switch st.(type) {
	case *model.EventState:
		fmt.Println("event")
		// handleEventState()
	case *model.OperationState:
		fmt.Println("operation")
		// handleOperationState(state)
	case *model.EventBasedSwitchState:
		fmt.Println("event based switch")
	case *model.DataBasedSwitchState:
		fmt.Println("data based switch")
		HandleDataBasedSwitch(st.(*model.DataBasedSwitchState), r.lastOutput)
	case *model.ForEachState:
		fmt.Println("foreach")
	case *model.ParallelState:
		fmt.Println("parallel")
	}
	return nil
}
