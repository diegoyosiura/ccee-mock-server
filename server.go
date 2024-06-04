package main

import (
	"ampereconsultoria.com.br/ccee/mock/pld"
	"context"
	mongoDb "github.com/ampere-consultoria/sagace-v2-mongod"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var srv *http.Server
var ctx = context.Background()

func main() {
	stop := make(chan bool)
	signalListener := make(chan os.Signal, 1)
	if mongoDb.InitMongoDatabase("localhost", 27017, "ampere", "Ampere159") != nil {
		return
	}

	go func() {
		signal.Notify(signalListener, os.Interrupt, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGSTOP, syscall.SIGTERM)
		<-signalListener
		stop <- true
	}()

	go serve()
	<-stop

	_ = srv.Shutdown(ctx)
}

func serve() {
	r := mux.NewRouter()
	r.HandleFunc("/ws/prec/PLDBSv1", pld.PLDHandler)

	srv = &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8888",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
