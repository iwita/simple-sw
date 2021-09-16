package runtime

import (
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"
	"strings"
	"bytes"
	"crypto/tls"
	"sync"
	//"time"

	"github.com/itchyny/gojq"
	"github.com/serverlessworkflow/sdk-go/model"
)

func handleEventState(state *model.EventState, r *Runtime) error {
	fmt.Println("--> Event:", state.GetName())
	if state.GetTransition() != nil {
		ns := state.Transition.NextState
		r.begin(r.nameToState[ns])
		return nil
	} else {
		fmt.Println("This is the end..")
		return nil
	}
}

func handleOperationState(state *model.OperationState, r *Runtime) error {
	fmt.Println("--> Operation:", state.GetName())
	functionRefs := handleSequentialActions(state) //getting the funcRefs of this op.state

	//skipping http security protocol for openwhisk cli
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// TODO
	// Check for the action Mode (default: sequential)
	switch state.ActionMode {
	case "sequential": //assuming 1 operation state => multiple dependable actions or just one independent
		fmt.Println("Type of Operation State: sequential")

		dataState := r.lastOutput //assuming for any operation state: first action gets input from inputFile.json
		for _, fr := range functionRefs {
			apiCall, _ := r.funcToEndpoint[fr]
			fmt.Println("making this apiCall: ", apiCall)

			bodyText, err := functionInvoker(apiCall, dataState, state, client, 0)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s\n", bodyText)
			dataState = bodyText //last action's output is current action's input
		}
		if state.GetTransition() != nil {
			ns := r.nameToState[state.GetTransition().NextState]
			r.begin((ns))
		} else {
			fmt.Println("This is the end..")
		}
		return nil
	case "parallel":
		//TODO: parallel invoked actions need to be synchronized(?)..
		parallelization := len(functionRefs)
		fmt.Println("Type of Operation State: parallel")
		fmt.Println("Numbers of parallel actions here: ", parallelization)
		dataState := r.lastOutput
		channel := make(chan string)

		var wg sync.WaitGroup
		counter := 0
		wg.Add(parallelization)

		for ii := 0; ii < parallelization; ii++ {
			go func(channel chan string, thisCounter int) {
				for {
					fr, more := <-channel
					if more == false {
						wg.Done()
						return
					}

					apiCall, _ := r.funcToEndpoint[fr]
					bodyText, err := functionInvoker(apiCall, dataState, state, client, thisCounter)

					if err != nil {
						log.Printf("nop")
					}
					fmt.Printf("%s\n", bodyText)
				}
			}(channel, counter)
			counter++
			//time.Sleep(100 * time.Millisecond)
		}
		//time.Sleep(5 * time.Second)
		for _, fr := range functionRefs {
			channel <- fr
		}
		close(channel)
		wg.Wait()

		if state.GetTransition() != nil {
			ns := r.nameToState[state.GetTransition().NextState]
			r.begin(ns)
		} else {
			fmt.Println("This is the end..")
		}

		return nil
	}
	return nil
}

func handleSequentialActions(st *model.OperationState) []string {
	var refs []string
	for _, action := range st.Actions {
		fName := action.FunctionRef.RefName
		// TODO
		// May we assume that there will be 1 action per sequential operation state?
		//fmt.Println(fName)
		refs = append(refs, fName)
	}
	return refs
}

func functionInvoker(apiCall string, dataState []uint8, state *model.OperationState, client *http.Client, i int) ([]uint8, error) {
	var data map[string]interface{} //data = input file
	json.Unmarshal(dataState, &data)

	//fmt.Println("state.Actions = ", state.Actions, i)
	var parsings []string //finding arguments of func
	args := state.Actions[i].FunctionRef.Arguments

	for _, value := range args {
		parsings = append(parsings, value.(string))
	}
	finalParsings := strings.Join(parsings, ", .")
	if (len(parsings) > 1) {
		finalParsings = "." + finalParsings
	}

	op, _ := gojq.Parse(finalParsings)
	iter := op.Run(data) //filtering the data for the function invocation

	//iterating through args to fill the POST request data
	for key, _ := range args {
		val, _ := iter.Next()
		data[key] = val
	}

	jsonData, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", apiCall, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth("23bc46b1-71f6-4ed5-8c54-816aa4f8c502", "123zO3xZCLrMN6v2BKK1dXYFpXlPkccOFqm12CdAsMgRU4VrNZ9lyGVCGuMDGIwP")
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return bodyText, err
}

func handleDataBasedSwitch(state *model.DataBasedSwitchState, r *Runtime) error {
	fmt.Println("--> DataBasedSwitch: ", state.GetName())
	for _, cond := range state.DataConditions {
		fmt.Println(cond.GetCondition())
		switch cond.(type) {
		case *model.TransitionDataCondition:
			var result map[string]interface{}
			json.Unmarshal(r.lastOutput, &result)
			op, _ := gojq.Parse(cond.GetCondition())
			iter := op.Run(result)
			v, _ := iter.Next()
			if err, ok := v.(error); ok {
				log.Fatalln(err)
			}

			if v.(bool) {
				fmt.Println("True")
				fmt.Println("GOTO", cond.(*model.TransitionDataCondition).Transition.NextState)
				ns := cond.(*model.TransitionDataCondition).Transition.NextState
				r.begin(r.nameToState[ns])
				return nil
			} else {
				fmt.Println("Not True")
				continue
				return nil
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
			fmt.Println("This is the end...")
		}

	}
	return nil
}

func handleInjectState(state *model.InjectState, r *Runtime) error {
	fmt.Println("--> Inject: ", state.GetName())
	//injectFilter := state.GetStateDataFilter()
	injectData := state.Data
	fmt.Println("Data of inject state: ", injectData)
	//fmt.Println("Input filter: ", injectFilter.Input, " Output filter: ", injectFilter.Output)
	//outFilter := strings.Split(injectFilter.Output, " ")[1]
	//outFilter = strings.Split(outFilter, ".")[1]
	if state.GetTransition() != nil {
		ns := state.Transition.NextState
		r.begin(r.nameToState[ns])
		return nil
	} else {
		fmt.Println("This is the end..")
		return nil
	}
}
