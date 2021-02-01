package api

import (
	"github.com/gorilla/mux"
)

const pathPrefix = "/api/v1"

// Register registers the API endpoints on the given router.
func Register(router *mux.Router, context *Context) {
	initializeRoutes(router, context)
}

// initializeRoutes instantiates routes that listen on the public interface
func initializeRoutes(rootRouter *mux.Router, context *Context) {
	apiRouter := rootRouter.PathPrefix(pathPrefix).Subrouter()

	initWorkspace(apiRouter, context)
	initStatic(rootRouter, context)
}
