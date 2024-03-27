package routes

import (
	"io"
	"net/http"
)

func Syncf(w http.ResponseWriter, r *http.Request) {
	/*body*/ _, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
