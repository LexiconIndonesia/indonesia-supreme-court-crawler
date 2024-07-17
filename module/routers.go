package module

import (
	"lexicon/go-template/common/utils"
	"net/http"

	"github.com/go-chi/chi"
)

func Router() *chi.Mux {

	r := chi.NewMux()
	r.Get("/", testRoute)
	return r
}

func testRoute(w http.ResponseWriter, req *http.Request) {

	utils.WriteMessage(w, 200, "Hello")
}
