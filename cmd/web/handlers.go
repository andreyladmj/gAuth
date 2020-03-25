package main

import (
	"fmt"
	"github.com/dghubble/gologin/google"
	"net/http"
)

func (app *application) oauth2callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	googleUser, err := google.UserFromContext(ctx)

	if err != nil {
		app.errorLog.Printf("Get user from context method failed %v", err)
		app.serverError(w, err)
		return
	}

	user, err := app.users.UpdateOrCreate(googleUser.Email, googleUser.Name, googleUser.Picture, googleUser.Gender, googleUser.Locale)

	if err != nil {
		app.errorLog.Printf("UpdateOrCreate method failed %v", err)
		app.serverError(w, err)
		return
	}

	token, err := app.users.CreateToken(user.Email)

	if err != nil {
		app.errorLog.Printf("CreateToken method failed %v", err)
		app.serverError(w, err)
		return
	}

	fmt.Println(token)
	retUrl := app.session.GetString(r, "returnUrl")
	fmt.Printf("%s?token=%s", retUrl, token)

	http.Redirect(w, r, fmt.Sprintf("%s?token=%s", retUrl, token), http.StatusFound)
}
