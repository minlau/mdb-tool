package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/minlau/mdb-tool/render"
	"net/http"
	"strings"
)

func initHandlers(r *chi.Mux, store *DatabaseStore) {
	ServeFiles(r, "/", getStaticDir())
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
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}

type queryRequest struct {
	GroupName *string
	GroupType string
	Query     string
}

func query(store *DatabaseStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req queryRequest
		groupNameString := r.URL.Query().Get("groupName")
		if groupNameString != "" {
			req.GroupName = &groupNameString
		}

		req.Query = r.URL.Query().Get("query")
		if req.Query == "" {
			render.JSON(w, http.StatusBadRequest, render.M{"error": "query is required"})
			return
		}

		req.GroupType = r.URL.Query().Get("groupType")
		if req.GroupType == "" {
			render.JSON(w, http.StatusBadRequest, render.M{"error": "groupType is required"})
			return
		}

		var res interface{}
		if req.GroupName == nil {
			res = store.QueryMultipleDatabases(r.Context(), req.GroupType, req.Query)
		} else {
			res = store.QueryDatabase(r.Context(), *req.GroupName, req.GroupType, req.Query)
		}
		render.JSON(w, http.StatusOK, res)
	}
}

type tablesMetadataRequest struct {
	GroupName string
	GroupType string
}

func getTablesMetadata(store *DatabaseStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req tablesMetadataRequest
		req.GroupName = r.URL.Query().Get("groupName")
		if req.GroupName == "" {
			render.JSON(w, http.StatusBadRequest, render.M{"error": "groupName is required"})
			return
		}

		req.GroupType = r.URL.Query().Get("groupType")
		if req.GroupType == "" {
			render.JSON(w, http.StatusBadRequest, render.M{"error": "groupType is required"})
			return
		}

		data, err := store.GetTablesMetadata(req.GroupName, req.GroupType)
		if err != nil {
			render.JSON(w, http.StatusBadRequest, render.M{"error": err.Error()})
			return
		}

		render.JSON(w, http.StatusOK, data)
	}
}

func getDatabases(store *DatabaseStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, http.StatusOK, store.GetDatabaseItems())
	}
}
