package runtime

import (
	"fmt"

	jq "github.com/savaki/jq"
	"github.com/serverlessworkflow/sdk-go/model"
)

func handleEventState(state *model.EventState) error {
	fmt.Println("Event:", state.GetName())
	//TODO
	return nil
}

func handleOperationState(state *model.OperationState) error {
	fmt.Println("Operation:", state.GetName())
	// TODO
	// Check for the action Mode (default: sequential)
	return nil
}

func HandleDataBasedSwitch(state *model.DataBasedSwitchState, in []byte) error {
	for _, cond := range state.DataConditions {
		fmt.Println(cond.GetCondition())
		switch cond.(type) {
		case *model.TransitionDataCondition:
			op, err := jq.Parse(cond.GetCondition())
			if err != nil {
				fmt.Printf("Error in data based switch", err)
			}
			if op.Apply(in) {
				fmt.Println(cond.(*model.TransitionDataCondition).Transition.NextState)
				// return cond.(*model.TransitionDataCondition).Transition.NextState
				return nil
			}
			// if this condition is true
			// HandleTransition(state, ns)
			//find next state object
			// InferType()
		case *model.EndDataCondition:
			fmt.Println(cond.(*model.EndDataCondition).End)
			// this is the end, you know
		}

	}
	return nil
}
