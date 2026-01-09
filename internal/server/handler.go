package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/matejeliash/medownloader/internal/dto"
)

func (s *Server) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if s.sessionManger.IsSessionValid(r) {

		// invalidate / remove cookie if
		http.SetCookie(w, &http.Cookie{
			Name:     "medownloader_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1, // Instant delete
			HttpOnly: true,
		})

		encodeJson(w, nil, http.StatusOK)
		return
	}

}

// used for login / verifying token in cookies
func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {

	// used for skipping login when session token is still valid
	if s.sessionManger.IsSessionValid(r) {
		encodeJson(w, dto.MsgResponse{Msg: "OK"}, http.StatusOK)
		return
	}

	var data dto.LoginDto
	decodeJson(r, &data)
	// compare passwords
	if data.Password != os.Getenv("ME_PASSWORD") {
		encodeErr(w, "incorrect password", http.StatusUnauthorized)
		return
	}
	s.sessionManger.CreateSession(w)

	resp := dto.MsgResponse{Msg: "ok"}

	encodeJson(w, resp, http.StatusOK)

}

func (s *Server) AddAndStartDownloadHandler(w http.ResponseWriter, r *http.Request) {

	var data dto.AddDownloadDto
	decodeJson(r, &data)
	//fmt.Printf("%v\n", data)

	if data.Url == "" || !isUrlValid(data.Url) {
		encodeErr(w, "url is invalid", http.StatusBadRequest)
		return
	}

	// get directory
	var dir string
	// user current directory of running program
	if data.Dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			encodeErr(w, "could not access directory", http.StatusInternalServerError)
			return
		}
		dir = wd
		// directory set in form
	} else {

		info, err := os.Stat(data.Dir)
		if err != nil {
			encodeErr(w, "could not access directory", http.StatusInternalServerError)
			return
		}

		if !info.IsDir() {
			encodeErr(w, "directory is file ", http.StatusInternalServerError)
			return
		}
	}

	var filename string
	if data.Filename == "" {
		filename = getFilenameFromUrl(data.Url)
	} else {
		filename = data.Filename
	}

	// create final filepath
	finalPath := filepath.Join(dir, filename)

	if PathExists(finalPath) {
		filename = GetCurTimeStr() + "-" + filename
		finalPath = filepath.Join(dir, filename)
	}

	item := s.downloadManager.AddDownload(data.Url, finalPath, filename)
	s.downloadManager.StartDownload(item)

	respData := dto.FileResponse{
		Id:       item.Id,
		Filename: item.Filename,
	}

	encodeJson(w, respData, http.StatusAccepted)

}

func (s *Server) GetAllDownloadsHandler(w http.ResponseWriter, r *http.Request) {

	downloads := s.downloadManager.GetAllDownloads()
	encodeJson(w, downloads, http.StatusAccepted)

}

// get current dir into {path, freespace}
func (s *Server) GetCurDirInfoHandler(w http.ResponseWriter, r *http.Request) {
	info := getCurDirInfo()

	encodeJson(w, info, http.StatusOK)

}

// resume / stop download
func (s *Server) ToggleHandler(w http.ResponseWriter, r *http.Request) {

	// get id
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		encodeErr(w, fmt.Sprintf("wrong id : %s", idStr), http.StatusInternalServerError)
		return
	}

	item := s.downloadManager.GetItemById(int64(id))
	if item == nil {
		encodeErr(w, fmt.Sprintf("downloadItem with id:%s not found", idStr), http.StatusInternalServerError)
		return

	}

	// decoding id stop download or resume
	if item.Active {
		s.downloadManager.StopDownload(item)
		log.Println("stopped download ", id)
	}

	if !item.Active && !item.Completed {
		s.downloadManager.ResumeDownload(int64(id))
		log.Println("resumed download ", id)
	}

}

// delete download with proper http client and goroutine cancellation
func (s *Server) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		encodeErr(w, fmt.Sprintf("wrong id: %s", idStr), http.StatusInternalServerError)
		return
	}

	err = s.downloadManager.DeleteDownload(int64(id))
	if err != nil {
		encodeErr(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encodeJson(w, "removed", http.StatusOK)

}

// get filename from url, e.g. http://...../123.txt -> 123.txt
func getFilenameFromUrl(url string) string {

	urlParts := strings.Split(url, "/")
	if len(urlParts) > 0 {
		return urlParts[len(urlParts)-1]
	} else {

		now := time.Now()
		// Layout string: 2006 = Year, 01 = Month, 02 = Day, 15 = Hour (24h), 04 = Minute, 05 = Second
		formatted := now.Format("2006-01-02-15-04-05")
		return formatted
	}
}

// find if resource is http
func isUrlValid(urlStr string) bool {
	// split URL into parts
	urlObj, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}

	if urlObj.Scheme != "http" && urlObj.Scheme != "https" {
		return false
	}

	if urlObj.Host == "" {
		return false
	}

	return true
}

// find if file with path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

// get current time in human readable string year ... second
func GetCurTimeStr() string {

	now := time.Now()
	// specific time format
	formatted := now.Format("2006-01-02-15-04-05")
	return formatted
}

func isOSUnixLike() bool {

	// all common unix-like
	switch runtime.GOOS {
	case "linux", "darwin", "freebsd", "openbsd", "netbsd", "dragonfly", "solaris":
		return true
	default:
		return false
	}
}

// get directory name where downloader is run from and also free space on disk
// works just for unix  now
func getCurDirInfo() dto.CurDirInfo {

	info := dto.CurDirInfo{}
	path, err := os.Getwd()
	if err != nil {
		info.Path = "unknown"
	} else {
		info.Path = path
	}

	if !isOSUnixLike() {
		info.FreeSpace = "unknown"
	} else {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(path, &stat); err != nil {
			info.FreeSpace = "unknown"
		} else {
			byteCount := stat.Bavail * uint64(stat.Bsize)
			info.FreeSpace = fmt.Sprintf("%.2f GB", float64(byteCount)/1_000_000_000.0) // in GB
		}
	}

	return info
}
