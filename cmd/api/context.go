package main

import (
    "context"
    "net/http"

    "github.com/tchenbz/test3AWT/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (a *applicationDependencies) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (a *applicationDependencies) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
