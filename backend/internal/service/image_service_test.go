package service

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"testing"
)

// mockFileHeader creates a mock multipart.FileHeader
type mockFile struct {
	*bytes.Reader
}

func (m *mockFile) Close() error {
	return nil
}

func createTestPNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func createTestJPEG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
	return buf.Bytes()
}

func TestImageService_ValidateType(t *testing.T) {
	service := NewImageService(ImageServiceConfig{
		AllowedTypes: []string{"image/png", "image/jpeg", "image/webp"},
	})

	tests := []struct {
		mimeType string
		valid    bool
	}{
		{"image/png", true},
		{"image/jpeg", true},
		{"image/webp", true},
		{"image/gif", false},
		{"text/plain", false},
	}

	for _, tt := range tests {
		err := service.ValidateType(tt.mimeType)
		if tt.valid && err != nil {
			t.Errorf("expected %s to be valid, got error: %v", tt.mimeType, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("expected %s to be invalid, got no error", tt.mimeType)
		}
	}
}

func TestImageService_ValidateSize(t *testing.T) {
	service := NewImageService(ImageServiceConfig{
		MaxSize: 1024 * 1024, // 1MB
	})

	tests := []struct {
		size  int64
		valid bool
	}{
		{500 * 1024, true},      // 500KB
		{1024 * 1024, true},     // 1MB exactly
		{2 * 1024 * 1024, false}, // 2MB
	}

	for _, tt := range tests {
		err := service.ValidateSize(tt.size)
		if tt.valid && err != nil {
			t.Errorf("expected size %d to be valid, got error: %v", tt.size, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("expected size %d to be invalid, got no error", tt.size)
		}
	}
}

func TestImageService_Process_PNG(t *testing.T) {
	service := NewImageService(ImageServiceConfig{
		MaxSize:      10 * 1024 * 1024,
		Quality:      80,
		MaxDimension: 2048,
		AllowedTypes: []string{"image/png", "image/jpeg"},
	})

	pngData := createTestPNG(100, 100)

	file := &mockFile{bytes.NewReader(pngData)}
	header := &multipart.FileHeader{
		Filename: "test.png",
		Size:     int64(len(pngData)),
	}

	ctx := context.Background()
	result, err := service.Process(ctx, file, header)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ID == "" {
		t.Error("expected ID to be set")
	}
	if result.Filename != "test.png" {
		t.Errorf("expected filename test.png, got %s", result.Filename)
	}
	if result.MimeType != "image/png" {
		t.Errorf("expected MIME type image/png, got %s", result.MimeType)
	}
	if result.Base64 == "" {
		t.Error("expected Base64 to be set")
	}
	if !bytes.HasPrefix([]byte(result.Base64), []byte("data:image/png;base64,")) {
		t.Error("expected Base64 to be a data URL")
	}
}

func TestImageService_Process_JPEG(t *testing.T) {
	service := NewImageService(ImageServiceConfig{
		MaxSize:      10 * 1024 * 1024,
		Quality:      80,
		MaxDimension: 2048,
		AllowedTypes: []string{"image/png", "image/jpeg"},
	})

	jpegData := createTestJPEG(100, 100)

	file := &mockFile{bytes.NewReader(jpegData)}
	header := &multipart.FileHeader{
		Filename: "test.jpg",
		Size:     int64(len(jpegData)),
	}

	ctx := context.Background()
	result, err := service.Process(ctx, file, header)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.MimeType != "image/jpeg" {
		t.Errorf("expected MIME type image/jpeg, got %s", result.MimeType)
	}
}

func TestImageService_Process_LargeImage(t *testing.T) {
	service := NewImageService(ImageServiceConfig{
		MaxSize:      10 * 1024 * 1024,
		Quality:      80,
		MaxDimension: 500, // Small max dimension to force resize
		AllowedTypes: []string{"image/png", "image/jpeg"},
	})

	// Create a 1000x800 image
	pngData := createTestPNG(1000, 800)

	file := &mockFile{bytes.NewReader(pngData)}
	header := &multipart.FileHeader{
		Filename: "large.png",
		Size:     int64(len(pngData)),
	}

	ctx := context.Background()
	result, err := service.Process(ctx, file, header)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// The image should have been resized
	if result.Base64 == "" {
		t.Error("expected Base64 to be set")
	}
}

func TestImageService_Process_InvalidType(t *testing.T) {
	service := NewImageService(ImageServiceConfig{
		MaxSize:      10 * 1024 * 1024,
		AllowedTypes: []string{"image/png"},
	})

	jpegData := createTestJPEG(100, 100)

	file := &mockFile{bytes.NewReader(jpegData)}
	header := &multipart.FileHeader{
		Filename: "test.jpg",
		Size:     int64(len(jpegData)),
	}

	ctx := context.Background()
	_, err := service.Process(ctx, file, header)
	if err == nil {
		t.Error("expected error for invalid type")
	}

	if _, ok := err.(*ErrInvalidImageFormat); !ok {
		t.Errorf("expected ErrInvalidImageFormat, got %T", err)
	}
}

func TestImageService_Process_TooLarge(t *testing.T) {
	service := NewImageService(ImageServiceConfig{
		MaxSize:      100, // Very small limit
		AllowedTypes: []string{"image/png"},
	})

	pngData := createTestPNG(100, 100)

	file := &mockFile{bytes.NewReader(pngData)}
	header := &multipart.FileHeader{
		Filename: "test.png",
		Size:     int64(len(pngData)),
	}

	ctx := context.Background()
	_, err := service.Process(ctx, file, header)
	if err == nil {
		t.Error("expected error for oversized image")
	}

	if _, ok := err.(*ErrImageTooLarge); !ok {
		t.Errorf("expected ErrImageTooLarge, got %T", err)
	}
}

func TestDetectMimeType(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{"PNG", createTestPNG(10, 10), "image/png"},
		{"JPEG", createTestJPEG(10, 10), "image/jpeg"},
		{"Empty", []byte{}, ""},
		{"Unknown", []byte("not an image"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectMimeType(tt.data)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Ensure mockFile implements multipart.File
var _ multipart.File = (*mockFile)(nil)

func (m *mockFile) ReadAt(p []byte, off int64) (n int, err error) {
	return m.Reader.ReadAt(p, off)
}

func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
	return m.Reader.Seek(offset, whence)
}

var _ io.ReaderAt = (*mockFile)(nil)
var _ io.Seeker = (*mockFile)(nil)
