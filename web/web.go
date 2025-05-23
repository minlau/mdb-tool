package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/minlau/mdb-tool/store"
	"github.com/minlau/mdb-tool/web/ui"
	"net/http"
	"strings"
)

func New(databaseStore store.DatabaseStoreI) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Compress(1))
	r.Use(ZeroLogLogger)
	r.Use(middleware.Recoverer)

	initHandlers(r, databaseStore)
	return r
}

func initHandlers(r *chi.Mux, store store.DatabaseStoreI) {
	ServeFiles(r, "/", ui.GetStaticDir())
	r.Get("/databases", getDatabases(store))
	r.Get("/tables-metadata", getTablesMetadata(store))
	r.Get("/query", query(store))
	r.Mount("/debug", middleware.Profiler())
}

func ServeFiles(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("ServeFiles does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}
