# CBZ WebP Converter

A terminal-based tool for converting image collections into compressed CBZ/CBR/CB7Z archives with WebP optimization. This tool is perfect for comic book collectors, digital archivists, and anyone looking to efficiently compress and organize their image libraries.

## Features

**Smart Operation Modes**
- Directory Mode: Convert multiple directories into separate archives
- File Mode: Combine multiple files into a single archive  
- Auto-Detection: Automatically determines the appropriate mode based on directory contents

**Image Processing**
- WebP Compression: Converts all images to WebP format for maximum compression
- Quality Optimization: 80% quality setting for optimal size/quality balance
- Format Support: JPEG, PNG, GIF, BMP, WebP input formats
- Non-Image Files: Preserves non-image files in their original format

**Archive Formats**
- CBZ: ZIP-based comic book archives (most compatible)
- CBR: RAR-based comic book archives
- CB7Z: 7-Zip-based comic book archives

**Terminal Interface**
- Interactive TUI: Clean, responsive terminal user interface
- File Preview: Visual previews of images and file contents
- Progress Tracking: Real-time conversion progress with detailed logging
- Keyboard Navigation: Intuitive controls for efficient workflow

**Performance & Safety**
- Batch Processing: Handle multiple directories/files efficiently
- Optional Cleanup: Choose whether to delete original files after conversion
- Error Handling: Robust error handling with detailed feedback
- Memory Efficient: Processes large image collections without memory issues

## Installation

### Pre-built Binaries (Recommended)

Download the latest release from the [Releases page](https://github.com/yourusername/cbz-webp-converter/releases). Pre-built binaries are available for:

- **Linux**: AMD64 and ARM64
- **Windows**: AMD64 and ARM64  
- **macOS**: AMD64 and ARM64 (Apple Silicon)

Simply download the appropriate binary for your platform and run it directly.

### Building from Source

**Prerequisites**
- Go 1.24 or later installed on your system
- A terminal that supports TUI applications (most modern terminals work)
- WebP development libraries (for building from source):
  - **macOS**: `brew install webp pkg-config`
  - **Ubuntu/Debian**: `sudo apt-get install libwebp-dev pkg-config`
  - **Windows**: No additional setup required

**Steps**

1. Clone the repository:
```bash
git clone https://github.com/yourusername/cbz-webp-converter.git
cd cbz-webp-converter
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o cbz-converter main.go
```

4. Run the application:
```bash
./cbz-converter
```

**Cross-platform Building**

Use the provided build script to create binaries for all supported platforms:

```bash
./build.sh
```

This will create binaries in the `build/` directory for all supported platforms.

## Usage

**Basic Workflow**

1. Launch the application - The TUI will start with a clean interface
2. Enter directory path - Type the path to your image collection
3. Select items - Use the interface to select files or directories
4. Choose format - Select your preferred archive format
5. Configure options - Toggle whether to delete original files
6. Start conversion - Watch the progress as files are processed

**Controls**

Navigation:
- Up/Down arrows or j/k - Move cursor up/down
- Space - Toggle selection of current item
- a - Select all items
- n - Deselect all items
- Enter - Confirm selection and proceed

General:
- Ctrl+C or q - Quit application
- r - Restart (on completion screen)
- Tab - Toggle delete original files option

**Operation Modes**

Directory Mode:
When your selected directory contains only subdirectories, each subdirectory becomes a separate archive. Archive names match directory names. This is perfect for organizing comic book series or photo collections.

File Mode:
When your selected directory contains only files, all selected files are combined into a single archive. The archive name matches the parent directory name. This is ideal for creating single archives from loose files.

## Interface Features

**File Preview System**
The tool includes a preview system that shows:
- Image thumbnails as ASCII art
- First 10 lines of text files
- File information including size, type, and modification date
- Smart detection that automatically shows appropriate previews based on file type

**Progress Tracking**
- Real-time updates showing conversion progress
- File-by-file logging with detailed information
- Visual progress bars
- Clear error messages for any issues

## Example Use Cases

**Comic Book Collection**
```
Comics/
├── Batman-2023/
│   ├── page001.jpg
│   ├── page002.jpg
│   └── ...
├── Superman-2023/
│   ├── cover.jpg
│   ├── page001.jpg
│   └── ...
└── Wonder-Woman-2023/
    ├── page001.png
    └── ...
```
Result: Batman-2023.cbz, Superman-2023.cbz, Wonder-Woman-2023.cbz

**Photo Collection**
```
Vacation-Photos/
├── IMG_001.jpg
├── IMG_002.jpg
├── IMG_003.png
└── ...
```
Result: Vacation-Photos.cbz (single archive with all photos)

**Mixed Media Archive**
```
Project-Files/
├── diagram.png
├── notes.txt
├── screenshot.jpg
└── README.md
```
Result: Project-Files.cbz (images converted to WebP, text files preserved)

## Configuration

**Archive Format Options**
- CBZ (ZIP): Most compatible, works with all comic readers
- CBR (RAR): Better compression, requires RAR support
- CB7Z (7-Zip): Best compression, requires 7-Zip support

**Quality Settings**
- WebP Quality: Fixed at 80% for optimal balance
- Compression: Uses Deflate compression for maximum efficiency

## Advanced Usage

**CLI Mode**
For scripting and automation:
```bash
./cbz-converter --cli [options]
```

**Batch Processing**
The tool automatically handles batch processing when multiple directories are selected, processing them sequentially with progress updates.

## Troubleshooting

**Common Issues**

"Directory contains both files and directories"
- Solution: Organize your content so directories contain either only files OR only subdirectories

"Cannot decode image"
- Solution: Ensure image files are not corrupted and are in supported formats

"Permission denied"
- Solution: Check file permissions and ensure you have write access to the target directory

**Supported Formats**
- Input Images: JPEG, PNG, GIF, BMP, WebP
- Archive Formats: CBZ (ZIP), CBR (RAR), CB7Z (7-Zip)
- Other Files: Any file type (preserved as-is)

## Development

### GitHub Actions

This project uses GitHub Actions for automated building and releasing:

- **Test Workflow**: Runs on every push and pull request to test the code and ensure it builds
- **Release Workflow**: Automatically builds binaries for all platforms and creates releases when you push a tag

### Creating a Release

1. Make sure all changes are committed and pushed
2. Create and push a version tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. GitHub Actions will automatically:
   - Build binaries for Linux (AMD64, ARM64), Windows (AMD64, ARM64), and macOS (AMD64, ARM64)
   - Create a GitHub release with all the binaries
   - Generate release notes automatically

### Local Development

Use the provided build script for local cross-platform testing:

```bash
./build.sh
```

This builds the application for all supported platforms in the `build/` directory.

## Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.

**Development Setup**
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly using `./build.sh`
5. Submit a pull request

The GitHub Actions will automatically test your changes when you create a pull request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with Bubble Tea for the TUI
- Uses Lip Gloss for styling
- Image processing powered by WebP library

## Performance

- Compression Ratio: Typically 60-80% size reduction with WebP
- Processing Speed: Approximately 100-500 images per minute (depending on hardware)
- Memory Usage: Efficient streaming processing for large collections
- Quality: 80% WebP quality maintains excellent visual fidelity