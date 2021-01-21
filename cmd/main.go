package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/ChristopherHolder/complexlb/loadbalancer"
	log "github.com/sirupsen/logrus"
)

func main() {
	var serverList string
	var algoType string
	var port int

	flag.StringVar(&serverList, "servers", "", "Load balanced servers, use commas to separate")
	flag.StringVar(&algoType, "algo", "cycle", "Algorithm to distribute requests: naive | wrr | wrri")
	flag.IntVar(&port, "port", 3030, "Port to serve")

	flag.Parse()

	if len(serverList) == 0 {
		log.Fatal("Please provide one or more servers to load balance")
	}

	serverManager, err := loadbalancer.NewServerManager(algoType)

	if err != nil {
		log.Fatal(err.Error())
	}
	// parse servers
	tokens := strings.Split(serverList, ",")
	for index, tok := range tokens {
		serverURL, err := url.Parse(tok)
		UID := uint32(index)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatal()
		}

		proxy := httputil.NewSingleHostReverseProxy(serverURL)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.WithFields(log.Fields{
				"host":  serverURL.Host,
				"error": e.Error(),
			}).Error()
			retries := loadbalancer.GetRetryFromContext(request)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), loadbalancer.Retry, retries+1)
					proxy.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}

			// after 3 retries, mark this server as down
			serverManager.MarkServerStatus(UID, false)

			// if the same request routing for few attempts with different servers, increase the count
			attempts := loadbalancer.GetAttemptsFromContext(request)
			log.WithFields(log.Fields{
				"address":  request.RemoteAddr,
				"path":     request.URL.Path,
				"attempts": attempts,
			}).Warning("Attempting retry")
			ctx := context.WithValue(request.Context(), loadbalancer.Attempts, attempts+1)
			loadbalancer.Handle(writer, request.WithContext(ctx), serverManager)
		}

		serverManager.AddServer(&loadbalancer.Server{
			UID:          UID,
			URL:          serverURL,
			Alive:        true,
			ReverseProxy: proxy,
		})
		log.WithFields(log.Fields{
			"url": serverURL,
		}).Info("Configured server")
	}

	// create http server
	httpServer := http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			loadbalancer.Handle(rw, r, serverManager)
		}),
	}

	// start health checking
	go loadbalancer.HealthCheckRun(serverManager)

	log.WithFields(log.Fields{
		"port": port,
	}).Info("Load Balancer started")
	if err := httpServer.ListenAndServe(); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal()
	}
}
