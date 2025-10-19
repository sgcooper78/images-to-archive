package tui

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"scottgcooper-cbz-webp-converter/fileops"

	"github.com/chai2010/webp"
)

// ProgressArchive handles archive creation with progress reporting
type ProgressArchive struct {
	progressCallback func(ProgressMsg)
	fileCallback     func(FileProcessedMsg)
}

// NewProgressArchive creates a new progress-aware archive handler
func NewProgressArchive(progressCallback func(ProgressMsg), fileCallback func(FileProcessedMsg)) *ProgressArchive {
	return &ProgressArchive{
		progressCallback: progressCallback,
		fileCallback:     fileCallback,
	}
}

// CreateArchiveWithProgress creates an archive with detailed progress reporting
func (pa *ProgressArchive) CreateArchiveWithProgress(sourceDir, archivePath, format string) error {
	// Count total files first
	totalFiles := 0
	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalFiles++
		}
		return nil
	})

	// Send initial progress
	pa.progressCallback(ProgressMsg{
		CurrentDir:     filepath.Base(sourceDir),
		CurrentDirNum:  1, // This will be updated by the caller
		TotalDirs:      1, // This will be updated by the caller
		ProcessedFiles: 0,
		TotalFiles:     totalFiles,
		Message:        fmt.Sprintf("Starting conversion of %s...", filepath.Base(sourceDir)),
	})

	// Create the archive file
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	processedFiles := 0

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
			err = pa.addImageAsWebPToZip(zipWriter, path, relPath)
			if err != nil {
				pa.progressCallback(ProgressMsg{
					CurrentDir:     filepath.Base(sourceDir),
					CurrentDirNum:  1,
					TotalDirs:      1,
					ProcessedFiles: processedFiles,
					TotalFiles:     totalFiles,
					Message:        fmt.Sprintf("Error converting %s: %v", filepath.Base(path), err),
				})
				return err
			}
		} else {
			// Add non-image file as-is
			err = pa.addFileToZip(zipWriter, path, relPath)
			if err != nil {
				return err
			}
		}

		processedFiles++

		// Send progress update
		pa.progressCallback(ProgressMsg{
			CurrentDir:     filepath.Base(sourceDir),
			CurrentDirNum:  1,
			TotalDirs:      1,
			ProcessedFiles: processedFiles,
			TotalFiles:     totalFiles,
			Message:        fmt.Sprintf("Processing %s...", filepath.Base(path)),
		})

		return nil
	})

	if err != nil {
		return err
	}

	// Send completion message
	pa.progressCallback(ProgressMsg{
		CurrentDir:     filepath.Base(sourceDir),
		CurrentDirNum:  1,
		TotalDirs:      1,
		ProcessedFiles: processedFiles,
		TotalFiles:     totalFiles,
		Message:        fmt.Sprintf("Completed %s", filepath.Base(sourceDir)),
	})

	return nil
}

// addImageAsWebPToZip converts an image to WebP and adds it to the ZIP
func (pa *ProgressArchive) addImageAsWebPToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	// Open the input file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Convert to WebP in memory
	img, format, err := fileops.DecodeImage(file)
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
	if err != nil {
		return err
	}

	// Send file processed message
	pa.fileCallback(FileProcessedMsg{
		FileName:    filepath.Base(filePath),
		FileType:    format,
		ConvertedTo: "WebP",
	})

	return nil
}

// addFileToZip adds a non-image file to the ZIP archive
func (pa *ProgressArchive) addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
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
