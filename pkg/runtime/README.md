# **Data transfer logic added**


## Things, so far, work this way:

A state gets its data by reading from Redis channel the value that belongs to the key "state.GetName()"

In this way, all states "know" apriori where to find their data..


### Operation state: 
 
 #### - actionMode: sequential
 
 Current state's output is the next state's input, so we have to be strict with arguments' "tuning" in the workflow.yaml 

 #### - actionMode: parallel
 
 Current state's output is NOT transfered to the next state
 
 The next state gets the same data as the current state got
 
### Event state:
 
 not ready yet

### ForEach state:

 Current state's output is NOT transfered to the next state
 
 The next state gets the same data as the current state got
 
### DataBasedSwitch state:

 Current state's output is NOT transfered to the next state
 
 The next state gets the same data as the current state got

### Inject state:

 Current state's output is the next state's input, so we have to be strict with arguments' "tuning" in the workflow.yaml 
 
 StateData filters also implemented here
