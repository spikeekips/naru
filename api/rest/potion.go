package rest

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spikeekips/naru/element"
)

func PotionMiddleware(potion element.Potion) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(
				w,
				r.WithContext(context.WithValue(r.Context(), "potion", potion)),
			)
		})
	}
}
