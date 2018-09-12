package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	_ "gitlab.com/nod/teyit/link/statik"
	"gitlab.com/nod/teyit/link/utils"
	"log"
	"net/http"
)

func CreateRoutes() *mux.Router {
	r := mux.NewRouter()

	// Handle homepage inline
	r.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
			RespondSuccessTemplate(w, r, "homepage", nil)
	}).Methods("GET")

	r.HandleFunc("/search", SearchArchives).Methods("GET")
	r.HandleFunc("/api/search", SearchArchivesJson).Methods("GET")


	// below are legacy links from v1, we plan to phase these out
	// but we can't immediately because we suspect programmatic usage
	r.HandleFunc("/new", CreateArchiveHandler).Methods("POST", "GET")
	r.HandleFunc("/bookmark", CreateArchiveHandler).Methods("POST", "GET")
	r.HandleFunc("/add", CreateArchiveHandler).Methods("POST", "GET")
	// Next up, the current API handler for creating archives
	r.HandleFunc("/api/archive", CreateArchiveApiHandler).Methods("POST", "GET")

	// Display the archive
	r.HandleFunc("/{slug}", ShowArchiveHandler).Methods("GET")
	r.HandleFunc("/{slug}", ShowArchiveApiHandler).Methods("GET").HeadersRegexp("Content-Type", "application/json")
	r.HandleFunc("/api/archives/{slug}", ShowArchiveApiHandler).Methods("GET")

	r.HandleFunc("/{slug}/screenshot", ShowArchiveScreenshotHandler).Methods("GET")
	r.HandleFunc("/{slug}/snapshot", ShowArchiveSnapshotHandler).Methods("GET")

	// Handle static files
	var staticServer http.Handler

	// If we are in development, just bind the directory
	if utils.GetConfig().Env == "development" {
		staticServer = http.FileServer(http.Dir("./public"))
	} else {
		// We use statik file system in production, meaning all the assets are bundled inside the binary
		statikFS, err := fs.New()
		if err != nil {
			log.Fatal("statik fs", err)
		}
		staticServer = http.FileServer(statikFS)
	}

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticServer))

	r.NotFoundHandler = http.HandlerFunc(NotFoundPage)

	return r
}

func NotFoundPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "<h2>Sorry Could not Find Resource. 404 Error</h2>")
}

func RespondSuccessJson(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func RespondJson(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// I (@batuhan) absolutely hate cryptic or non-descriptive error messages so this is a
// @TODO: Make sure we handle each error individually with helpful error messages

func RespondInternalServerError(w http.ResponseWriter, _ error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h2>Internal Server Error. Please try again.</h2>")
}

func RespondInternalServerErrorJson(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": "internal server error",
	})
}

func RespondBadRequestErrorJson(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": data,
	})
}

func RespondSuccessTemplate(w http.ResponseWriter, r *http.Request, page string, data interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	view := utils.NewView("default", page)
	view.Render(w, r, data)
}
