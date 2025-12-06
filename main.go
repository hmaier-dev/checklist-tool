package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	_ "embed"
	"os/signal"
	"syscall"
	"time"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/hmaier-dev/checklist-tool/internal/server"
	"github.com/hmaier-dev/checklist-tool/internal/handlers"

	// blank import for handlers. They initalize theirself by init()
	_ "github.com/hmaier-dev/checklist-tool/internal/handlers/all"
	_ "github.com/hmaier-dev/checklist-tool/internal/handlers/checklist"
	_ "github.com/hmaier-dev/checklist-tool/internal/handlers/delete"
	_ "github.com/hmaier-dev/checklist-tool/internal/handlers/new"
	_ "github.com/hmaier-dev/checklist-tool/internal/handlers/upload"
)

//go:embed schema.sql
var ddl string

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
  dbArg := flag.String("db", "", "Path to sqlite database")
  port := flag.String("port", "8080", "Port handling http requests")
  flag.Parse()
  if *dbArg == "" {
    flag.Usage()
    log.Fatalln("database file is mandatory")
  }
	ctx := context.Background()
	db, err := sql.Open("sqlite3", *dbArg)
	if err != nil {
		log.Fatal(err)
	}
	// Server should hold the router and the db-handler
  srv := server.NewServer(db)
	// create tables if not exist
	if _, err := srv.DB.ExecContext(ctx, ddl); err != nil {
		log.Fatal(err)
	}
	
	// Call all registered handlers
	// The handlers register theirself by init(), which is called by blank import
	for _, h := range handlers.GetHandlers() {
		// srv make router and DB accessibly to the handlers
		h.New(srv)
		// bring up the routes for each handler
		h.Routes()
	}

	addr := fmt.Sprintf("0.0.0.0:%s", *port)
	httpServer := &http.Server{
		Addr: addr,
		// srv.Router is filled up by h.Routes() above
		Handler: srv.Router,
		// TODO: test these!
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 10 * time.Second,
	}

	srv.LogRoutes()

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

	err = httpServer.Shutdown(ctxShutdown)
	if err != nil{
		log.Fatalf("shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
