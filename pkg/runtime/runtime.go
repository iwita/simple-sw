package runtime

import (
	"fmt"
	"io/ioutil"
	"os"
	"context"
	"net/http"

	"github.com/serverlessworkflow/sdk-go/model"
	"github.com/go-redis/redis/v8"
	"github.com/apache/openwhisk-client-go/whisk"
)

func (r *Runtime) Start() {
	r.exec()
}

func (r *Runtime) exec() {
	//----------------------------------
	//initializing redis client
	var ctx = context.Background()
	r.Red = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "", // no password set
		DB: 0, // use default DB
	})

	//initializing Openwhisk client
	r.Whisk, _ = whisk.NewClient(http.DefaultClient, nil)

	//_ = r.Red.HSet(ctx, "channel", [])
	//_ = rdb.HSet(ctx, "channel", "fruit", "strawberry")
	//_ = rdb.HSet(ctx, "channel", "fruit", "banana")
	//vals := rdb.HGetAll(ctx, "channel")
	//fmt.Println(vals)
	//for _, name := range r.Workflow.States {
		//fmt.Println("state name: ", name.GetName())
	//	_ = r.Red.HSet(ctx, "channel", name.GetName(), "")
	//}
	//fmt.Println(r.Red.HGetAll(ctx, "channel"))

	// ----------------------------------
        //initState := r.Workflow.States[0]
        //maybe States[0] is not the starting State..
        initStateName := r.Workflow.Start.StateName
	fmt.Println("Deploying workflow: ", r.Workflow.Name)
        fmt.Println("Name of the input file is: ", r.InputFile)
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

        initState := r.nameToState[initStateName]
	//putting input data to FirstState's channel..
	_ = r.Red.HSet(ctx, "channel", initStateName, r.lastOutput)
        r.begin(initState)

}

func (r *Runtime) begin(st model.State) error {
	switch st.(type) {
	case *model.EventState:
		//fmt.Println("event")
		handleEventState(st.(*model.EventState), r)
	case *model.OperationState:
		// fmt.Println("operation")
		err := handleOperationState(st.(*model.OperationState), r)
		if err != nil {
			fmt.Println("Error in Operation State")
			return err
		}
	case *model.EventBasedSwitchState:
		fmt.Println("event based switch")
	case *model.DataBasedSwitchState:
		err := handleDataBasedSwitch(st.(*model.DataBasedSwitchState), r)
		if err != nil {
			fmt.Println("Error in DataBasedSwitchState")
		}
	case *model.ForEachState:
		//fmt.Println("foreach")
		err := handleForEachState(st.(*model.ForEachState), r)
		if err != nil {
			fmt.Println("Error in ForEach State")
			return err
		}
	case *model.ParallelState:
		fmt.Println("parallel")
	case *model.InjectState:
		err := handleInjectState(st.(*model.InjectState), r)
		if err != nil {
			fmt.Println("Error in Inject State")
			return err
		}
	}
	return nil
}
