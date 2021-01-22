package api

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// initStatic registers static endpoints on the given router.
func initStatic(rootRouter *mux.Router, context *Context) {
	rootRouter.Handle("/robots.txt", http.HandlerFunc(robotsHandler))
}

type notFoundNoCacheResponseWriter struct {
	http.ResponseWriter
}

func (w *notFoundNoCacheResponseWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusNotFound {
		// we have a 404, update our cache header first then fall through
		w.Header().Set("Cache-Control", "no-cache, public")
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

var robotsTxt = []byte("User-agent: *\nDisallow: /\n")

func robotsHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/") {
		http.NotFound(w, r)
		return
	}
	w.Write(robotsTxt)
}
