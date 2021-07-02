package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	rClient, closeFunc := initRedisClient()
	defer closeFunc()

	w, closeWorker := newWorker(rClient)
	defer closeWorker()
	go w.Do()

	errChn := make(chan error)
	go func() {
		stopChn := make(chan os.Signal, 1)
		signal.Notify(stopChn, syscall.SIGTERM, syscall.SIGINT)
		log.Printf("exit by signal: %v\n", <-stopChn)
		errChn <- nil
	}()

	srv := http.Server{
		Handler: &hdl{
			rClient: rClient,
			w:       w,
		},
		Addr: httpAddr,
	}
	go func() {
		log.Printf("localhost%s runing...\n", httpAddr)
		errChn <- srv.ListenAndServe()
	}()

	err := <-errChn
	if err != nil {
		log.Printf("Shutdown err: %v\n", err)
	}

	err = srv.Shutdown(context.Background())
	if err != nil {
		log.Printf("http.Server shutdown err: %v\n", err)
	}
}
