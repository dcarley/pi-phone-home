package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

var (
	reqTimeout  = flag.Duration("timeout", 5 * time.Second, "Request timeout")
	reqInterval = flag.Duration("interval", 10 * time.Second, "Request interval")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s <URL>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	phone(flag.Args()[0], *reqTimeout, *reqInterval)
}

func phone(homeURL string, timeout, interval time.Duration) {
	client := &http.Client{
		Timeout: timeout,
	}

	for {
		_, err := client.Head(homeURL)
		if err == nil {
			break
		}

		fmt.Println(err)
		fmt.Printf("Sleeping for %s\n", interval)
		time.Sleep(interval)
	}

	fmt.Println("Successfully phoned home")
}
