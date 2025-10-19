package fileops

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
)

// IsImageFile checks if a file is an image based on its extension
func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

// DecodeImage attempts to decode an image file using various decoders
func DecodeImage(file *os.File) (image.Image, string, error) {
	var img image.Image
	var format string

	// Try JPEG first
	file.Seek(0, 0) // Reset file position
	img, err := jpeg.Decode(file)
	if err == nil {
		return img, "JPEG", nil
	}

	// Try PNG if JPEG failed
	file.Seek(0, 0)
	img, err = png.Decode(file)
	if err == nil {
		return img, "PNG", nil
	}

	// Try GIF if PNG failed
	file.Seek(0, 0)
	img, err = gif.Decode(file)
	if err == nil {
		return img, "GIF", nil
	}

	// If all failed, try generic decode
	file.Seek(0, 0)
	img, format, err = image.Decode(file)
	if err != nil {
		return nil, "", err
	}

	return img, format, nil
}

// ConvertToWebP converts an image to WebP format with specified quality
func ConvertToWebP(img image.Image, quality float32) ([]byte, error) {
	var buf bytes.Buffer
	err := webp.Encode(&buf, img, &webp.Options{Quality: quality})
	return buf.Bytes(), err
}
