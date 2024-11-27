package main

import (
	"errors"
	"net/http"

	"github.com/tchenbz/test3AWT/internal/data"
	"github.com/tchenbz/test3AWT/internal/validator"
)

func (a *applicationDependencies) createReviewHandler(w http.ResponseWriter, r *http.Request) {
    user := a.contextGetUser(r)

    bookID, err := a.readIDParam(r)
    if err != nil {
        a.notFoundResponse(w, r)
        return
    }

    var input struct {
        Content string `json:"content"`
        Rating  int    `json:"rating"`
    }

    err = a.readJSON(w, r, &input)
    if err != nil {
        a.badRequestResponse(w, r, err)
        return
    }

    review := &data.Review{
        BookID:  bookID,
        Author:  user.Username,
        Content: input.Content,
        Rating:  input.Rating,
    }

    err = a.reviewModel.Insert(review)
    if err != nil {
        a.serverErrorResponse(w, r, err)
        return
    }

    data := envelope{"review": review}
    err = a.writeJSON(w, http.StatusCreated, data, nil)
    if err != nil {
        a.serverErrorResponse(w, r, err)
    }
}

func (a *applicationDependencies) updateReviewHandler(w http.ResponseWriter, r *http.Request) {
    user := a.contextGetUser(r) 

    reviewID, err := a.readIDParam(r)
    if err != nil {
        a.notFoundResponse(w, r)
        return
    }

    review, err := a.reviewModel.Get(reviewID)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            a.notFoundResponse(w, r)
        default:
            a.serverErrorResponse(w, r, err)
        }
        return
    }

    if review.Author != user.Username {
        a.notPermittedResponse(w, r)
        return
    }

    var input struct {
        Content      *string `json:"content"`
        Rating       *int    `json:"rating"`
        HelpfulCount *int    `json:"helpful_count"`
    }

    err = a.readJSON(w, r, &input)
    if err != nil {
        a.badRequestResponse(w, r, err)
        return
    }

    if input.Content != nil {
        review.Content = *input.Content
    }
    if input.Rating != nil {
        review.Rating = *input.Rating
    }
    if input.HelpfulCount != nil {
        review.HelpfulCount = *input.HelpfulCount
    }

    err = a.reviewModel.Update(review)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrEditConflict):
            a.editConflictResponse(w, r)
        default:
            a.serverErrorResponse(w, r, err)
        }
        return
    }

    data := envelope{"review": review}
    err = a.writeJSON(w, http.StatusOK, data, nil)
    if err != nil {
        a.serverErrorResponse(w, r, err)
    }
}

func (a *applicationDependencies) deleteReviewHandler(w http.ResponseWriter, r *http.Request) {
    user := a.contextGetUser(r) 

    reviewID, err := a.readIDParam(r)
    if err != nil {
        a.notFoundResponse(w, r)
        return
    }

    review, err := a.reviewModel.Get(reviewID)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            a.notFoundResponse(w, r)
        default:
            a.serverErrorResponse(w, r, err)
        }
        return
    }

    if review.Author != user.Username {
        a.notPermittedResponse(w, r)
        return
    }

    err = a.reviewModel.Delete(reviewID)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            a.notFoundResponse(w, r)
        default:
            a.serverErrorResponse(w, r, err)
        }
        return
    }

    data := envelope{"message": "review successfully deleted"}
    err = a.writeJSON(w, http.StatusOK, data, nil)
    if err != nil {
        a.serverErrorResponse(w, r, err)
    }
}

func (a *applicationDependencies) listBookReviewsHandler(w http.ResponseWriter, r *http.Request) {
	bookID, err := a.readIDParam(r)
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

	reviews, metadata, err := a.reviewModel.GetAllForBook(bookID, input.Content, input.Author, input.Rating, input.Filters)
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
