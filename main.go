package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hmaier-dev/checklist-tool/internal/database"
	"github.com/hmaier-dev/checklist-tool/internal/server"
)



func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
  dbArg := flag.String("db", "", "Path to sqlite database")
  port := flag.String("port", "8080", "Port handling http requests")
  flag.Parse()
  if *dbArg == "" {
    flag.Usage()
    log.Fatalln("database file is mandatory")
  }
  database.DBfilePath = *dbArg

  srv := server.NewServer()
	addr := fmt.Sprintf("0.0.0.0:%s", *port)
	httpServer := &http.Server{
		Addr: addr,
		Handler: srv.Router, // this is my custom router!
		// TODO: test these!
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 10 * time.Second,
	}


	go func() {
		log.Printf("Starting tool on %s \n", addr)
		// http.ErrServerClosed is returned form httpServer.Shutdown
		// It is a normal error and should not trigger log.Fatalf
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed{
			log.Fatalf("Error while listening: %v \n", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	// os.Interrupt for terminal session,
	// syscall.SIGTERM for stopping a container
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	// this stops main, because it listens for the upper signals
	<-stop

	// Give the server 5 seconds to shutdown
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	err := httpServer.Shutdown(ctxShutdown)
	if err != nil{
		log.Fatalf("shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
