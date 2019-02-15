package restv1

import (
	"fmt"
	"net/http"
)

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome! This is Naru.\n")
}
