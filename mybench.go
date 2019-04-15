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
}

type summaryInfo struct {
	requested int64
	responded int64
}

type ResponseBody struct {
	ServerHostname string `json:"server_hostname"`
}

func main() {
	fmt.Println("Golang course pre-work")

	requests := flag.Int64("n", 1, "Number of requests to perform")
	concurrency := flag.Int64("c", 1, "Number of multiple requests to make at a time")
	timeout := flag.Int64("s", 1, "Seconds to max. wait for each response. Default is 30 seconds")
	timelimit := flag.Int64("t", 1, "Seconds to max. to spend on benchmarking. This implies -n 50000")

	fmt.Println(requests, concurrency, timeout, timelimit)

	flag.Parse()
	//flag.PrintDefaults()

	if flag.NArg() == 0 || *requests == 0 || *requests < *concurrency {
		flag.PrintDefaults()
		os.Exit(-1)
	}

	link := flag.Arg(0)

	c := make(chan responseInfo)
	summary := summaryInfo{}
	for i := int64(0); i < *concurrency; i++ {
		summary.requested++
		go checkLink(link, c)
	}

	for response := range c {
		if summary.requested < *requests {
			summary.requested++
			go checkLink(link, c)
		}
		summary.responded++
		fmt.Println(response)
		if summary.requested == summary.responded {
			break
		}
	}
}

func checkLink(link string, c chan responseInfo) {
	start := time.Now()
	res, err := http.Get(link)
	if err != nil {
		panic(err)
	}

	read, _ := io.Copy(ioutil.Discard, res.Body)

	if err != nil {
		fmt.Println(err.Error())
	}

	c <- responseInfo{
		status: res.StatusCode,
		bytes: read,
		duration: time.Now().Sub(start),
	}
}