package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/husio/bb/forum"
	"github.com/husio/bb/tmpl"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
)

type respwrt struct {
	code int
	http.ResponseWriter
}

func (w *respwrt) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func ctxhandler(ctx context.Context, fn func(context.Context, http.ResponseWriter, *http.Request)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		rw := &respwrt{code: http.StatusOK, ResponseWriter: w}
		c := forum.WithParams(ctx, ps)
		start := time.Now()
		fn(c, rw, r)
		path := r.URL.String() + strings.Repeat(".", 60-len(r.URL.String()))
		fmt.Printf("%4s %d %s %s\n", r.Method, rw.code, path, time.Now().Sub(start))
	}
}

func main() {
	httpAddrFl := flag.String("addr", "localhost:8000", "HTTP server address")
	staticsFl := flag.String("statics", "", "Optional static files directory")
	flag.Parse()

	if err := tmpl.LoadTemplates(); err != nil {
		log.Fatalf("cannot load templates: %s", err)
	}

	ctx := context.Background()
	ctx, err := forum.WithPG(ctx, "user=bb password=bb dbname=bb sslmode=disable")
	if err != nil {
		log.Fatalf("cannot connect to database: %s", err)
	}

	rt := httprouter.New()
	rt.RedirectTrailingSlash = true

	// TODO - configurable?
	rt.GET("/", ctxhandler(ctx, forum.HandleListTopics))

	rt.POST("/nt/", ctxhandler(ctx, forum.HandleCreateTopic))
	rt.GET("/nt/", ctxhandler(ctx, forum.HandleCreateTopic))

	rt.GET("/t/", ctxhandler(ctx, forum.HandleListTopics))
	rt.GET("/t/:topicid/:slug/", ctxhandler(ctx, forum.HandleListTopicMessages))
	rt.POST("/t/:topicid/:slug/", ctxhandler(ctx, forum.HandleCreateMessage))
	rt.GET("/c/", ctxhandler(ctx, forum.HandleListCategories))
	rt.GET("/u/:userid/:slug/", ctxhandler(ctx, forum.HandleUserDetails))

	if *staticsFl != "" {
		rt.ServeFiles("/static/*filepath", http.Dir(*staticsFl))
	}

	log.Println("running server")
	if err := http.ListenAndServe(*httpAddrFl, rt); err != nil {
		log.Printf("HTTP server error: %s", err)
	}
}
