// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/mattermost/pillar/utils"
)

type contextHandlerFunc func(c *Context, w http.ResponseWriter, r *http.Request)

var _ http.Handler = contextHandler{}

type contextHandler struct {
	context  *Context
	handler  contextHandlerFunc
	isStatic bool
}

func (h contextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	context := h.context.Clone()
	context.RequestID = utils.NewID()
	context.Logger = context.Logger.WithFields(logrus.Fields{
		"path":    r.URL.Path,
		"request": context.RequestID,
	})

	h.setDefaultHeaders(w, r)
	h.handler(context, w, r)
}

func (h contextHandler) setDefaultHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Request-ID", h.context.RequestID)

	if h.isStatic {
		// Instruct the browser not to display us in an iframe unless is the same origin for anti-clickjacking
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		// Set content security policy. This is also specified in the root.html of the webapp in a meta tag.
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		w.Header().Set("Expires", "0")
	}
}

func newStaticHandler(context *Context, handler contextHandlerFunc) *contextHandler {
	return &contextHandler{
		context:  context,
		handler:  handler,
		isStatic: true,
	}
}

func newAPIHandler(context *Context, handler contextHandlerFunc) *contextHandler {
	return &contextHandler{
		context: context,
		handler: handler,
	}
}
