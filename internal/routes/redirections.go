// routes is just an example routes package that allows clients to create, and
// delete redirection. And a catch all GET route to redirect to the URL. A lot
// of functionality is missing.
package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/koenbollen/go-tested-api-with-sqlite/internal"
	"github.com/koenbollen/go-tested-api-with-sqlite/internal/util/timeutil"
	"github.com/koenbollen/logging"
)

type CreateRequest struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

func Redirections(ctx context.Context, mux *http.ServeMux, deps *internal.Dependencies) error {
	db := deps.DB

	mux.HandleFunc("POST /redirections", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.GetLogger(ctx)
		request := &CreateRequest{}
		if err := json.NewDecoder(r.Body).Decode(request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Debug("creating redirection", "key", request.Key, "url", request.URL)

		if request.Key == "" || request.URL == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "key and url are required"}`)) //nolint:errcheck
			return
		}

		now := timeutil.Now(ctx)
		if _, err := db.ExecContext(ctx, "INSERT INTO redirection (key, url, created_at, updated_at) VALUES (?, ?, ?, ?)", request.Key, request.URL, now, now); err != nil {
			logger.Error("failed to create redirection", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.Info("created redirection", "key", request.Key, "url", request.URL)
	})

	mux.HandleFunc("GET /{key}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.GetLogger(ctx)
		key := r.PathValue("key")

		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		row := db.QueryRow("SELECT url FROM redirection WHERE key = ?", key)
		var url string
		if err := row.Scan(&url); err != nil && err != sql.ErrNoRows {
			logger.Error("failed to query redirection", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if url == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Location", url)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusFound)
		w.Write([]byte(`<script type=\"text/javascript\">window.location = "` + url + `";</script>`)) //nolint:errcheck
	})

	mux.HandleFunc("DELETE /redirections/{key}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.GetLogger(ctx)
		key := r.PathValue("key")

		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if _, err := db.ExecContext(ctx, "DELETE FROM redirection WHERE key = ?", key); err != nil {
			logger.Error("failed to delete redirection", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

		logger.Info("deleted redirection", "key", key)
	})

	return nil
}
