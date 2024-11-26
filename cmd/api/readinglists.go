package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/tchenbz/test3AWT/internal/data"
	"github.com/tchenbz/test3AWT/internal/validator"
)

// createReadingListHandler handles creating a new reading list.
func (a *applicationDependencies) createReadingListHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedBy   int64  `json:"created_by"`
		Status      string `json:"status"`
	}

	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	list := &data.ReadingList{
		Name:        input.Name,
		Description: input.Description,
		CreatedBy:   input.CreatedBy,
		Status:      input.Status,
	}

	// v := validator.New()
	// data.ValidateReadingList(v, list)
	// if !v.IsEmpty() {
	// 	a.failedValidationResponse(w, r, v.Errors)
	// 	return
	// }

	err = a.readingListModel.Insert(list)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", "/api/v1/lists/"+strconv.FormatInt(list.ID, 10))

	data := envelope{"reading_list": list}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// displayReadingListHandler handles fetching a specific reading list by ID.
func (a *applicationDependencies) displayReadingListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	list, err := a.readingListModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{"reading_list": list}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// updateReadingListHandler handles updating a reading list by ID.
func (a *applicationDependencies) updateReadingListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	list, err := a.readingListModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		list.Name = *input.Name
	}
	if input.Description != nil {
		list.Description = *input.Description
	}
	if input.Status != nil {
		list.Status = *input.Status
	}

	// v := validator.New()
	// data.ValidateReadingList(v, list)
	// if !v.IsEmpty() {
	// 	a.failedValidationResponse(w, r, v.Errors)
	// 	return
	// }

	err = a.readingListModel.Update(list)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{"reading_list": list}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// deleteReadingListHandler handles deleting a reading list by ID.
func (a *applicationDependencies) deleteReadingListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.readingListModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{"message": "reading list successfully deleted"}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// addBookToReadingListHandler handles adding a book to a reading list.
func (a *applicationDependencies) addBookToReadingListHandler(w http.ResponseWriter, r *http.Request) {
	listID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	var input struct {
		BookID int64 `json:"book_id"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	err = a.readingListModel.AddBook(listID, input.BookID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{"message": "book added to reading list"}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// removeBookFromReadingListHandler handles removing a book from a reading list.
func (a *applicationDependencies) removeBookFromReadingListHandler(w http.ResponseWriter, r *http.Request) {
	listID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	var input struct {
		BookID int64 `json:"book_id"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	err = a.readingListModel.RemoveBook(listID, input.BookID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{"message": "book removed from reading list"}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// listReadingListsHandler handles listing all reading lists with optional pagination.
func (a *applicationDependencies) listReadingListsHandler(w http.ResponseWriter, r *http.Request) {
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

	lists, metadata, err := a.readingListModel.GetAll(input.Name, input.Filters)
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
