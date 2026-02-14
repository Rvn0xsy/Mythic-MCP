// Package filestore provides a file-vending system that temporarily stores
// Mythic artifacts (payloads, screenshots, downloaded files) and serves them
// via one-time-use download tokens. This enables AI agents to consume files
// through simple HTTP URLs instead of base64-encoded blobs.
package filestore

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileType categorizes the stored file for directory organization.
type FileType string

const (
	FileTypePayload    FileType = "payloads"
	FileTypeScreenshot FileType = "screenshots"
	FileTypeDownload   FileType = "downloads"
)

// VendedFile holds metadata about a file ready for one-time download.
type VendedFile struct {
	ID          string    `json:"id"`
	Path        string    `json:"-"` // disk path (not exposed)
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	FileType    FileType  `json:"file_type"`
	Token       string    `json:"token"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Downloaded  bool      `json:"-"`
}

// VendedFileResponse is returned to the MCP tool caller.
type VendedFileResponse struct {
	DownloadURL      string `json:"download_url"`
	Filename         string `json:"filename"`
	Size             int64  `json:"size"`
	ContentType      string `json:"content_type"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
	FileID           string `json:"file_id"`
}

// Config holds file store configuration.
type Config struct {
	Enabled          bool
	StoragePath      string
	TokenExpiry      time.Duration
	MaxFileSizeMB    int
	CleanupInterval  time.Duration
	BaseURL          string // e.g. "http://169.254.32.156:3333"
	Secret           []byte // HMAC signing secret
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		// fallback — should never happen
		secret = []byte("mythic-mcp-file-vending-secret-key!")
	}
	return &Config{
		Enabled:         true,
		StoragePath:     "/tmp/mythic-files",
		TokenExpiry:     5 * time.Minute,
		MaxFileSizeMB:   100,
		CleanupInterval: 60 * time.Second,
		BaseURL:         "http://localhost:3333",
		Secret:          secret,
	}
}

// FileStore manages temporary file storage and one-time download tokens.
type FileStore struct {
	cfg   *Config
	mu    sync.RWMutex
	files map[string]*VendedFile // keyed by file ID
	stop  chan struct{}
}

// New creates a new FileStore, creates storage directories, cleans stale
// files from any prior run, and starts the background cleanup goroutine.
func New(cfg *Config) (*FileStore, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Create storage directories
	for _, subdir := range []FileType{FileTypePayload, FileTypeScreenshot, FileTypeDownload} {
		dir := filepath.Join(cfg.StoragePath, string(subdir))
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create storage dir %s: %w", dir, err)
		}
	}

	fs := &FileStore{
		cfg:   cfg,
		files: make(map[string]*VendedFile),
		stop:  make(chan struct{}),
	}

	// Clean stale files from previous runs
	fs.cleanAll()

	// Start background cleanup
	go fs.cleanupLoop()

	log.Printf("[filestore] initialized storage at %s (token TTL=%s, cleanup=%s)",
		cfg.StoragePath, cfg.TokenExpiry, cfg.CleanupInterval)

	return fs, nil
}

// Close stops the background goroutine and cleans all files.
func (fs *FileStore) Close() {
	close(fs.stop)
	fs.cleanAll()
}

// StoreFile persists raw bytes to disk and returns a VendedFileResponse
// containing the one-time download URL.
func (fs *FileStore) StoreFile(data []byte, filename string, fileType FileType, contentType string) (*VendedFileResponse, error) {
	if !fs.cfg.Enabled {
		return nil, fmt.Errorf("file vending is disabled")
	}

	maxSize := int64(fs.cfg.MaxFileSizeMB) * 1024 * 1024
	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("file exceeds maximum size of %d MB", fs.cfg.MaxFileSizeMB)
	}

	// Generate unique ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("failed to generate file ID: %w", err)
	}
	fileID := hex.EncodeToString(idBytes)

	// Generate token
	token := fs.generateToken(fileID)

	// Write file to disk
	dir := filepath.Join(fs.cfg.StoragePath, string(fileType))
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".bin"
	}
	diskPath := filepath.Join(dir, fileID+ext)

	if err := os.WriteFile(diskPath, data, 0600); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	now := time.Now()
	vf := &VendedFile{
		ID:          fileID,
		Path:        diskPath,
		Filename:    filename,
		ContentType: contentType,
		Size:        int64(len(data)),
		FileType:    fileType,
		Token:       token,
		CreatedAt:   now,
		ExpiresAt:   now.Add(fs.cfg.TokenExpiry),
	}

	fs.mu.Lock()
	fs.files[fileID] = vf
	fs.mu.Unlock()

	expirySeconds := int(fs.cfg.TokenExpiry.Seconds())

	resp := &VendedFileResponse{
		DownloadURL:      fmt.Sprintf("%s/download/%s?token=%s", fs.cfg.BaseURL, fileID, token),
		Filename:         filename,
		Size:             vf.Size,
		ContentType:      contentType,
		ExpiresInSeconds: expirySeconds,
		FileID:           fileID,
	}

	log.Printf("[filestore] stored %s (%s, %d bytes, expires %s)",
		filename, fileType, len(data), vf.ExpiresAt.Format(time.RFC3339))

	return resp, nil
}

