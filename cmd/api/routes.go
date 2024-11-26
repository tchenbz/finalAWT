package main

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodPost, "/v1/products", a.createProductHandler)
	router.HandlerFunc(http.MethodGet, "/v1/products/:id", a.displayProductHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/products/:id", a.updateProductHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/products/:id", a.deleteProductHandler)
	router.HandlerFunc(http.MethodGet, "/v1/products", a.listProductsHandler)

	router.HandlerFunc(http.MethodPost, "/v1/products/:id/reviews", a.createReviewHandler)
	router.HandlerFunc(http.MethodGet, "/v1/products/:id/reviews/:review_id", a.displayReviewHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/products/:id/reviews/:review_id", a.updateReviewHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/products/:id/reviews/:review_id", a.deleteReviewHandler)
	router.HandlerFunc(http.MethodGet, "/v1/reviews", a.listReviewsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/products/:id/reviews", a.listProductReviewsHandler)

	//return a.recoverPanic(router)
	return a.recoverPanic(a.rateLimit(router))
}