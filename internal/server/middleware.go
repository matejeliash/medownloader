package server

import (
	"log"
	"net/http"
)

// we  use http.Handler interface  so we can use middleware on ServeMux
// with little modification we can use http.HandleFunc so we can easily use middleware on single handler

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// just simple log before handling actual request
		log.Printf("-> %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// checking token and it's expiration date from cookie
func (s *Server) middlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.sessionManger.IsSessionValid(r) {

			// invalidate / remove cookie if
			http.SetCookie(w, &http.Cookie{
				Name:     "medownloader_token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1, // Instant delete
				HttpOnly: true,
			})

			encodeErr(w, "session token not valid / not provided", http.StatusUnauthorized)
			log.Printf("-> %s %s [STOPPED]", r.Method, r.URL.Path)
			return
		}
		next.ServeHTTP(w, r)

	})

}
