package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/hudsn/morph"
	docfiles "github.com/hudsn/morph/doc_files"
)

func main() {
	generateHTML()
}

func generateHTML() {
	tt, err := template.ParseFS(docfiles.DocFS, "*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	templateData := morph.NewFunctionDocs(morph.DefaultFunctionStore())

	out, err := os.Create("doc_files/out/index.html")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = tt.ExecuteTemplate(out, "base", templateData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generated index.html")

	router := http.NewServeMux()

	fileHandler := http.FileServer(http.Dir("doc_files/out"))
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cache-control", "no-cache")
		fileHandler.ServeHTTP(w, r)

	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	s := &http.Server{
		Handler: router,
		Addr:    ":8080",
	}
	go s.ListenAndServe()
	fmt.Println("Listening on http://localhost:8080")

	<-ctx.Done()
	stop()
	fmt.Println("\nShutting down...")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dChan := make(chan struct{})
	go func() {
		if err := s.Shutdown(timeoutCtx); err != nil {
			log.Fatalln(err)
		}
		dChan <- struct{}{}
	}()

	select {
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() == context.DeadlineExceeded {
			log.Fatalln("timeout exceeded, forcing shutdown")
		}
	case <-dChan:
		os.Exit(0)
	}
}
