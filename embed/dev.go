package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const port = ":8081"

func main() {
	log.SetFlags(0)
	log.Print("starting the service...")
	defer log.Print("shutting down the service...")
	// Closing signal
	stopChan := make(chan os.Signal, 1)
	signal.Notify(
		stopChan,
		// syscall.SIGKILL, // never gets caught on POSIX
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	sm := http.NewServeMux()

	sm.Handle("/", l{})
	srv := &http.Server{
		Addr:    port,
		Handler: sm,
	}
	log.Println("serving on", port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-stopChan
	ctxShutDown, cancel := context.WithTimeout(
		context.Background(),
		time.Second*10,
	)
	defer func() { cancel() }()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("server Shutdown Failed: %v", err)
	}

}

type l struct{}

func (b l) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] serving %s\n", time.Now().Format(time.RFC3339), r.URL.Path)
	http.FileServer(http.Dir(".")).ServeHTTP(w, r)
}
