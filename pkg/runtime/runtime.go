package runtime

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/serverlessworkflow/sdk-go/model"
)

func (r *Runtime) Start() {
	r.exec()
}

func (r *Runtime) exec() {
	initState := r.Workflow.States[0]
	fmt.Println(r.InputFile)
	if r.InputFile != "" {
		jsonFile, _ := os.Open(r.InputFile)
		byteValue, _ := ioutil.ReadAll(jsonFile)
		r.lastOutput = byteValue
	}

	// Store function in a hasmap, to obtain the API endpoint
	r.funcToEndpoint = make(map[string]string, len(r.Workflow.Functions))
	for _, f := range r.Workflow.Functions {
		r.funcToEndpoint[f.Name] = f.Operation
	}

	r.nameToState = make(map[string]model.State, len(r.Workflow.States))
	for _, s := range r.Workflow.States {
		r.nameToState[s.GetName()] = s
	}

	r.begin(initState)

}

func (r *Runtime) begin(st model.State) error {
	switch st.(type) {
	case *model.EventState:
		fmt.Println("event")
		// handleEventState()
	case *model.OperationState:
		// fmt.Println("operation")
		functionRefs, err := handleOperationState(st.(*model.OperationState))
		if err != nil {
			fmt.Println("Error in Operation State")
			return err
		}
		// Call the Function(s) of this state
		// TODO: Maybe we need to assume 1 action per state
		for _, fr := range functionRefs {
			apiCall, _ := r.funcToEndpoint[fr]
			fmt.Println(apiCall)
		}

		// Next we need to determine the next state
		if st.GetTransition() != nil {
			ns := r.nameToState[st.GetTransition().NextState]
			r.begin((ns))
		}

	case *model.EventBasedSwitchState:
		fmt.Println("event based switch")
	case *model.DataBasedSwitchState:
		fmt.Println("data based switch")
		ns, err := HandleDataBasedSwitch(st.(*model.DataBasedSwitchState), r.lastOutput)
		if err != nil {
			fmt.Println("Error here")
		}
		fmt.Println(ns)
		r.begin(r.nameToState[ns])
	case *model.ForEachState:
		fmt.Println("foreach")
	case *model.ParallelState:
		fmt.Println("parallel")
	case *model.InjectState:
		ns, err := handleInjectState(st.(*model.InjectState))
		if err != nil {
			fmt.Println("Error in Inject State")
			return err
		}
		r.begin(r.nameToState[ns])
	}
	return nil
}
