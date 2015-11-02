package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
)

func ctxhandler(ctx context.Context, fn func(context.Context, http.ResponseWriter, *http.Request)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		c := context.WithValue(ctx, "httprouter:params", ps)
		start := time.Now()
		fn(c, w, r)
		fmt.Printf("%4s %-20s: work time: %s\n", r.Method, r.URL.Path, time.Now().Sub(start))
	}
}

func Param(ctx context.Context, name string) string {
	ps := ctx.Value("httprouter:params").(httprouter.Params)
	return ps.ByName(name)
}

func main() {
	httpAddrFl := flag.String("addr", "localhost:8000", "HTTP server address")
	flag.Parse()

	ctx := context.Background()
	ctx, err := WithPG(ctx, "user=bb password=bb dbname=bb sslmode=disable")
	if err != nil {
		log.Fatalf("cannot connect to database: %s", err)
	}

	rt := httprouter.New()
	rt.RedirectTrailingSlash = true
	rt.GET("/", ctxhandler(ctx, handleListTopics))
	rt.GET("/t/", ctxhandler(ctx, handleCreateTopic))
	rt.POST("/t/", ctxhandler(ctx, handleCreateTopic))
	rt.GET("/t/:topic/*ignore", ctxhandler(ctx, handleTopicMessages))
	rt.POST("/t/:topic/*ignore", ctxhandler(ctx, handleCreateMessage))

	log.Println("running server")
	if err := http.ListenAndServe(*httpAddrFl, rt); err != nil {
		log.Printf("HTTP server error: %s", err)
	}
}
