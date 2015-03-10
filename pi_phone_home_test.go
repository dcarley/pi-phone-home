package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
)

func testRequestTimings(server *ghttp.Server, requests chan time.Time, delay time.Duration) {
	expectedRequests := 3

	Expect(server.ReceivedRequests()).Should(HaveLen(expectedRequests))
	Expect(requests).Should(HaveLen(expectedRequests))
	close(requests)

	var lastReq time.Time
	Expect(requests).Should(Receive(&lastReq))

	for req := range requests {
		Expect(req).To(BeTemporally("~", lastReq.Add(delay), delay/10))
		lastReq = req
	}
}

var _ = Describe("PiPhoneHome", func() {
	var (
		logbuf *gbytes.Buffer

		lookupServer net.Listener
		lookupAddr   string

		server   *ghttp.Server
		phoneURL *url.URL
		reqPath  = "/pi-phone-home/test"

		defaultDuration = time.Millisecond

		requests      chan time.Time
		recordRequest = func(w http.ResponseWriter, r *http.Request) {
			requests <- time.Now()
		}

		quit     chan struct{}
		sendQuit = func(w http.ResponseWriter, r *http.Request) {
			close(quit)
		}
	)

	BeforeEach(func() {
		logbuf = gbytes.NewBuffer()
		logger = log.New(logbuf, "", 0)

		var err error
		lookupServer, err = net.Listen("tcp", "0.0.0.0:0")
		Expect(err).Should(BeNil())
		lookupAddr = lookupServer.Addr().String()

		quit = make(chan struct{})
		requests = make(chan time.Time, 10)

		server = ghttp.NewServer()
		phoneURL, _ = url.Parse(server.URL())
		phoneURL.Path = reqPath
	})

	AfterEach(func() {
		lookupServer.Close()
		server.Close()
	})

	Describe("successful requests", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					http.HandlerFunc(recordRequest),
					ghttp.VerifyRequest("HEAD", reqPath),
				),
				ghttp.CombineHandlers(
					http.HandlerFunc(recordRequest),
					ghttp.VerifyRequest("HEAD", reqPath),
				),
				ghttp.CombineHandlers(
					http.HandlerFunc(sendQuit),
					http.HandlerFunc(recordRequest),
					ghttp.VerifyRequest("HEAD", reqPath),
				),
			)
		})

		It("should wait interval period between successful requests", func() {
			customDuration := 100 * time.Millisecond

			phoneForever(
				phoneURL,
				defaultDuration,
				defaultDuration,
				customDuration,
				lookupAddr,
				quit,
			)

			testRequestTimings(server, requests, customDuration)

			for i := 1; i < 4; i++ {
				Expect(logbuf).To(gbytes.Say("Success: phoned home"))
				Expect(logbuf).To(gbytes.Say(fmt.Sprintf("Sleeping for: %s", customDuration)))
			}
		})
	})

	Describe("failed requests", func() {
		var customDuration = 100 * time.Millisecond

		BeforeEach(func() {
			delayResponse := func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(customDuration * 2)
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					http.HandlerFunc(recordRequest),
					http.HandlerFunc(delayResponse),
					ghttp.VerifyRequest("HEAD", reqPath),
				),
				ghttp.CombineHandlers(
					http.HandlerFunc(recordRequest),
					http.HandlerFunc(delayResponse),
					ghttp.VerifyRequest("HEAD", reqPath),
				),
				ghttp.CombineHandlers(
					http.HandlerFunc(sendQuit),
					http.HandlerFunc(recordRequest),
					http.HandlerFunc(delayResponse),
					ghttp.VerifyRequest("HEAD", reqPath),
				),
			)
		})

		It("should retry requests if they have exceeded the timeout period", func() {
			phoneForever(
				phoneURL,
				customDuration,
				defaultDuration,
				defaultDuration,
				lookupAddr,
				quit,
			)

			testRequestTimings(server, requests, customDuration)

			for i := 1; i < 4; i++ {
				Expect(logbuf).To(gbytes.Say("Error: Head .+ use of closed network connection"))
				Expect(logbuf).To(gbytes.Say(fmt.Sprintf("Sleeping for: %s", defaultDuration)))
			}
		})

		It("should wait retry period between failed requests", func() {
			phoneForever(
				phoneURL,
				defaultDuration,
				customDuration,
				defaultDuration,
				lookupAddr,
				quit,
			)

			testRequestTimings(server, requests, customDuration)

			for i := 1; i < 4; i++ {
				Expect(logbuf).To(gbytes.Say("Error: Head .+ use of closed network connection"))
				Expect(logbuf).To(gbytes.Say(fmt.Sprintf("Sleeping for: %s", customDuration)))
			}
		})
	})

	Describe("reporting primary IP address", func() {
		var lookupPort string

		Describe("connecting from IPv4 loopback", func() {
			BeforeEach(func() {
				_, lookupPort, _ = net.SplitHostPort(lookupAddr)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						http.HandlerFunc(sendQuit),
						ghttp.VerifyRequest("HEAD", reqPath, "local=127.0.0.1"),
					),
				)
			})

			It("should send query param with IPv4 loopback address", func() {
				phoneForever(
					phoneURL,
					defaultDuration,
					defaultDuration,
					defaultDuration,
					net.JoinHostPort("127.0.0.1", lookupPort),
					quit,
				)
			})
		})

		Describe("connecting from IPv6 loopback", func() {
			BeforeEach(func() {
				_, lookupPort, _ = net.SplitHostPort(lookupAddr)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						http.HandlerFunc(sendQuit),
						ghttp.VerifyRequest("HEAD", reqPath, "local=%3A%3A1"),
					),
				)
			})

			It("should send query param with IPv6 loopback address", func() {
				phoneForever(
					phoneURL,
					defaultDuration,
					defaultDuration,
					defaultDuration,
					net.JoinHostPort("::1", lookupPort),
					quit,
				)
			})
		})
	})
})
