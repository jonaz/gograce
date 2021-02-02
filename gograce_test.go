package gograce

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeoutPendingRequest(t *testing.T) {
	testQuit = make(chan os.Signal)

	mux := http.NewServeMux()

	server, shutdown := NewServerWithTimeout(1500 * time.Millisecond)
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		rw.WriteHeader(201)
	})
	server.Handler = mux

	testserver := httptest.NewUnstartedServer(nil)
	testserver.Config = server
	testserver.Start()

	go func() {
		err := server.ListenAndServe()
		t.Log(err)
	}()

	got200 := int64(0)
	go func() {
		resp, err := http.Get(testserver.URL)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		atomic.AddInt64(&got200, 1)
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		testQuit <- syscall.SIGTERM
	}()

	<-shutdown
	assert.Equal(t, int64(1), got200)
}

func TestTimeout(t *testing.T) {
	testQuit = make(chan os.Signal)

	mux := http.NewServeMux()

	server, shutdown := NewServerWithTimeout(500*time.Millisecond, 1*time.Second)
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(1200 * time.Millisecond)
		rw.WriteHeader(201)
	})
	server.Handler = mux

	testserver := httptest.NewUnstartedServer(nil)
	testserver.Config = server
	testserver.Start()

	go func() {
		err := server.ListenAndServe()
		t.Log(err)
	}()

	got200 := int64(0)
	go func() {
		resp, err := http.Get(testserver.URL)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		atomic.AddInt64(&got200, 1)
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		testQuit <- syscall.SIGTERM
	}()

	<-shutdown
	assert.Equal(t, int64(1), got200)
}
