package main

import (
	"github.com/go-chi/chi"
	"github.com/minlau/mdb-tool/render"
	"net/http"
	"strconv"
	"strings"
)

func initHandlers(r *chi.Mux, store *DatabaseStore) {
	ServeFiles(r, "/", getStaticDir())
	r.Get("/databases", getDatabases(store))
	r.Get("/tables-metadata", getTablesMetadata(store))
	r.Get("/query", query(store))
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
	GroupId   *int
	GroupType string
	Query     string
}

func query(store *DatabaseStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req queryRequest
		groupIdString := r.URL.Query().Get("groupId")
		if groupIdString != "" {
			groupIdInt, err := strconv.Atoi(groupIdString)
			if err != nil {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, render.M{"error": err.Error()})
				return
			}
			req.GroupId = &groupIdInt
		}

		req.Query = r.URL.Query().Get("query")
		if req.Query == "" {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, render.M{"error": "query is required"})
			return
		}

		req.GroupType = r.URL.Query().Get("groupType")
		if req.GroupType == "" {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, render.M{"error": "groupType is required"})
			return
		}

		if req.GroupId == nil {
			render.Status(r, http.StatusOK)
			render.JSON(w, r, store.QueryMultipleDatabases(req.GroupType, req.Query))
		} else {
			render.Status(r, http.StatusOK)
			render.JSON(w, r, store.QueryDatabase(*req.GroupId, req.GroupType, req.Query))
		}
	}
}

type tablesMetadataRequest struct {
	GroupId   int
	GroupType string
}

func getTablesMetadata(store *DatabaseStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req tablesMetadataRequest
		groupIdString := r.URL.Query().Get("groupId")
		if groupIdString == "" {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, render.M{"error": "groupId is required"})
			return
		}
		groupIdInt, err := strconv.Atoi(groupIdString)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, render.M{"error": err.Error()})
			return
		}
		req.GroupId = groupIdInt

		req.GroupType = r.URL.Query().Get("groupType")
		if req.GroupType == "" {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, render.M{"error": "groupType is required"})
			return
		}

		data, err := store.GetTablesMetadata(req.GroupId, req.GroupType)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, render.M{"error": err.Error()})
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, data)
	}
}

func getDatabases(store *DatabaseStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, store.GetDatabaseItems())
	}
}
