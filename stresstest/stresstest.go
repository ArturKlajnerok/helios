package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Request struct {
	Number int
	Url    string
}

type RequestResult struct {
	Number int
	Time   float64
}

func performRequests(chInput chan Request, chTimes chan RequestResult, wg *sync.WaitGroup) {

	for {
		request, more := <-chInput
		if more {
			before := time.Now()
			resp, err := http.Get(request.Url)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
			}
			if resp != nil {
				defer resp.Body.Close()
			}
			after := time.Now()
			dur := after.Sub(before)
			result := RequestResult{request.Number, float64(dur.Nanoseconds()) / 1000000}
			chTimes <- result
		} else {
			wg.Done()
			break
		}
	}
}

func calculateTimes(chTimes chan RequestResult, chResultTime chan float64) {
	average := float64(0)
	valueCount := 0
	for {
		reqResult, more := <-chTimes
		if more {
			valueCount++
			if average == 0 {
				average = reqResult.Time
			} else {
				average = average + (reqResult.Time-average)/float64(valueCount)
			}
			if reqResult.Number > 0 && reqResult.Number%1000 == 0 {
				fmt.Printf("Request number %d - average time %.3f ms\n", reqResult.Number, reqResult.Time)
			}
		} else {
			chResultTime <- average
			break
		}
	}
}

func main() {
	fmt.Println("Starting stress test")

	requestCount := 10
	threadCount := 1
	var wg sync.WaitGroup
	wg.Add(threadCount)

	chInput := make(chan Request, requestCount)
	chTimes := make(chan RequestResult, requestCount)
	chResultTime := make(chan float64, 1)

	for i := 0; i < requestCount; i++ {
		request := Request{i, "http://dev-auth-s1:8080/token?grant_type=password&client_id=123456&client_secret=aabbccdd&username=test&password=test"}
		chInput <- request
	}
	close(chInput)

	for i := 0; i < threadCount; i++ {
		go performRequests(chInput, chTimes, &wg)
	}

	go calculateTimes(chTimes, chResultTime)
	wg.Wait()
	close(chTimes)

	fmt.Printf("Average time is %f\n", <-chResultTime)
	fmt.Println("Test finished successfully")
}
