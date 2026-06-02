package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/naimbob95/sharejer/internal/db"
	"github.com/naimbob95/sharejer/internal/model"
	"github.com/naimbob95/sharejer/internal/storage"
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	db         *db.DB
	store      storage.Storage
	baseURL    string
	maxUpload  int64
	fileExpiry time.Duration
}

func New(database *db.DB, store storage.Storage, baseURL string, maxUpload int64, fileExpiry time.Duration) *Server {
	return &Server{
		db:         database,
		store:      store,
		baseURL:    baseURL,
		maxUpload:  maxUpload,
		fileExpiry: fileExpiry,
	}
}

func (s *Server) FileUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.maxUpload)

	if err := r.ParseMultipartForm(s.maxUpload); err != nil {
		msg := fmt.Sprintf(`{"error":"file too large, max %dMB"}`, s.maxUpload>>20)
		http.Error(w, msg, http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"missing file field"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := sanitizeFilename(header.Filename)
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(filename))
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	password := r.FormValue("password")

	var passwordHash *string
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}
		hashStr := string(hash)
		passwordHash = &hashStr
	}

	id := generateSlug(8)
	storageFilename := id + "-" + filename
	storagePath, err := s.store.Save(storageFilename, file)
	if err != nil {
		log.Printf("failed to save file: %v", err)
		http.Error(w, `{"error":"failed to store file"}`, http.StatusInternalServerError)
		return
	}

	now := time.Now()
	f := &model.File{
		ID:           id,
		Filename:     filename,
		StoragePath:  storagePath,
		Size:         header.Size,
		MimeType:     mimeType,
		PasswordHash: passwordHash,
		CreatedAt:    now,
	}
	if s.fileExpiry > 0 {
		expiresAt := now.Add(s.fileExpiry)
		f.ExpiresAt = &expiresAt
	}

	if err := s.db.InsertFile(f); err != nil {
		s.store.Delete(storagePath)
		log.Printf("failed to insert file record: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	resp := model.UploadResponse{
		ID:       id,
		ShareURL: s.baseURL + "/f/" + id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Config exposes the server's upload limits so the frontend can show them.
func (s *Server) Config(w http.ResponseWriter, r *http.Request) {
	resp := model.ConfigResponse{
		MaxUploadBytes: s.maxUpload,
		MaxUploadMB:    s.maxUpload >> 20,
		ExpirySeconds:  int64(s.fileExpiry.Seconds()),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) FileMeta(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"missing id"}`, http.StatusBadRequest)
		return
	}

	f, ok := s.liveFile(w, id)
	if !ok {
		return
	}

	meta := model.FileMeta{
		ID:            f.ID,
		Filename:      f.Filename,
		Size:          f.Size,
		MimeType:      f.MimeType,
		HasPassword:   f.PasswordHash != nil,
		CreatedAt:     f.CreatedAt.Format(time.RFC3339),
		DownloadCount: f.DownloadCount,
	}
	if f.ExpiresAt != nil {
		meta.ExpiresAt = f.ExpiresAt.Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}

func (s *Server) FileDownload(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"missing id"}`, http.StatusBadRequest)
		return
	}

	f, ok := s.liveFile(w, id)
	if !ok {
		return
	}

	if f.PasswordHash != nil {
		var body struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"password required"}`, http.StatusUnauthorized)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*f.PasswordHash), []byte(body.Password)); err != nil {
			http.Error(w, `{"error":"incorrect password"}`, http.StatusUnauthorized)
			return
		}
	}

	file, err := s.store.Open(f.StoragePath)
	if err != nil {
		http.Error(w, `{"error":"file not found on disk"}`, http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", f.MimeType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+f.Filename+`"`)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", f.Size))

	if _, err := io.Copy(w, file); err != nil {
		log.Printf("error streaming file: %v", err)
		return
	}

	go s.db.IncrementDownload(id)
}

// FileDelete removes a file (anyone with the link may delete). If the file is
// password protected, the correct password must be supplied in the JSON body.
func (s *Server) FileDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"missing id"}`, http.StatusBadRequest)
		return
	}

	f, err := s.db.GetFile(id)
	if err != nil {
		http.Error(w, `{"error":"file not found"}`, http.StatusNotFound)
		return
	}

	if f.PasswordHash != nil {
		var body struct {
			Password string `json:"password"`
		}
		json.NewDecoder(r.Body).Decode(&body) // empty password if no/invalid body
		if err := bcrypt.CompareHashAndPassword([]byte(*f.PasswordHash), []byte(body.Password)); err != nil {
			http.Error(w, `{"error":"incorrect password"}`, http.StatusUnauthorized)
			return
		}
	}

	s.store.Delete(f.StoragePath)
	if err := s.db.DeleteFile(id); err != nil {
		log.Printf("failed to delete file %s: %v", id, err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) FileQR(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"missing id"}`, http.StatusBadRequest)
		return
	}

	if _, ok := s.liveFile(w, id); !ok {
		return
	}

	shareURL := s.baseURL + "/f/" + id
	png, err := qrcode.Encode(shareURL, qrcode.Medium, 256)
	if err != nil {
		log.Printf("failed to generate QR: %v", err)
		http.Error(w, `{"error":"qr generation failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(png)))
	w.Write(png)
}

// liveFile fetches a file by id, treating expired files as gone. If the file is
// missing or expired it writes the appropriate error response and returns false.
// Expired files are cleaned up lazily (file removed from disk and record deleted).
func (s *Server) liveFile(w http.ResponseWriter, id string) (*model.File, bool) {
	f, err := s.db.GetFile(id)
	if err != nil {
		http.Error(w, `{"error":"file not found"}`, http.StatusNotFound)
		return nil, false
	}

	if f.ExpiresAt != nil && time.Now().After(*f.ExpiresAt) {
		s.store.Delete(f.StoragePath)
		if err := s.db.DeleteFile(id); err != nil {
			log.Printf("failed to delete expired file %s: %v", id, err)
		}
		http.Error(w, `{"error":"file has expired"}`, http.StatusGone)
		return nil, false
	}

	return f, true
}

func generateSlug(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == 0 {
			return -1
		}
		return r
	}, name)
	if name == "." || name == "" {
		return "file"
	}
	return name
}

