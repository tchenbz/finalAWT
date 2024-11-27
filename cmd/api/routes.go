package main

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	// Book routes
	router.HandlerFunc(http.MethodGet, "/api/v1/books/search", a.searchBooksHandler)
	router.HandlerFunc(http.MethodPost, "/v1/books", a.requirePermission("comments:write", a.createBookHandler))	
	router.HandlerFunc(http.MethodGet, "/v1/books/:id", a.requirePermission("comments:read",a.displayBookHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/books/:id", a.requirePermission("comments:write", a.updateBookHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/books/:id", a.requirePermission("comments:write", a.deleteBookHandler))
    router.HandlerFunc(http.MethodGet, "/v1/books", a.requirePermission("comments:read", a.listBooksHandler))


	// Reading List routes
	router.HandlerFunc(http.MethodPost, "/v1/lists", a.createReadingListHandler)
	router.HandlerFunc(http.MethodGet, "/v1/lists/:id", a.displayReadingListHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/lists/:id", a.updateReadingListHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/lists/:id", a.deleteReadingListHandler)
	router.HandlerFunc(http.MethodGet, "/v1/lists", a.listReadingListsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/lists/:id/books", a.addBookToReadingListHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/lists/:id/books", a.removeBookFromReadingListHandler)

	// Review routes
	router.HandlerFunc(http.MethodPost, "/v1/books/:id/reviews", a.createReviewHandler)
	router.HandlerFunc(http.MethodGet, "/v1/books/:id/reviews", a.listBookReviewsHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/reviews/:id", a.updateReviewHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/reviews/:id", a.deleteReviewHandler)

	// User routes
	router.HandlerFunc(http.MethodPost, "/v1/users", a.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", a.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", a.createAuthenticationTokenHandler)

	// Return router with panic recovery and rate limiting
	return a.recoverPanic(a.rateLimit(a.authenticate(router)))

}
