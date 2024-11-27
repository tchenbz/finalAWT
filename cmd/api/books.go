package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tchenbz/test3AWT/internal/data"
	"github.com/tchenbz/test3AWT/internal/validator"
)

func (a *applicationDependencies) createBookHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string   `json:"title"`
		Authors     []string `json:"authors"`
		ISBN        string   `json:"isbn"`
		Publication string   `json:"publication_date"`
		Genre       string   `json:"genre"`
		Description string   `json:"description"`
	}

	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	parsedDate, err := time.Parse("2006-01-02", input.Publication)
	if err != nil {
		a.failedValidationResponse(w, r, map[string]string{"publication_date": "must be a valid date in YYYY-MM-DD format"})
		return
	}

	book := &data.Book{
		Title:       input.Title,
		Authors:     input.Authors,
		ISBN:        input.ISBN,
        Publication: parsedDate, 
		Genre:       input.Genre,
		Description: input.Description,
	}

	v := validator.New()
	data.ValidateBook(v, book)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.bookModel.Insert(book)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/books/%d", book.ID))

	data := envelope{"book": book}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) displayBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	book, err := a.bookModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{"book": book}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) updateBookHandler(w http.ResponseWriter, r *http.Request) {
    id, err := a.readIDParam(r)
    if err != nil {
        a.notFoundResponse(w, r)
        return
    }

    book, err := a.bookModel.Get(id)
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
        Title       *string   `json:"title"`
        Authors     *[]string `json:"authors"`
        ISBN        *string   `json:"isbn"`
        Publication *string   `json:"publication_date"` 
        Genre       *string   `json:"genre"`
        Description *string   `json:"description"`
    }

    err = a.readJSON(w, r, &input)
    if err != nil {
        a.badRequestResponse(w, r, err)
        return
    }

    if input.Title != nil {
        book.Title = *input.Title
    }
    if input.Authors != nil {
        book.Authors = *input.Authors
    }
    if input.ISBN != nil {
        book.ISBN = *input.ISBN
    }
    if input.Publication != nil {
        parsedDate, err := time.Parse("2006-01-02", *input.Publication)
        if err != nil {
            a.failedValidationResponse(w, r, map[string]string{"publication_date": "must be a valid date in YYYY-MM-DD format"})
            return
        }
        book.Publication = parsedDate
    }
    if input.Genre != nil {
        book.Genre = *input.Genre
    }
    if input.Description != nil {
        book.Description = *input.Description
    }

    v := validator.New()
    data.ValidateBook(v, book)
    if !v.IsEmpty() {
        a.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = a.bookModel.Update(book)
    if err != nil {
        a.serverErrorResponse(w, r, err)
        return
    }

    data := envelope{"book": book}
    err = a.writeJSON(w, http.StatusOK, data, nil)
    if err != nil {
        a.serverErrorResponse(w, r, err)
    }
}

func (a *applicationDependencies) deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.bookModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{"message": "book successfully deleted"}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) listBooksHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Author string
		Genre  string
		data.Filters
	}

	query := r.URL.Query()
	input.Title = a.getSingleQueryParameter(query, "title", "")
	input.Author = a.getSingleQueryParameter(query, "author", "")
	input.Genre = a.getSingleQueryParameter(query, "genre", "")
	input.Filters.Page = a.getSingleIntegerParameter(query, "page", 1, validator.New())
	input.Filters.PageSize = a.getSingleIntegerParameter(query, "page_size", 10, validator.New())
	input.Filters.Sort = a.getSingleQueryParameter(query, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "title", "genre", "authors", "-id", "-title", "-genre", "-authors"}

	v := validator.New()
	data.ValidateFilters(v, input.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	books, metadata, err := a.bookModel.GetAll(input.Title, input.Author, input.Genre, input.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"books":    books,
		"metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) searchBooksHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Author string
		Genre  string
		data.Filters
	}

	query := r.URL.Query()
	input.Title = a.getSingleQueryParameter(query, "title", "")
	input.Author = a.getSingleQueryParameter(query, "author", "")
	input.Genre = a.getSingleQueryParameter(query, "genre", "")
	input.Filters.Page = a.getSingleIntegerParameter(query, "page", 1, validator.New())
	input.Filters.PageSize = a.getSingleIntegerParameter(query, "page_size", 10, validator.New())
	input.Filters.Sort = a.getSingleQueryParameter(query, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "title", "genre", "authors", "-id", "-title", "-genre", "-authors"}

	v := validator.New()
	data.ValidateFilters(v, input.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	books, metadata, err := a.bookModel.GetAll(input.Title, input.Author, input.Genre, input.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"books":    books,
		"metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
