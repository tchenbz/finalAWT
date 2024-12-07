package main

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthCheckHandler)

	// Book routes
	router.HandlerFunc(http.MethodGet, "/api/v1/books/search", a.searchBooksHandler)
	router.HandlerFunc(http.MethodPost, "/v1/books", a.requirePermission("comments:write", a.createBookHandler))	
	router.HandlerFunc(http.MethodGet, "/v1/books/:id", a.requirePermission("comments:read",a.displayBookHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/books/:id", a.requirePermission("comments:write", a.updateBookHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/books/:id", a.requirePermission("comments:write", a.deleteBookHandler))
    router.HandlerFunc(http.MethodGet, "/v1/books", a.requirePermission("comments:read", a.listBooksHandler))

	// Reading List routes
	router.HandlerFunc(http.MethodPost, "/v1/lists", a.requirePermission("readinglists:write", a.createReadingListHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/lists/:id", a.requirePermission("readinglists:write", a.updateReadingListHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/lists/:id", a.requirePermission("readinglists:write", a.deleteReadingListHandler))
	router.HandlerFunc(http.MethodPost, "/v1/lists/:id/books", a.requirePermission("readinglists:write", a.addBookToReadingListHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/lists/:id/books", a.requirePermission("readinglists:write", a.removeBookFromReadingListHandler))
	router.HandlerFunc(http.MethodGet, "/v1/lists", a.requirePermission("reading_lists:read", a.listReadingListsHandler))
	router.HandlerFunc(http.MethodGet, "/v1/lists/:id", a.requirePermission("reading_lists:read", a.displayReadingListHandler))

	// Review routes
	router.HandlerFunc(http.MethodPost, "/v1/books/:id/reviews", a.requirePermission("reviews:write", a.createReviewHandler))
	router.HandlerFunc(http.MethodGet, "/v1/books/:id/reviews", a.requirePermission("reviews:read", a.listBookReviewsHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/reviews/:id", a.requirePermission("reviews:write", a.updateReviewHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/reviews/:id", a.requirePermission("reviews:write", a.deleteReviewHandler))

	// User routes
	router.HandlerFunc(http.MethodPost, "/v1/users", a.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", a.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", a.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/users/:id", a.requireActivatedUser(a.getUserProfileHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/users/:id/lists", a.requireActivatedUser(a.getUserReadingListsHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/users/:id/reviews", a.requireActivatedUser(a.getUserReviewsHandler))
	
	router.HandlerFunc(http.MethodPost, "/api/v1/tokens/password-reset", (a.createPasswordResetTokenHandler))


	// Return router with panic recovery and rate limiting
	return a.recoverPanic(a.enableCORS(a.rateLimit(a.authenticate(router))))
}
