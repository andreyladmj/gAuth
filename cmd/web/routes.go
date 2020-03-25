package main

import (
	"github.com/bmizerany/pat"
	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/google"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, app.session.Enable)
	dynamicMiddleware := alice.New(app.saveRedirectUri)

	mux := pat.New()

	gcallback := google.CallbackHandler(app.oauth2Config, http.HandlerFunc(app.oauth2callback), nil)

	stateConfig := gologin.DebugOnlyCookieConfig
	mux.Get("/", dynamicMiddleware.Then(google.StateHandler(stateConfig, google.LoginHandler(app.oauth2Config, nil))))
	mux.Get("/oauth2callback", google.StateHandler(stateConfig, gcallback))
	mux.Get("/ping", http.HandlerFunc(ping))

	return standardMiddleware.Then(mux)
}
