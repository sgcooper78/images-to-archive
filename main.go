package main

import (
	"archive/zip"
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
)

var deleteOriginal = flag.Bool("delete-original-files", false, "Delete original image files after conversion")
var format = flag.String("format", "cbz", "Archive format (cbz/zip, cbr/rar, cb7z/7z)")

func main() {
	flag.Parse()

	var dirPath string

	// Get remaining arguments after flag parsing
	args := flag.Args()
	if len(args) > 0 {
		dirPath = args[0]
	}

	if dirPath == "" {
		// Prompt user for directory path
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter directory path: ")
		input, _ := reader.ReadString('\n')
		dirPath = strings.TrimSpace(input)
	}

	fmt.Printf("Processing directory recursively: %s\n", dirPath)
	fmt.Printf("Archive format: %s\n", *format)
	if *deleteOriginal {
		fmt.Println("WARNING: Original files will be deleted after conversion!")
	}
	fmt.Println("=====================================")

	// Process directory recursively
	err := processDirectoryRecursively(dirPath)
	if err != nil {
		fmt.Printf("Error processing directory: %v\n", err)
	}

	fmt.Println("\nAll directories processed!")
}

func processDirectoryRecursively(rootPath string) error {
	// First, collect all directories to process
	var directories []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process directories
		if info.IsDir() {
			// Skip the root directory itself
			if path != rootPath {
				directories = append(directories, path)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Process each directory
	for _, dirPath := range directories {
		fmt.Printf("\nProcessing directory: %s\n", dirPath)
		fmt.Println("------------------------")

		// Create archive for this directory
		err = createArchiveForDirectory(dirPath)
		if err != nil {
			fmt.Printf("Error creating %s for %s: %v\n", strings.ToUpper(*format), dirPath, err)
			continue
		}

		// Delete the entire directory if flag is set
		if *deleteOriginal {
			err = os.RemoveAll(dirPath)
			if err != nil {
				fmt.Printf("Error deleting directory %s: %v\n", dirPath, err)
			} else {
				fmt.Printf("Deleted directory: %s\n", dirPath)
			}
		}
	}

	return nil
}

func createArchiveForDirectory(sourceDir string) error {
	// Get the parent directory and create archive filename
	parentDir := filepath.Dir(sourceDir)
	dirName := filepath.Base(sourceDir)
	archivePath := filepath.Join(parentDir, dirName+"."+*format)

	// Determine archive type and create accordingly
	switch strings.ToLower(*format) {
	case "cbz", "zip":
		return createZipArchive(sourceDir, archivePath)
	case "cbr", "rar":
		return createRarArchive(sourceDir, archivePath)
	case "cb7z", "7z":
		return create7zArchive(sourceDir, archivePath)
	default:
		// Default to ZIP for unknown formats
		fmt.Printf("Unknown format '%s', defaulting to ZIP\n", *format)
		return createZipArchive(sourceDir, archivePath)
	}
}

func createZipArchive(sourceDir, archivePath string) error {
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
		if isImageFile(path) {
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

		fmt.Printf("  Added to %s: %s\n", strings.ToUpper(*format), relPath)
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("Created %s: %s\n", strings.ToUpper(*format), archivePath)
	return nil
}

func createRarArchive(sourceDir, archivePath string) error {
	// Check if rar command is available
	_, err := exec.LookPath("rar")
	if err != nil {
		return fmt.Errorf("rar command not found. Please install WinRAR or RAR for Linux/Mac")
	}

	// First create a temporary ZIP with converted images
	tempZipPath := archivePath + ".temp.zip"
	err = createZipArchive(sourceDir, tempZipPath)
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

	fmt.Printf("Created %s: %s\n", strings.ToUpper(*format), archivePath)
	return nil
}

func create7zArchive(sourceDir, archivePath string) error {
	// Check if 7z command is available
	_, err := exec.LookPath("7z")
	if err != nil {
		return fmt.Errorf("7z command not found. Please install p7zip")
	}

	// First create a temporary ZIP with converted images
	tempZipPath := archivePath + ".temp.zip"
	err = createZipArchive(sourceDir, tempZipPath)
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

	fmt.Printf("Created %s: %s\n", strings.ToUpper(*format), archivePath)
	return nil
}

func addImageAsWebPToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	// Open the input file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Convert to WebP in memory
	img, format, err := decodeImage(file)
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

func decodeImage(file *os.File) (image.Image, string, error) {
	// Try to decode with different decoders
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
		return nil, "", fmt.Errorf("unable to decode image: %v", err)
	}

	return img, format, nil
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

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
