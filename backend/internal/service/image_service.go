package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"

	_ "golang.org/x/image/webp" // WebP decoder
)

// ImageService handles image processing
type ImageService interface {
	// Process processes an uploaded image: validates, compresses, and converts to Base64
	Process(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*ImageData, error)

	// ValidateType validates the image MIME type
	ValidateType(mimeType string) error

	// ValidateSize validates the image size
	ValidateSize(size int64) error
}

// ImageServiceConfig contains image processing configuration
type ImageServiceConfig struct {
	MaxSize      int64    // Maximum size in bytes
	Quality      int      // JPEG quality (1-100)
	MaxDimension int      // Maximum width/height in pixels
	AllowedTypes []string // Allowed MIME types
}

// imageService implements ImageService
type imageService struct {
	config ImageServiceConfig
}

// NewImageService creates a new ImageService
func NewImageService(config ImageServiceConfig) ImageService {
	return &imageService{config: config}
}

// ErrInvalidImageFormat is returned when the image format is not supported
type ErrInvalidImageFormat struct {
	MimeType string
}

func (e *ErrInvalidImageFormat) Error() string {
	return "invalid image format: " + e.MimeType
}

// ErrImageTooLarge is returned when the image exceeds the size limit
type ErrImageTooLarge struct {
	Size    int64
	MaxSize int64
}

func (e *ErrImageTooLarge) Error() string {
	return fmt.Sprintf("image too large: %d bytes (max: %d bytes)", e.Size, e.MaxSize)
}

// Process processes an uploaded image
func (s *imageService) Process(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*ImageData, error) {
	// Validate size
	if err := s.ValidateSize(header.Size); err != nil {
		return nil, err
	}

	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect MIME type from content
	mimeType := detectMimeType(data)
	if err := s.ValidateType(mimeType); err != nil {
		return nil, err
	}

	// Decode image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize if necessary
	img = s.resizeIfNeeded(img)

	// Compress and encode
	compressed, err := s.compress(img, mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to compress image: %w", err)
	}

	// Convert to Base64 data URL
	base64URL := s.toBase64URL(compressed, mimeType)

	return &ImageData{
		ID:       uuid.New().String(),
		Filename: header.Filename,
		MimeType: mimeType,
		Base64:   base64URL,
	}, nil
}

// ValidateType validates the image MIME type
func (s *imageService) ValidateType(mimeType string) error {
	for _, allowed := range s.config.AllowedTypes {
		if mimeType == allowed {
			return nil
		}
	}
	return &ErrInvalidImageFormat{MimeType: mimeType}
}

// ValidateSize validates the image size
func (s *imageService) ValidateSize(size int64) error {
	if size > s.config.MaxSize {
		return &ErrImageTooLarge{Size: size, MaxSize: s.config.MaxSize}
	}
	return nil
}

// resizeIfNeeded resizes the image if it exceeds the maximum dimension
func (s *imageService) resizeIfNeeded(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	maxDim := s.config.MaxDimension
	if maxDim <= 0 {
		maxDim = 2048
	}

	if width <= maxDim && height <= maxDim {
		return img
	}

	// Calculate new dimensions maintaining aspect ratio
	if width > height {
		return imaging.Resize(img, maxDim, 0, imaging.Lanczos)
	}
	return imaging.Resize(img, 0, maxDim, imaging.Lanczos)
}

// compress compresses the image to JPEG format
func (s *imageService) compress(img image.Image, mimeType string) ([]byte, error) {
	var buf bytes.Buffer

	quality := s.config.Quality
	if quality <= 0 || quality > 100 {
		quality = 80
	}

	// For PNG, keep as PNG if it has transparency, otherwise convert to JPEG
	if mimeType == "image/png" {
		// Encode as PNG
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	// Encode as JPEG for other formats
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// toBase64URL converts image data to a data URL
func (s *imageService) toBase64URL(data []byte, mimeType string) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
}

// detectMimeType detects the MIME type from file content
func detectMimeType(data []byte) string {
	if len(data) < 12 {
		return ""
	}

	// PNG magic number
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}

	// JPEG magic number
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}

	// WebP magic number (RIFF....WEBP)
	if string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "image/webp"
	}

	return ""
}

// IsAllowedImageType checks if the MIME type is in the allowed list
func IsAllowedImageType(mimeType string, allowedTypes []string) bool {
	mimeType = strings.ToLower(mimeType)
	for _, allowed := range allowedTypes {
		if mimeType == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}
