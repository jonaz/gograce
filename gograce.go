package gograce

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var testQuit chan os.Signal

// NewServerWithTimeout takes 1 or 2 arguments. The first one is how long the webservers will wait to process current requests.
// The second is how long it will wait after getting signal before telling the webserver to start shutdown. This is
// useful for when running in kubernetes and instead of adding a preStop sleep to the pod in order to let the ingress controller
// update it's backends before the service stops accepting new requests.
func NewServerWithTimeout(t ...time.Duration) (*http.Server, chan struct{}) {
	if len(t) == 0 {
		t = []time.Duration{10 * time.Second}
	}

	shutdown := make(chan struct{})
	srv := &http.Server{}

	quit := make(chan os.Signal)
	if testQuit != nil { // hack to make this testable
		quit = testQuit
	}

	signal.Notify(quit, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("gograce: Shutdown Server ...")

		if len(t) > 1 {
			time.Sleep(t[1])
		}

		ctx, cancel := context.WithTimeout(context.Background(), t[0])
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Println("gograce: error server shutdown:", err)
		}
		log.Println("gograce: server exited")
		close(shutdown)
	}()

	return srv, shutdown
}
