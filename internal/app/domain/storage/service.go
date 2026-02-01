package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service handles file storage operations
type Service struct {
	uploadDir string
	baseURL   string
}

// UploadedFile represents metadata about an uploaded file
type UploadedFile struct {
	URL          string
	FileName     string
	Size         int64
	MimeType     string
	ThumbnailURL string
}

// NewService creates a new storage service
// For now, uses local filesystem. Can be extended to support S3 later.
func NewService(uploadDir, baseURL string) *Service {
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0o750); err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %v", err))
	}

	return &Service{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

// UploadChatFile uploads a file for chat messages
func (s *Service) UploadChatFile(
	ctx context.Context,
	file io.Reader,
	fileName string,
	mimeType string,
	size int64,
	userID string,
	conversationID string,
) (*UploadedFile, error) {
	// Generate unique filename to avoid collisions
	ext := filepath.Ext(fileName)
	uniqueName := fmt.Sprintf(
		"%s_%s%s",
		time.Now().Format("20060102_150405"),
		uuid.New().String()[:8],
		ext,
	)

	// Create conversation subdirectory
	conversationDir := filepath.Join(s.uploadDir, "chat", conversationID)
	if err := os.MkdirAll(conversationDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create conversation directory: %w", err)
	}

	// Full path for the file - filepath.Clean prevents path traversal
	filePath := filepath.Join(conversationDir, filepath.Clean(uniqueName))

	// Create the file
	outFile, err := os.Create(filePath) // #nosec G304 - filename is generated internally
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if cerr := outFile.Close(); cerr != nil {
			// Log but don't override the main error
			fmt.Printf("warning: failed to close file: %v\n", cerr)
		}
	}()

	// Copy data to file
	written, err := io.Copy(outFile, file)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Build public URL
	publicURL := fmt.Sprintf("%s/uploads/chat/%s/%s", s.baseURL, conversationID, uniqueName)

	result := &UploadedFile{
		URL:      publicURL,
		FileName: fileName,
		Size:     written,
		MimeType: mimeType,
	}

	// Generate thumbnail for images
	if isImage(mimeType) {
		// TODO: Implement thumbnail generation
		// For now, use the same URL
		result.ThumbnailURL = publicURL
	}

	return result, nil
}

// DeleteFile removes a file from storage
func (s *Service) DeleteFile(ctx context.Context, fileURL string) error {
	// Extract relative path from URL
	// Example: http://localhost:8080/uploads/chat/conv-id/file.jpg -> chat/conv-id/file.jpg
	relativePath := fileURL
	if strings.HasPrefix(fileURL, s.baseURL) {
		relativePath = fileURL[len(s.baseURL)+len("/uploads/"):]
	}

	filePath := filepath.Join(s.uploadDir, relativePath)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// isImage checks if the MIME type is an image
func isImage(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

// IsAllowedMimeType checks if a file type is allowed for upload
func IsAllowedMimeType(mimeType string) bool {
	allowed := map[string]bool{
		// Images
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		// Documents
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"text/plain": true,
		// Audio (for voice messages)
		"audio/webm": true,
		"audio/ogg":  true,
		"audio/mp4":  true,
		"audio/mpeg": true,
		"audio/wav":  true,
	}

	return allowed[mimeType]
}

// FormatFileSize converts bytes to human-readable format
func FormatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
