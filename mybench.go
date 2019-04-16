package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type responseInfo struct {
	status int
	bytes int64
	duration time.Duration
	ServerHostname string
	TimePerRequest time.Duration
	CompleteRequest int64
	FailedRequest int64
}

type summaryInfo struct {
	requested int64
	responded int64
	ServerHostname string
	TimePerRequest time.Duration
	CompleteRequest int64
	FailedRequest int64
}

func main() {
	fmt.Println("Golang course pre-work")

	requests := flag.Int64("n", 1, "Number of requests to perform")
	concurrency := flag.Int64("c", 1, "Number of multiple requests to make at a time")
	s := flag.Int64("s", 30, "Seconds to max. wait for each response. Default is 30 seconds")
	t := flag.Int64("t", 30, "Seconds to max. to spend on benchmarking. This implies -n 50000")

	timeout := time.Duration(*s) * time.Millisecond
	timelimit := time.Duration(*t) * time.Millisecond

	fmt.Println(requests, concurrency, timeout, timelimit)

	flag.Parse()
	//flag.PrintDefaults()

	if flag.NArg() == 0 || *requests == 0 || *requests < *concurrency {
		flag.PrintDefaults()
		os.Exit(-1)
	}

	var timePerRequest []time.Duration
	link := flag.Arg(0)

	c := make(chan responseInfo)
	summary := summaryInfo{}
	for i := int64(0); i < *concurrency; i++ {
		summary.requested++
		go checkLink(link, timeout, timelimit, c)
	}

	for response := range c {
		if summary.requested < *requests {
			summary.requested++
			go checkLink(link, timeout, timelimit, c)
		}
		summary.responded++
		summary.CompleteRequest++
		fmt.Println(response)
		summary.ServerHostname = response.ServerHostname
		timePerRequest = append(timePerRequest, response.TimePerRequest)
		if summary.requested == summary.responded {
			break
		}
	}

	var timePerRequestAverage float64
	for time := range timePerRequest {
		timePerRequestAverage += float64(time)
	}
	timePerRequestAverage = timePerRequestAverage / float64(len(timePerRequest))

	summary.FailedRequest = *requests - summary.CompleteRequest
	fmt.Println("--- summary ---")
	fmt.Println("+ Server Hostname: ", summary.ServerHostname)
	fmt.Println("+ Time Per Request: ", timePerRequestAverage)
	fmt.Println("+ Complete Request: ", summary.CompleteRequest)
	fmt.Println("+ Failed Request: ", summary.FailedRequest)

}

func checkLink(link string, timeout time.Duration, timelimit time.Duration, c chan responseInfo) {
	start := time.Now()
	//res, err := http.Get(link)

	client := http.Client{}
	request, _ := http.NewRequest("GET", link, nil)
	resp, err := client.Do(request)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	read, _ := io.Copy(ioutil.Discard, resp.Body)

	if err != nil {
		fmt.Println(err.Error())
	}

	c <- responseInfo{
		status: resp.StatusCode,
		bytes: read,
		duration: time.Now().Sub(start),
		ServerHostname: request.Host,
		TimePerRequest: time.Now().Sub(start),
	}
}