package tui

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"scottgcooper-cbz-webp-converter/fileops"

	"github.com/chai2010/webp"
)

// SilentArchive creates archives without printing to stdout
type SilentArchive struct{}

// CreateSilentZipArchive creates a ZIP archive without console output
func CreateSilentZipArchive(sourceDir, archivePath string) error {
	// Create the archive file
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	// Walk through the directory and add files to archive
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories in the archive
		if info.IsDir() {
			return nil
		}

		// Create relative path for archive
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Check if file is an image
		if fileops.IsImageFile(path) {
			// Convert to WebP and add to archive
			err = addImageAsWebPToZipSilent(zipWriter, path, relPath)
			if err != nil {
				return err
			}
		} else {
			// Add non-image file as-is
			err = addFileToZipSilent(zipWriter, path, relPath)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// addImageAsWebPToZipSilent converts an image to WebP and adds it to the ZIP (silent)
func addImageAsWebPToZipSilent(zipWriter *zip.Writer, filePath, zipPath string) error {
	// Open the input file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Convert to WebP in memory
	img, _, err := fileops.DecodeImage(file)
	if err != nil {
		return err
	}

	// Create WebP filename
	webpPath := strings.TrimSuffix(zipPath, filepath.Ext(zipPath)) + ".webp"

	// Create zip file header for WebP
	header := &zip.FileHeader{
		Name:   webpPath,
		Method: zip.Deflate,
	}

	// Create writer for this file in the zip
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// Encode as WebP directly to zip
	err = webp.Encode(writer, img, &webp.Options{Quality: 80})
	return err
}

// addFileToZipSilent adds a non-image file to the ZIP archive (silent)
func addFileToZipSilent(zipWriter *zip.Writer, filePath, zipPath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create zip file header
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Set the name in the zip
	header.Name = zipPath
	header.Method = zip.Deflate

	// Create writer for this file in the zip
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// Copy file contents to zip
	_, err = io.Copy(writer, file)
	return err
}
