package states

import (
	"fmt"

	"github.com/serverlessworkflow/sdk-go/model"
)

func InferType(st model.State) error {
	switch st.(type) {
	case *model.EventState:
		fmt.Println("event")
		// handleEventState()
	case *model.OperationState:
		fmt.Println("operation")
		// 	// handleOperationState(state)
	case *model.EventBasedSwitchState:
		fmt.Println("event based switch")
	case *model.DataBasedSwitchState:
		fmt.Println("data based switch")
		HandleDataBasedSwitch(st.(*model.DataBasedSwitchState))
	case *model.ForEachState:
		fmt.Println("foreach")
	case *model.ParallelState:
		fmt.Println("parallel")
	}
	return nil
}

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

func HandleDataBasedSwitch(state *model.DataBasedSwitchState) error {
	for _, cond := range state.DataConditions {
		fmt.Println(cond.GetCondition())
		switch cond.(type) {
		case *model.TransitionDataCondition:
			fmt.Println(cond.(*model.TransitionDataCondition).Transition.NextState)
			// if this condition is true
			// GOTO HandleTransition
			//find next state object
			// InferType()
		case *model.EndDataCondition:
			fmt.Println(cond.(*model.EndDataCondition).End)
			// this is the end, you know
		}

	}
	return nil
}
