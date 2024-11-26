package main

import (
	"errors"
	"net/http"

	"github.com/tchenbz/AWT_Test1/internal/data"
	"github.com/tchenbz/AWT_Test1/internal/validator"
)

// createReviewHandler handles POST requests for creating a new review.
func (a *applicationDependencies) createReviewHandler(w http.ResponseWriter, r *http.Request) {
	productID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	var input struct {
		Content string `json:"content"`
		Author  string `json:"author"`
		Rating  int    `json:"rating"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	review := &data.Review{
		ProductID: productID,
		Content:   input.Content,
		Author:    input.Author,
		Rating:    input.Rating,
	}

	// v := validator.New()
	// data.ValidateReview(v, review)
	// if !v.IsEmpty() {
	// 	a.failedValidationResponse(w, r, v.Errors)
	// 	return
	// }

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

// displayReviewHandler handles GET requests for displaying a specific review by product and review ID.
func (a *applicationDependencies) displayReviewHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the product ID and review ID from the URL
	productID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}
	reviewID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Fetch the review from the database
	review, err := a.reviewModel.Get(productID, reviewID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Send the review data in JSON format
	data := envelope{"review": review}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// updateReviewHandler handles PATCH requests for updating a specific review by product and review ID.
func (a *applicationDependencies) updateReviewHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the product ID and review ID from the URL
	productID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}
	reviewID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Retrieve the existing review from the database
	review, err := a.reviewModel.Get(productID, reviewID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Create a temporary struct for incoming updates
	var input struct {
		Content      *string `json:"content"`
		Author       *string `json:"author"`
		Rating       *int    `json:"rating"`
		HelpfulCount *int    `json:"helpful_count"`
	}

	// Decode the request JSON into the input struct
	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Update the review fields as needed
	if input.Content != nil {
		review.Content = *input.Content
	}
	if input.Author != nil {
		review.Author = *input.Author
	}
	if input.Rating != nil {
		review.Rating = *input.Rating
	}
	if input.HelpfulCount != nil {
		review.HelpfulCount = *input.HelpfulCount
	}

	// // Validate the updated review
	// v := validator.New()
	// data.ValidateReview(v, review)
	// if !v.IsEmpty() {
	// 	a.failedValidationResponse(w, r, v.Errors)
	// 	return
	// }

	err = a.reviewModel.Update(review)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{"review": review}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) deleteReviewHandler(w http.ResponseWriter, r *http.Request) {
	productID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}
	reviewID, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.reviewModel.Delete(productID, reviewID)
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

func (a *applicationDependencies) listReviewsHandler(w http.ResponseWriter, r *http.Request) {
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

	reviews, metadata, err := a.reviewModel.GetAll(input.Content, input.Author, input.Rating, input.Filters)
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

func (a *applicationDependencies) listProductReviewsHandler(w http.ResponseWriter, r *http.Request) {
	productID, err := a.readIDParam(r)
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

	reviews, metadata, err := a.reviewModel.GetAllForProduct(productID, input.Content, input.Author, input.Rating, input.Filters)
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