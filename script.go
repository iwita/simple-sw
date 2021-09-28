package main

import (
	"fmt"
	"os/exec"
	"time"
)

//script for getting more accurate results among the workflows
func main() {

	start := time.Now()
	for i := 0; i < 10; i++ {
		_, err := exec.Command("go", "run", "main.go").Output()

		if (err != nil) {
			fmt.Printf("%s", err)
		}
	}
        fmt.Println("Command Successfully Executed")
	elapsed := time.Since(start)
	fmt.Printf("Time: %s\n", elapsed)
}

