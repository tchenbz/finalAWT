package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tchenbz/test3AWT/internal/data"
	"github.com/tchenbz/test3AWT/internal/validator"
)

func (a *applicationDependencies) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Authentication handler triggered")
	fmt.Println("Request method:", r.Method)
	fmt.Println("Request headers:", r.Header)
	var incomingData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()

	data.ValidateEmail(v, incomingData.Email)
	data.ValidatePasswordPlaintext(v, incomingData.Password)

	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	user, err := a.userModel.GetByEmail(incomingData.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.invalidCredentialsResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}
	match, err := user.Password.Matches(incomingData.Password)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		a.invalidCredentialsResponse(w, r)
		return
	}
	token, err := a.tokenModel.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"authentication_token": token,
	}

	err = a.writeJSON(w, http.StatusCreated, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) createPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Email string `json:"email"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()

	data.ValidateEmail(v, incomingData.Email)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := a.userModel.GetByEmail(incomingData.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.invalidCredentialsResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	if !user.Activated {
		a.inactiveAccountResponse(w, r)
		return
	}
	token, err := a.tokenModel.New(user.ID, 60*time.Minute, data.ScopeAuthentication)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	a.background(func() {

		data := map[string]any{
			"password_reset_token": token.Plaintext,
		}

		err = a.mailer.Send(user.Email, "password_reset.tmpl" , data)
		if err != nil {
			a.logger.Error(err.Error())
		}
	})

	message := envelope{
		"message": "Email containing reset token has been sent.",
	}
	err = a.writeJSON(w, http.StatusOK, message, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}
