package server

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

type SesssionManager struct {
	mu       sync.RWMutex
	sessions map[string]time.Time // identidy by token string
	validity time.Duration
}

func NewSessionManager(validity time.Duration) *SesssionManager {
	return &SesssionManager{
		sessions: make(map[string]time.Time),
		validity: validity,
	}
}

func (s *SesssionManager) CreateSession(w http.ResponseWriter) {

	// create random 32 byte array and encode it to base64
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	token := base64.URLEncoding.EncodeToString(randomBytes)
	expTime := time.Now().Add(s.validity)

	s.mu.Lock()
	s.sessions[token] = expTime
	s.mu.Unlock()

	cookie := &http.Cookie{
		Expires:  expTime,
		Name:     "medownloader_token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   false,

		//Secure: true,  // use this for https only
	}

	http.SetCookie(w, cookie)
}

// find if token in map and if it is still valid
func (s *SesssionManager) IsSessionValid(r *http.Request) bool {
	cookie, err := r.Cookie("medownloader_token")
	if err != nil {
		return false
	}
	s.mu.RLock()
	expTime, exists := s.sessions[cookie.Value]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().Before(expTime) {
		return true
	}

	s.mu.Lock()
	delete(s.sessions, cookie.Value)
	s.mu.Unlock()

	return false

}
