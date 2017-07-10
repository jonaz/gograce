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

func NewServerWithTimeout(t time.Duration) (*http.Server, chan struct{}) {

	shutdown := make(chan struct{})
	srv := &http.Server{}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("gograce: Shutdown Server ...")

		ctx, cancel := context.WithTimeout(context.Background(), t)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Println("gograce: error server shutdown:", err)
		}
		close(shutdown)
		log.Println("gograce: server exited")
	}()

	return srv, shutdown
}
