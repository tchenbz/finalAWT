package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/tchenbz/test3AWT/internal/data"
	"github.com/tchenbz/test3AWT/internal/validator"
)

func (a *applicationDependencies) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Username:  incomingData.Username,
		Email:     incomingData.Email,
		Activated: false,
	}

	err = user.Password.Set(incomingData.Password)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateUser(v, user)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.userModel.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			a.failedValidationResponse(w, r, v.Errors)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.permissionModel.AddForUser(user.ID, "comments:read")
    if err != nil {
        a.serverErrorResponse(w, r, err)
        return
    }


	token, err := a.tokenModel.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	responseData := envelope{
		"user": user,
	}

	a.background(func() {
		mailData := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		err = a.mailer.Send(user.Email, "user_welcome.tmpl", mailData)
		if err != nil {
			a.logger.Error(err.Error())
		}
	})

	err = a.writeJSON(w, http.StatusCreated, responseData, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}


func (a *applicationDependencies) activateUserHandler(w http.ResponseWriter, r *http.Request) {
		var incomingData struct {
			TokenPlaintext string  `json:"token"`
		}
		err := a.readJSON(w, r, &incomingData)
		if err != nil {
		  a.badRequestResponse(w, r, err)
		  return
		}
	v := validator.New()
	data.ValidateTokenPlaintext(v, incomingData.TokenPlaintext)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := a.userModel.GetForToken(data.ScopeActivation, incomingData.TokenPlaintext)
	if err != nil {
        switch {
            case errors.Is(err, data.ErrRecordNotFound):
                v.AddError("token", "invalid or expired activation token")
                a.failedValidationResponse(w, r, v.Errors)
           default:
                a.serverErrorResponse(w, r, err)
        }
      return
   }
	user.Activated = true
	err = a.userModel.Update(user)
	if err != nil {
	switch {
		case errors.Is(err, data.ErrEditConflict):
			a.editConflictResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
	}
	return
	}

	err = a.tokenModel.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
	a.serverErrorResponse(w, r, err)
	return
	}

	data := envelope {
		"user": user,
	}
    err = a.writeJSON(w, http.StatusOK, data, nil)
    if err != nil {
       a.serverErrorResponse(w, r, err)
    }
}

func (a *applicationDependencies) getUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	user, err := a.userModel.GetByID(userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{"user": user}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) getUserReadingListsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	var input struct {
		Name string
		data.Filters
	}

	query := r.URL.Query()
	input.Name = a.getSingleQueryParameter(query, "name", "")
	input.Filters.Page = a.getSingleIntegerParameter(query, "page", 1, validator.New())
	input.Filters.PageSize = a.getSingleIntegerParameter(query, "page_size", 10, validator.New())
	input.Filters.Sort = a.getSingleQueryParameter(query, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "name", "-id", "-name"}

	v := validator.New()
	data.ValidateFilters(v, input.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	lists, metadata, err := a.readingListModel.GetAllByUser(userID, input.Name, input.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"reading_lists": lists,
		"metadata":      metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) getUserReviewsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	var input struct {
		Content string
		Author  string
		Rating  int
		data.Filters
	}

	query := r.URL.Query()
	input.Content = a.getSingleQueryParameter(query, "content", "")
	input.Author = a.getSingleQueryParameter(query, "author", "")
	input.Rating = a.getSingleIntegerParameter(query, "rating", 0, validator.New())
	input.Filters.Page = a.getSingleIntegerParameter(query, "page", 1, validator.New())
	input.Filters.PageSize = a.getSingleIntegerParameter(query, "page_size", 10, validator.New())
	input.Filters.Sort = a.getSingleQueryParameter(query, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "rating", "helpful_count", "-id", "-rating", "-helpful_count"}

	v := validator.New()
	data.ValidateFilters(v, input.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	reviews, metadata, err := a.reviewModel.GetAllByUser(userID, input.Content, input.Author, input.Rating, input.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"reviews":  reviews,
		"metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