// ServeDownload is the HTTP handler for GET /download/{file_id}?token=...
// It validates the token, serves the file, then immediately deletes it.
func (fs *FileStore) ServeDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract file ID from path: /download/{file_id}
	fileID := filepath.Base(r.URL.Path)
	if fileID == "" || fileID == "." || fileID == "download" {
		http.Error(w, `{"error":"file ID required"}`, http.StatusBadRequest)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, `{"error":"token required"}`, http.StatusUnauthorized)
		return
	}

	fs.mu.Lock()
	vf, exists := fs.files[fileID]
	if !exists {
		fs.mu.Unlock()
		http.Error(w, `{"error":"file not found"}`, http.StatusNotFound)
		return
	}

	// Check expiration
	if time.Now().After(vf.ExpiresAt) {
		// Clean up expired file
		delete(fs.files, fileID)
		fs.mu.Unlock()
		os.Remove(vf.Path) //nolint:errcheck
		http.Error(w, `{"error":"token expired"}`, http.StatusUnauthorized)
		return
	}

	// Check one-time use
	if vf.Downloaded {
		delete(fs.files, fileID)
		fs.mu.Unlock()
		os.Remove(vf.Path) //nolint:errcheck
		http.Error(w, `{"error":"token already used"}`, http.StatusGone)
		return
	}

	// Validate token
	expectedToken := fs.generateToken(fileID)
	if !hmac.Equal([]byte(token), []byte(expectedToken)) {
		fs.mu.Unlock()
		http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
		return
	}

	// Mark as downloaded immediately (before releasing lock)
	vf.Downloaded = true
	delete(fs.files, fileID)
	fs.mu.Unlock()

	// Read file
	data, err := os.ReadFile(vf.Path)
	if err != nil {
		http.Error(w, `{"error":"file read error"}`, http.StatusInternalServerError)
		return
	}

	// Serve file
	w.Header().Set("Content-Type", vf.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, vf.Filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data) //nolint:errcheck

	// Delete file after successful serve
	os.Remove(vf.Path) //nolint:errcheck

	log.Printf("[filestore] served and deleted %s (%s, %d bytes)",
		vf.Filename, vf.FileType, len(data))
}

// generateToken creates an HMAC-SHA256 token for a file ID.
func (fs *FileStore) generateToken(fileID string) string {
	mac := hmac.New(sha256.New, fs.cfg.Secret)
	mac.Write([]byte(fileID))
	return hex.EncodeToString(mac.Sum(nil))
}

// cleanupLoop periodically removes expired files.
func (fs *FileStore) cleanupLoop() {
	ticker := time.NewTicker(fs.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-fs.stop:
			return
		case <-ticker.C:
			fs.cleanExpired()
		}
	}
}

// cleanExpired removes files whose tokens have expired.
func (fs *FileStore) cleanExpired() {
	now := time.Now()
	var removed int

	fs.mu.Lock()
	for id, vf := range fs.files {
		if now.After(vf.ExpiresAt) {
			os.Remove(vf.Path) //nolint:errcheck
			delete(fs.files, id)
			removed++
		}
	}
	fs.mu.Unlock()

	if removed > 0 {
		log.Printf("[filestore] cleaned %d expired files", removed)
	}
}

// cleanAll removes all tracked files and empties the storage directories.
func (fs *FileStore) cleanAll() {
	fs.mu.Lock()
	for id, vf := range fs.files {
		os.Remove(vf.Path) //nolint:errcheck
		delete(fs.files, id)
	}
	fs.mu.Unlock()

	// Also clean any orphaned files from previous runs
	for _, subdir := range []FileType{FileTypePayload, FileTypeScreenshot, FileTypeDownload} {
		dir := filepath.Join(fs.cfg.StoragePath, string(subdir))
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				os.Remove(filepath.Join(dir, entry.Name())) //nolint:errcheck
			}
		}
	}
}

// Stats returns current file store statistics.
func (fs *FileStore) Stats() map[string]interface{} {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var totalSize int64
	for _, vf := range fs.files {
		totalSize += vf.Size
	}

	return map[string]interface{}{
		"tracked_files": len(fs.files),
		"total_size":    totalSize,
		"storage_path":  fs.cfg.StoragePath,
		"enabled":       fs.cfg.Enabled,
	}
}

// StatusJSON returns a JSON-serializable status for /download/status.
func (fs *FileStore) StatusJSON() ([]byte, error) {
	return json.Marshal(fs.Stats())
}
