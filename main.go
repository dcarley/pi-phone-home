package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

var logger = log.New(os.Stdout, "", 0)

func main() {
	var (
		timeout    = flag.Duration("timeout", 5*time.Second, "Timeout for individual request")
		retry      = flag.Duration("retry", 10*time.Second, "Delay between failed requests")
		interval   = flag.Duration("interval", 6*time.Hour, "Delay between successful requests")
		lookupAddr = flag.String("lookupAddr", "google.com:80", "Public host:port for IP lookup")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s <URL>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	parsedURL, err := url.Parse(flag.Args()[0])
	if err != nil {
		logger.Fatalln(err)
	}

	quit := make(chan struct{})
	phoneForever(parsedURL, *timeout, *retry, *interval, *lookupAddr, quit)
}

func phoneForever(
	homeURL *url.URL,
	timeout, retry, interval time.Duration,
	lookupAddr string,
	quit chan struct{},
) {
	client := &http.Client{
		Timeout: timeout,
	}

	var delay time.Duration
	for {
		select {
		case <-quit:
			return
		case <-time.After(delay):
			break
		}

		err := phoneOnce(client, homeURL, lookupAddr)
		if err != nil {
			logger.Println("Error:", err)
			delay = retry
		} else {
			logger.Println("Success: phoned home")
			delay = interval
		}

		logger.Println("Sleeping for:", delay)
	}
}

func phoneOnce(client *http.Client, homeURL *url.URL, lookupAddr string) error {
	ip, err := findPrimaryIP(lookupAddr)
	if err != nil {
		return err
	}

	query := homeURL.Query()
	query.Set("local", ip)
	homeURL.RawQuery = query.Encode()

	_, err = client.Head(homeURL.String())
	return err
}

func findPrimaryIP(addr string) (string, error) {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	host, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return "", err
	}

	return host, nil
}
