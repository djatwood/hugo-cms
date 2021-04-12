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

type templateData struct {
	Title   string
	Request *http.Request
	Data    interface{}
	Site    *site
	Section *section
}

type site struct {
	Sections  map[string]*section
	Dir       string
	templates map[string]tmpl
}

type section struct {
	Label     string
	Path      string
	Match     string
	Extension string
	Templates []string
}

type tmpl struct {
	Label        string
	HideBody     bool
	DisplayField string
	Blocks       []block
}

type block struct {
	Kind        string
	Label       string
	Name        string
	Description string
	Config      map[string]interface{}
}

func main() {
	err := runServer(":4120")
	if err != nil {
		log.Fatal(err)
	}
}

func runServer(addr string) error {
	e := server()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		err := e.Start(addr)
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Println("Server started at " + addr)
	<-done
	log.Println("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return e.Shutdown(ctx)
}
