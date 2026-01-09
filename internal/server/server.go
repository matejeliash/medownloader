package server

import (
	"net/http"

	_ "embed"

	"github.com/matejeliash/medownloader/internal/downloader"
)

//go:embed index.html
var indexHTML []byte

//go:embed script.js
var scriptJS []byte

//go:embed img.jpg
var img []byte

type Server struct {
	downloadManager *downloader.DownloadManager
	sessionManger   *SesssionManager
	*http.Server
}

func New(dManager *downloader.DownloadManager, sManager *SesssionManager, address string) *Server {

	mainMux := http.NewServeMux()

	// register routes
	server := &Server{
		downloadManager: dManager,
		sessionManger:   sManager,
	}
	// serve index.html /{$} just allow /
	mainMux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexHTML)
	})
	// serve js script
	mainMux.HandleFunc("GET /script.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(scriptJS)
	})

	mainMux.HandleFunc("GET /img.jpg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpg")
		w.Write(img)
	})

	// unprotected login route
	mainMux.HandleFunc("POST /api/login", server.LoginHandler)

	// create subrouter for all api router
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /downloads", server.GetAllDownloadsHandler)
	apiMux.HandleFunc("GET /info", server.GetCurDirInfoHandler)
	apiMux.HandleFunc("POST /add", server.AddAndStartDownloadHandler)
	apiMux.HandleFunc("GET /toggle/{id}", server.ToggleHandler)
	apiMux.HandleFunc("GET /delete/{id}", server.DeleteHandler)
	apiMux.HandleFunc("GET /logout", server.LogoutHandler)

	// user middleware and assign /api prefix
	protectedApiMux := server.middlewareAuth(apiMux)
	mainMux.Handle("/api/", http.StripPrefix("/api", protectedApiMux))

	server.Server = &http.Server{
		Addr:    address,
		Handler: middlewareLog(mainMux), // apply log to all endpoints
	}

	return server
}

// run server
func (s *Server) Run() error {
	return s.ListenAndServe()
}
