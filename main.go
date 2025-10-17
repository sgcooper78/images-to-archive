package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"scottgcooper-cbz-webp-converter/archive"
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
		parentDir := filepath.Dir(dirPath)
		dirName := filepath.Base(dirPath)
		archivePath := filepath.Join(parentDir, dirName+"."+*format)

		err = archive.CreateArchive(dirPath, archivePath, archive.ArchiveType(*format))
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
