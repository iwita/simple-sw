package runtime

import (
	"encoding/json"
	"fmt"
	"log"
	"context"
	//"io/ioutil"
	"net/http"
	"strings"
	//"bytes"
	//"crypto/tls"
	"sync"
	//"os/exec"
	//"runtime"
	//"time"

	"github.com/apache/openwhisk-client-go/whisk"
	"github.com/itchyny/gojq"
	"github.com/serverlessworkflow/sdk-go/model"
)

func handleEventState(state *model.EventState, r *Runtime) error {
	fmt.Println("--> Event:", state.GetName())
	switch state.Exclusive {
	case false:
		fmt.Println("exclusive: False")
	default:
		fmt.Println("exclusive: True")
	}
	if state.GetTransition() != nil {
		ns := state.Transition.NextState
		r.begin(r.nameToState[ns])
		return nil
	} else {
		fmt.Println("This is the end..")
		redisEraser(r)
		return nil
	}
}

func handleOperationState(state *model.OperationState, r *Runtime) error {
	fmt.Println("--> Operation:", state.GetName())
	functionRefs := handleSequentialActions(state) //getting the funcRefs of this op.state

	client, _ := whisk.NewClient(http.DefaultClient, nil)

	// Check for the action Mode (default: sequential)
	switch state.ActionMode {
	case "sequential": //assuming 1 operation state => multiple dependable actions or just one independent
		fmt.Println("Type of Operation State: sequential")

		dataState := r.lastOutput //assuming for any operation state: first action gets input from inputFile.json
		for i, fr := range functionRefs {
			apiCall, _ := r.funcToEndpoint[fr]
			//fmt.Println("making this apiCall: ", apiCall)

			form, result, err := functionInvoker(apiCall, dataState, state, client, i)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Result: %s\n", result)
			dataState = form //last action's output is current action's input
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
		channel := make(chan string) //channel for funcRefs
		channel2 := make(chan int) //channel for enumerating the funcRefs

		var wg sync.WaitGroup
		wg.Add(parallelization)

		for ii := 0; ii < parallelization; ii++ {
			go func(channel chan string, channel2 chan int) {
				for {
					fr, more := <-channel
					num, _ := <-channel2
					if more == false {
						wg.Done()
						return
					}

					apiCall, _ := r.funcToEndpoint[fr]
					_, result, err := functionInvoker(apiCall, dataState, state, client, num)

					if err != nil {
						log.Printf("nop")
					}
					fmt.Printf("Result: %s\n", result)
				}
			}(channel, channel2)
		}

		for num, fr := range functionRefs {
			channel <- fr
			channel2 <- num
		}

		close(channel)
		close(channel2)
		wg.Wait()

		if state.GetTransition() != nil {
			ns := r.nameToState[state.GetTransition().NextState]
			r.begin(ns)
		} else {
			fmt.Println("This is the end..")
			redisEraser(r)
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

func functionInvoker(apiCall string, dataState []uint8, state *model.OperationState, client *whisk.Client, i int) ([]uint8, string, error) {
	//for OperationState arguments have to be in form: ".field1.field2.."
	var data map[string]interface{} //data = input file
	data2 := make(map[string]interface{})
	json.Unmarshal(dataState, &data)
	var parsings []string //finding arguments of func
	args := state.Actions[i].FunctionRef.Arguments

	for _, value := range args {
		parsings = append(parsings, value.(string))
	}

	finalParsings := strings.Join(parsings, ", ")

	query, _ := gojq.Parse(finalParsings)

	iter := query.Run(data) //filtering the data for the function invocation

	//iterating through args to fill the POST request data
	for key, _ := range args {
		val, _ := iter.Next()
		data2[key] = val
	}


	//jsonData, _ := json.Marshal(data2)
	//result, _, err := client.Actions.Invoke(apiCall, bytes.NewBuffer(jsonData), true, true)
	result, _, err := client.Actions.Invoke(apiCall, data2, true, true)

	if err != nil {
		log.Fatal(err)
	}

	var bodyText string
	for _, v := range result {
		bodyText = v.(string)
	}

	form, _ := json.Marshal(result)
	//fmt.Println("result = ", result, form)
	//fmt.Printf("type is %T\n", form)

	return form, bodyText, err
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
			redisEraser(r)
			return nil
		}

	}
	return nil
}

func handleForEachState(state *model.ForEachState, r *Runtime) error {
	fmt.Println("--> ForEach: ", state.GetName())

	client, _ := whisk.NewClient(http.DefaultClient, nil)
	functionRefs := handleForEachActions(state)

	var data map[string]interface{}
	json.Unmarshal(r.lastOutput, &data)

	//getting the filtered Data based on InputCollection filters
	in := strings.Split(state.InputCollection, "${ [")[1]
	in = strings.Split(in, "] }")[0]

	query, _ := gojq.Parse(in)
	iter := query.Run(data)

	val, ok := iter.Next()
	var inputCollection []interface{}
	for (ok != false){
		inputCollection = append(inputCollection, val)
		val, ok = iter.Next()
	}

	jsonData, _ := json.Marshal(inputCollection)



	parallelization := len(jsonData) //all elements of inputCollection array need to be executed in parallel
	channel, channel2 := make(chan int), make(chan string)

	var wg sync.WaitGroup
	wg.Add(parallelization)
	outputCollection := make(map[string][]string)

	//for every apiCall: execute in parallel for all elements in InputCollection
	//assuming that not only apiCalls but also actions are executed in parallel
	for ii := 0; ii < parallelization; ii++ {
		go func(channel chan int, channel2 chan string) {
			for {
				fr, more2 := <-channel2
				num, _ := <-channel
				if more2 == false {
					wg.Done()
					return
				}

				apiCall, _ := r.funcToEndpoint[fr]
				results, err := functionInvoker2(apiCall, inputCollection, client, state, num)
				if err != nil {
					log.Fatal(err)
				}
				//fmt.Println("for apiCall: ", apiCall)
				var array []string
				for _, result := range results {
					array = append(array, string(result))
					//fmt.Println(string(result))
				}
				outputCollection[apiCall + string(num)] = array
			}
		}(channel, channel2)
	}

	//sending to the channel all apiCalls that need to be executed
	for num, fr := range functionRefs {
		channel2 <- fr
		channel <- num
	}
	close(channel)
	close(channel2)
	wg.Wait()

	//printing the concentrated results
	for key, values := range outputCollection {
		fmt.Println("apiCall: ", key)
		for _, value := range values {
			fmt.Println(value)
		}
	}

	//prepei na filtrarw to dataState vasi tou inputCollection
	//gia kathe element pou epilgetai apo to inputCollection, px to acceptedApplicant,
	//to opoio tha perasei stis parallel executed actions
	//ta apotelesmata kathe parallel executed action tha apothikeutoun sto outputCollection

	if state.GetTransition() != nil {
		ns := state.Transition.NextState
		r.begin(r.nameToState[ns])
		return nil
	} else {
		fmt.Println("This is the end..")
		redisEraser(r)
		return nil
	}
}

func functionInvoker2(apiCall string, inputCollection []interface{}, client *whisk.Client, state *model.ForEachState, i int) ([][]byte, error) {

	//for ForEachState arguments have to be in form: "${ .field1.field2.. }"
	parallelization := len(inputCollection)
	iterationParam := state.IterationParam
	channel3 := make(chan interface{})
	var results [][]byte

	var wg sync.WaitGroup
	wg.Add(parallelization)

	args := state.Actions[i].FunctionRef.Arguments //find arguments of the appropriate apiCall
	var parsings []string

	//building the gojq.Parse()
	for _, value := range args {
		value2 := value.(string)
		value3 := strings.Split(value2, "${ .")[1]
		value3 = strings.Split(value3, " }")[0]
		value3 = strings.Split(value3, iterationParam)[1]
		parsings = append(parsings, value3)
	}

	finalParsings := strings.Join(parsings, ", ")
	query, _ := gojq.Parse(finalParsings)
	var input []interface{} //input = filtered data to be sent to apiCall

	for _, obj := range inputCollection {
		//fmt.Println("obj = ", obj)
		data2 := make(map[string]interface{})
		iter := query.Run(obj)
		for key, _ := range args {
			val, _ := iter.Next()
			data2[key] = val
		}
		input = append(input, data2)
	}

	for ii := 0; ii < parallelization; ii++ {
		go func(channel3 chan interface{}) {
			for {
				element, more := <-channel3
				if more == false {
					wg.Done()
					return
				}

				/*
				jsonElement, _ := json.Marshal(element)
				req, err := http.NewRequest("POST", apiCall, bytes.NewBuffer(jsonElement))
				req.Header.Set("Content-Type", "application/json; charset=UTF-8")
				if err != nil {
					log.Fatal(err)
				}

				req.SetBasicAuth("23bc46b1-71f6-4ed5-8c54-816aa4f8c502", "123zO3xZCLrMN6v2BKK1dXYFpXlPkccOFqm12CdAsMgRU4VrNZ9lyGVCGuMDGIwP")
				resp, err := client.Do(req) //making the API request
				*/
				result, _, err := client.Actions.Invoke(apiCall, element, true, true)
				if err != nil {
					log.Fatal(err)
				}

				var bodyText string
				for _, v := range result {
					bodyText  = v.(string)
				}

				/*bodyText, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err)
				}*/

				results = append(results, []byte(bodyText)) //results of one apiCall
			}
		}(channel3)
	}
	for _, element := range input {
		channel3 <- element
	}
	close(channel3)
	wg.Wait()
	return results, nil
}

func handleForEachActions(st *model.ForEachState) []string {
        var refs []string
        for _, action := range st.Actions {
                fName := action.FunctionRef.RefName
                refs = append(refs, fName)
        }
        return refs
}


func handleInjectState(state *model.InjectState, r *Runtime) error {
	fmt.Println("--> Inject: ", state.GetName())
	injectFilter := state.GetStateDataFilter()

	var ctx = context.Background()
	incomingData, _ := r.Red.HGet(ctx, "channel", state.GetName()).Bytes()
	fmt.Println("incomingData = ", string(incomingData))
	//fmt.Println("injectFilter: ", injectFilter)

	var processingData []uint8
	var outgoingData []uint8

	if (injectFilter != nil) && (injectFilter.Input != "") {
		processingData = stateDataFilter(injectFilter.Input, incomingData)
	} else {
		processingData = incomingData
	}

	// ----------------LOGIC OF STATE--------------------
	// there is no logic in an inject state, so processingData = outgoingData
	// ---------------/LOGIC OF STATE--------------------

	if injectFilter != nil && injectFilter.Output != "" {
		outgoingData = stateDataFilter(injectFilter.Output, processingData)
	} else {
		outgoingData = processingData
	}

	if state.GetTransition() != nil {
		ns := state.Transition.NextState
		_ = r.Red.HSet(ctx, "channel", ns, outgoingData)
		r.begin(r.nameToState[ns])
		return nil
	} else {
		fmt.Println("This is the end..")
		redisEraser(r)
		return nil
	}
}

//e.g for filter: "${ {applicants: [.applicants[] | select(.age >= 18)]} }"
//gets filter and data and produces filtered data for the state
func stateDataFilter(filter string, incomingData []uint8) ([]uint8) {
	var data map[string]interface{}
	var result []uint8
	json.Unmarshal(incomingData, &data)

	filter = strings.Split(filter, "${ ")[1]
	filter = strings.Split(filter, " }")[0]
	query, _ := gojq.Parse(filter)
	iter := query.Run(data)
	val, _ := iter.Next()

	result, _ = json.Marshal(val)
	return result
}

func redisEraser(r *Runtime) {
        var ctx = context.Background()
        for _, state := range r.Workflow.States {
                _ = r.Red.HDel(ctx, "channel", state.GetName())
        }

}
