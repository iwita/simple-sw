package runtime

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/itchyny/gojq"
	"github.com/serverlessworkflow/sdk-go/model"
)

func handleEventState(state *model.EventState) error {
	fmt.Println("Event:", state.GetName())
	//TODO
	return nil
}

func handleOperationState(state *model.OperationState) ([]string, error) {
	fmt.Println("Operation:", state.GetName())
	// TODO
	// Check for the action Mode (default: sequential)
	switch state.ActionMode {
	case "sequential":
		fmt.Println("sequential")
		return handleSequentialActions(state), nil
	case "parallel":
		fmt.Println("operation")
	}
	return nil, nil
}

func handleSequentialActions(st *model.OperationState) []string {
	var refs []string
	for _, action := range st.Actions {
		fName := action.FunctionRef.RefName
		// TODO
		// May we assume that there will be 1 action per sequential operation state?
		fmt.Println(fName)
		refs = append(refs, fName)
	}
	return refs
}

func HandleDataBasedSwitch(state *model.DataBasedSwitchState, in []byte) (string, error) {
	for _, cond := range state.DataConditions {
		fmt.Println(cond.GetCondition())
		switch cond.(type) {
		case *model.TransitionDataCondition:
			var result map[string]interface{}
			json.Unmarshal(in, &result)
			op, _ := gojq.Parse(cond.GetCondition())
			iter := op.Run(result)
			v, _ := iter.Next()
			if err, ok := v.(error); ok {
				log.Fatalln(err)
			}
			// fmt.Printf("%v\n", v)
			if v.(bool) {
				fmt.Println("GOTO", cond.(*model.TransitionDataCondition).Transition.NextState)
				return cond.(*model.TransitionDataCondition).Transition.NextState, nil
			} else {
				fmt.Println("Not True")
				return cond.(*model.TransitionDataCondition).Transition.NextState, nil
			}
			// test := map[string]interface{}{"foo": []interface{}{"age", 2, 3}}

			// fmt.Println("Result is:", string(res))

			// return cond.(*model.TransitionDataCondition).Transition.NextState
			// if this condition is true
			// HandleTransition(state, ns)
			//find next state object
			// InferType()
		case *model.EndDataCondition:
			fmt.Println(cond.(*model.EndDataCondition).End)
			// this is the end, you know
		}

	}
	return "", nil
}

func handleInjectState(state *model.InjectState) (string, error) {
	fmt.Println("Inject: ", state.GetName())
	//injectFilter := state.GetStateDataFilter()
	injectData := state.Data
	fmt.Println("Data of inject state: ", injectData)
	//fmt.Println("Input filter: ", injectFilter.Input, " Output filter: ", injectFilter.Output)
	//outFilter := strings.Split(injectFilter.Output, " ")[1]
	//outFilter = strings.Split(outFilter, ".")[1]
	if state.GetTransition() != nil {
		return state.Transition.NextState, nil
	} else {
		fmt.Println("This is the end..")
		return "", nil
	}
}
