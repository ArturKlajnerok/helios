package main

import (
	"fmt"
	"net/http"
	"time"
)

var iteration = 0

func performRequest(printTime bool, startedIteration int) {
	before := time.Now()
	//resp, err := http.Get("http://example.com/")
	resp, err := http.Get("http://dev-auth-s1:8080/token?grant_type=password&client_id=1234&client_secret=aabbccdd&username=test&password=test")
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	after := time.Now()
	dur := after.Sub(before)
	if printTime {
		fmt.Printf("First iteration request took: %f s\n", dur.Seconds())
	}
	if iteration > startedIteration {
		fmt.Printf("Request fell into next iteration - took: %f s\n", dur.Seconds())
	}
}

func main() {
	fmt.Println("Starting stress test")

	throughtput := 20

	for {
		before := time.Now()
		iteration++
		iter := iteration
		for i := 0; i < throughtput; i++ {
			go performRequest(i == 0, iter)
		}
		after := time.Now()
		dur := after.Sub(before)
		sleepTime := 1000 - (dur.Nanoseconds() / 1000000)
		if sleepTime > 0 {
			time.Sleep(time.Duration(sleepTime * 1000000))
		}
	}
	fmt.Println("Test finished successfully")
}
