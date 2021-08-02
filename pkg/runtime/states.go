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

			} else {
				fmt.Println("Not True")
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
	return nil
}
