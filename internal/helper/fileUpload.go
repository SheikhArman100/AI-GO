package helper

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// UploadedFile represents the result of a file upload operation
type UploadedFile struct {
	Path         string // Full path to the file on disk
	WebPath      string // Web-accessible path for the file
	DiskType     string // Type of storage (e.g., "local", "s3")
	OriginalName string // Original filename from the upload
	ModifiedName string // New filename used for storage
}

// IsImageFile checks if the provided file header represents an image based on its content type.
func IsImageFile(fileHeader *multipart.FileHeader) bool {
	if fileHeader == nil {
		return false
	}

	// Open the file to read its content type
	file, err := fileHeader.Open()
	if err != nil {
		// Cannot open file, so cannot determine if it's an image
		return false
	}
	defer file.Close()

	// Read the first 512 bytes to determine the content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		// Cannot read file, so cannot determine if it's an image
		return false
	}

	// Reset the read pointer to the beginning of the file
	// This is important if the file needs to be read again (e.g., for saving)
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return false
	}

	contentType := http.DetectContentType(buffer)

	switch contentType {
	case "image/jpeg", "image/jpg":
		return true
	case "image/png":
		return true
	case "image/gif":
		return true
	case "image/webp":
		return true
	default:
		return false
	}
}

// UploadFile saves the uploaded file to the specified directory with a unique name.
// It returns an UploadedFile struct containing file information or an error.
func UploadFileLocally(fileHeader *multipart.FileHeader, uploadDir string) (*UploadedFile, error) {
	if fileHeader == nil {
		return nil, fmt.Errorf("file header is nil")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create the upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory %s: %w", uploadDir, err)
	}

	// Generate a unique filename to prevent overwrites and ensure security
	extension := filepath.Ext(fileHeader.Filename)
	newFileName := uuid.New().String() + strings.ToLower(extension)
	dstPath := filepath.Join(uploadDir, newFileName)

	// Create the destination file
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file %s: %w", dstPath, err)
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err = io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to copy uploaded file to %s: %w", dstPath, err)
	}

	// Generate web-accessible path
	// Extract the relative path from uploadDir by removing the leading './' if present
	webDir := uploadDir
	if strings.HasPrefix(webDir, ".") {
		webDir = webDir[1:] // Remove the leading '.'
	}
	// Ensure the path starts with a slash
	if !strings.HasPrefix(webDir, "/") {
		webDir = "/" + webDir
	}
	webPath := filepath.Join(webDir, newFileName)
	// Convert backslashes to forward slashes for web URLs
	webPath = strings.ReplaceAll(webPath, "\\", "/")

	// Return the uploaded file information
	return &UploadedFile{
		Path:         dstPath,
		WebPath:      webPath,
		DiskType:     "local", // Default to local storage
		OriginalName: fileHeader.Filename,
		ModifiedName: newFileName,
	}, nil
}
