package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"scottgcooper-cbz-webp-converter/fileops"

	"github.com/chai2010/webp"
)

// ArchiveType represents the type of archive to create
type ArchiveType string

const (
	ZIP  ArchiveType = "zip"
	RAR  ArchiveType = "rar"
	Z7   ArchiveType = "7z"
	CBZ  ArchiveType = "cbz"
	CBR  ArchiveType = "cbr"
	CB7Z ArchiveType = "cb7z"
)

// CreateArchive creates an archive for the given directory
func CreateArchive(sourceDir, archivePath string, archiveType ArchiveType) error {
	switch strings.ToLower(string(archiveType)) {
	case "cbz", "zip":
		return CreateZipArchive(sourceDir, archivePath)
	case "cbr", "rar":
		return CreateRarArchive(sourceDir, archivePath)
	case "cb7z", "7z":
		return Create7zArchive(sourceDir, archivePath)
	default:
		// Default to ZIP for unknown formats
		fmt.Printf("Unknown format '%s', defaulting to ZIP\n", archiveType)
		return CreateZipArchive(sourceDir, archivePath)
	}
}

// CreateZipArchive creates a ZIP archive with WebP converted images
func CreateZipArchive(sourceDir, archivePath string) error {
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
			err = addImageAsWebPToZip(zipWriter, path, relPath)
			if err != nil {
				fmt.Printf("Error converting and adding %s: %v\n", path, err)
				return err
			}
		} else {
			// Add non-image file as-is
			err = addFileToZip(zipWriter, path, relPath)
			if err != nil {
				return err
			}
		}

		fmt.Printf("  Added to ZIP: %s\n", relPath)
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("Created ZIP: %s\n", archivePath)
	return nil
}

// CreateRarArchive creates a RAR archive using the rar command
func CreateRarArchive(sourceDir, archivePath string) error {
	// Check if rar command is available
	_, err := exec.LookPath("rar")
	if err != nil {
		return fmt.Errorf("rar command not found. Please install WinRAR or RAR for Linux/Mac")
	}

	// First create a temporary ZIP with converted images
	tempZipPath := archivePath + ".temp.zip"
	err = CreateZipArchive(sourceDir, tempZipPath)
	if err != nil {
		return err
	}
	defer os.Remove(tempZipPath) // Clean up temp file

	// Convert ZIP to RAR using rar command
	cmd := exec.Command("rar", "a", "-ep1", archivePath, tempZipPath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create RAR archive: %v", err)
	}

	fmt.Printf("Created RAR: %s\n", archivePath)
	return nil
}

// Create7zArchive creates a 7Z archive using the 7z command
func Create7zArchive(sourceDir, archivePath string) error {
	// Check if 7z command is available
	_, err := exec.LookPath("7z")
	if err != nil {
		return fmt.Errorf("7z command not found. Please install p7zip")
	}

	// First create a temporary ZIP with converted images
	tempZipPath := archivePath + ".temp.zip"
	err = CreateZipArchive(sourceDir, tempZipPath)
	if err != nil {
		return err
	}
	defer os.Remove(tempZipPath) // Clean up temp file

	// Convert ZIP to 7Z using 7z command
	cmd := exec.Command("7z", "a", "-t7z", archivePath, tempZipPath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create 7Z archive: %v", err)
	}

	fmt.Printf("Created 7Z: %s\n", archivePath)
	return nil
}

// addImageAsWebPToZip converts an image to WebP and adds it to the ZIP
func addImageAsWebPToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
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

	fmt.Printf("  Converted %s -> %s (%s)\n", filepath.Base(filePath), filepath.Base(webpPath), format)
	return nil
}

// addFileToZip adds a non-image file to the ZIP archive
func addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
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
