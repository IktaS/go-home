package handlers

import (
	"net/http"

	"github.com/IktaS/go-home/internal/app"
)

func appHandlerWrapper(f func(http.ResponseWriter, *http.Request, *app.App), a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r, a)
	}
}
