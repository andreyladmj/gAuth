package main

import (
	"errors"
	"fmt"
	"net/http"
)


func (app *application) saveRedirectUri(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retUrl, ok := r.Header["Return-URL"]

		if ok {
			app.session.Put(r, "returnUrl", retUrl)
		} else {
			retUrl, ok = r.URL.Query()["returnUrl"]

			if ok {
				app.session.Put(r, "returnUrl", retUrl[0])
			}
		}

		app.infoLog.Println("Got return url", retUrl)

		if !ok {
			app.errorLog.Println("There is not return url", fmt.Sprintf("//%s%s", r.Host, r.URL.Path))
			app.infoLog.Printf("%s %s %s", r.Proto, r.Method, r.URL.RequestURI())
			app.serverError(w, errors.New("there is no return url param"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
