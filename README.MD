# CBZ WebP Converter

A terminal-based tool for converting image directories into CBZ/CBR/CB7Z archives with WebP compression.

## Features

- Interactive TUI (Terminal User Interface)
- Two operation modes:
  - Directory mode: Convert multiple directories into separate archives
  - File mode: Combine multiple files into a single archive
- WebP compression for images
- Multiple archive format support (CBZ, CBR, CB7Z)
- Option to delete original files after conversion

## Requirements

- Go 1.24 or later
- A terminal that supports TUI applications

## Installation

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
go build -o cbz-converter
```

## Usage

1. Run the application:
```bash
./cbz-converter
```

2. Using the TUI:
   - Enter the path to your directory containing images or subdirectories
   - The tool will automatically detect if the directory contains:
     - Only directories (Directory mode)
     - Only files (File mode)
   - Use arrow keys (↑/↓) or j/k to navigate
   - Use Space to select/deselect items
   - Use 'a' to select all items
   - Use 'n' to deselect all items
   - Press Enter to continue when items are selected
   - Choose your archive format (CBZ, CBR, CB7Z)
   - Toggle whether to delete original files after conversion
   - Press Enter to start the conversion process

3. CLI mode (for scripting):
```bash
./cbz-converter --cli [options]
```

## Controls

- Navigation:
  - ↑/↓ or j/k: Move cursor up/down
  - Space: Toggle selection of current item
  - a: Select all items
  - n: Deselect all items
  - Enter: Confirm selection
  - Tab: Toggle delete original files option
  - Ctrl+C or q: Quit application
  - r: Restart (on completion screen)

## Notes

- The tool will only work with directories that contain either:
  - Only subdirectories to be archived
  - Only files to be included in an archive
- Mixed content (both files and directories) is not supported
- Hidden files and directories (starting with '.') are ignored