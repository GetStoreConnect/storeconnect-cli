package theme

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
)

// AssetsDirName is the directory inside a theme that holds binary assets.
const AssetsDirName = "assets"

// LocalAsset is a binary asset read from the local theme's assets/ directory.
// Key is the path relative to assets/ (forward-slashed, may contain slashes),
// matching the server's asset key contract.
type LocalAsset struct {
	Key         string
	Path        string
	ContentType string
	ContentHash string
	Content     []byte
}

// ReadLocalAssets walks the assets/ directory inside themeDir and returns every
// file as a LocalAsset with its SHA-256 hex digest computed. A missing assets/
// directory is not an error - it simply yields no assets.
func ReadLocalAssets(themeDir string) ([]LocalAsset, error) {
	assetsDir := filepath.Join(themeDir, AssetsDirName)

	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		return nil, nil
	}

	var assets []LocalAsset

	err := filepath.Walk(assetsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(assetsDir, path)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(relPath)

		assets = append(assets, LocalAsset{
			Key:         key,
			Path:        path,
			ContentType: InferContentType(key),
			ContentHash: HashContent(content),
			Content:     content,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to read assets: %w", err)
	}

	return assets, nil
}

// HashContent returns the SHA-256 hex digest of the given bytes.
func HashContent(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// SelectChangedAssets returns the local assets that need uploading: those whose
// key is absent on the server, or whose content hash differs from the server's.
// Assets whose hash matches the server are skipped. This is a pure function so
// the selection rule is trivially unit-testable.
func SelectChangedAssets(local []LocalAsset, server []api.ThemeAsset) []LocalAsset {
	serverHashes := make(map[string]string, len(server))
	for _, a := range server {
		serverHashes[a.Key] = a.ContentHash
	}

	var changed []LocalAsset
	for _, a := range local {
		serverHash, exists := serverHashes[a.Key]
		if !exists || serverHash != a.ContentHash {
			changed = append(changed, a)
		}
	}

	return changed
}

// imageExtensions are the file extensions treated as images by the media flow.
var imageExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
	".svg":  true,
}

// contentTypes maps known image extensions to their MIME type.
var contentTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",
	".pdf":  "application/pdf",
}

// InferContentType guesses the MIME type from a filename's extension, defaulting
// to application/octet-stream for unknown extensions.
func InferContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// InferFileType returns the StoreConnect media file_type for a filename:
// "image" for known image extensions, "document" otherwise.
func InferFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if imageExtensions[ext] {
		return "image"
	}
	return "document"
}
