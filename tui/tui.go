package tui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AppState represents the current state of the application
type AppState int

const (
	StateSelectDirectory AppState = iota
	StateSelectMode               // New state for mode detection
	StateSelectItems              // New state for selecting items (files or directories)
	StateSelectFormat
	StateProcessing
	StateComplete
	StateError
)

// OperationMode represents the mode of operation (files or directories)
type OperationMode int

const (
	ModeUnknown     OperationMode = iota
	ModeDirectories               // Directory selection mode
	ModeFiles                     // File selection mode
)

// Model represents the application state
type Model struct {
	state          AppState
	directoryPath  string
	selectedFormat string
	deleteOriginal bool
	formats        []string
	cursor         int
	width          int
	height         int
	processingMsg  string
	errorMsg       string
	completedDirs  []string
	totalDirs      int
	currentDir     int
	currentDirName string
	processedFiles int
	totalFiles     int
	conversionLog  []string

	// New fields for mode-based operation
	operationMode  OperationMode
	availableItems []string        // List of available files or directories
	selectedItems  map[string]bool // Map of selected items
	itemStartIndex int             // For scrolling through items
	itemsPerPage   int             // Number of items to display per page

	// Preview system
	previewContent string // Current preview content
	previewType    string // Type of preview (image, text, video, etc.)
	showPreview    bool   // Whether to show preview panel
}

// InitialModel returns the initial state of the application
func InitialModel() Model {
	return Model{
		state:          StateSelectDirectory,
		formats:        []string{"CBZ (ZIP)", "CBR (RAR)", "CB7Z (7Z)"},
		selectedFormat: "CBZ (ZIP)",
		deleteOriginal: false,
		cursor:         0,
		operationMode:  ModeUnknown,
		selectedItems:  make(map[string]bool),
		itemsPerPage:   10,
	}
}

// Init implements the tea.Model interface
func (m Model) Init() tea.Cmd {
	return nil
}

// determineOperationMode scans the directory to determine if it contains only files or only directories
func (m *Model) determineOperationMode() error {
	hasFiles := false
	hasDirs := false
	m.availableItems = []string{} // Reset available items

	entries, err := os.ReadDir(m.directoryPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Skip hidden files and directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if entry.IsDir() {
			hasDirs = true
			m.availableItems = append(m.availableItems, entry.Name())
		} else {
			hasFiles = true
			m.availableItems = append(m.availableItems, entry.Name())
		}

		// If we find both files and directories, we can stop
		if hasFiles && hasDirs {
			return fmt.Errorf("directory contains both files and directories - please choose a directory with only files or only directories")
		}
	}

	if hasDirs {
		m.operationMode = ModeDirectories
	} else if hasFiles {
		m.operationMode = ModeFiles
	} else {
		return fmt.Errorf("directory is empty")
	}

	return nil
}

// Update implements the tea.Model interface
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.itemsPerPage = (m.height - 10) // Adjust items per page based on window height
		if m.itemsPerPage < 5 {
			m.itemsPerPage = 5 // Minimum items to show
		}
		return m, nil

	case DirectoryCountMsg:
		m.totalDirs = msg.TotalDirs
		m.state = StateProcessing
		// Start processing the directories
		return m, m.processDirectories(msg.Directories)

	case ProcessDirectoryMsg:
		// Process the current directory
		if msg.CurrentIndex >= len(msg.Directories) {
			// All directories processed
			return m, tea.Cmd(func() tea.Msg {
				return ProcessingCompleteMsg{
					CompletedDirs: msg.CompletedDirs,
					TotalDirs:     len(msg.Directories),
				}
			})
		}

		// Update progress
		m.currentDir = msg.CurrentIndex + 1
		m.currentDirName = filepath.Base(msg.Directories[msg.CurrentIndex])
		m.processingMsg = fmt.Sprintf("Processing %s...", filepath.Base(msg.Directories[msg.CurrentIndex]))

		// Process this item
		itemPath := msg.Directories[msg.CurrentIndex]
		format := strings.ToLower(strings.Split(m.selectedFormat, " ")[0])

		var err error
		if m.operationMode == ModeDirectories {
			// Process directory
			parentDir := filepath.Dir(itemPath)
			dirName := filepath.Base(itemPath)
			archivePath := filepath.Join(parentDir, dirName+"."+format)

			// Create the archive (silent version)
			err = CreateSilentZipArchive(itemPath, archivePath)
		} else {
			// Process files
			// Create a temporary directory to hold the files
			tempDir, tempErr := os.MkdirTemp("", "cbz-temp-*")
			if tempErr != nil {
				m.state = StateError
				m.errorMsg = fmt.Sprintf("Failed to create temp directory: %v", tempErr)
				return m, nil
			}
			defer os.RemoveAll(tempDir)

			// Copy selected files to temp directory
			destPath := filepath.Join(tempDir, filepath.Base(itemPath))
			if err := copyFile(itemPath, destPath); err != nil {
				m.state = StateError
				m.errorMsg = fmt.Sprintf("Failed to copy file: %v", err)
				return m, nil
			}

			// Create archive from temp directory
			// Use the directory name as the archive name
			archiveName := filepath.Base(m.directoryPath)
			archivePath := filepath.Join(m.directoryPath, archiveName+"."+format)
			err = CreateSilentZipArchive(tempDir, archivePath)
		}

		completedDirs := msg.CompletedDirs
		if err == nil {
			completedDirs = append(completedDirs, itemPath)
			// Delete the original if flag is set
			if m.deleteOriginal {
				os.RemoveAll(itemPath)
			}
		}

		// Process next directory with a small delay to show progress
		return m, tea.Cmd(func() tea.Msg {
			time.Sleep(500 * time.Millisecond) // Small delay to show progress
			return ProcessDirectoryMsg{
				Directories:   msg.Directories,
				CurrentIndex:  msg.CurrentIndex + 1,
				CompletedDirs: completedDirs,
			}
		})

	case ProcessingCompleteMsg:
		m.state = StateComplete
		m.completedDirs = msg.CompletedDirs
		m.totalDirs = msg.TotalDirs
		return m, nil

	case ProgressMsg:
		m.currentDirName = msg.CurrentDir
		m.currentDir = msg.CurrentDirNum
		m.totalDirs = msg.TotalDirs
		m.processedFiles = msg.ProcessedFiles
		m.totalFiles = msg.TotalFiles
		m.processingMsg = msg.Message
		return m, nil

	case FileProcessedMsg:
		logEntry := fmt.Sprintf("✓ %s (%s) → %s", msg.FileName, msg.FileType, msg.ConvertedTo)
		m.conversionLog = append(m.conversionLog, logEntry)
		// Keep only last 10 log entries to avoid cluttering
		if len(m.conversionLog) > 10 {
			m.conversionLog = m.conversionLog[len(m.conversionLog)-10:]
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case StateSelectDirectory:
			return m.updateDirectorySelection(msg)
		case StateSelectItems:
			return m.updateItemSelection(msg)
		case StateSelectFormat:
			return m.updateFormatSelection(msg)
		case StateProcessing:
			return m.updateProcessing(msg)
		case StateComplete, StateError:
			return m.updateComplete(msg)
		}
	}

	return m, nil
}

// View implements the tea.Model interface
func (m Model) View() string {
	switch m.state {
	case StateSelectDirectory:
		return m.viewDirectorySelection()
	case StateSelectItems:
		return m.viewItemSelection()
	case StateSelectFormat:
		return m.viewFormatSelection()
	case StateProcessing:
		return m.viewProcessing()
	case StateComplete:
		return m.viewComplete()
	case StateError:
		return m.viewError()
	default:
		return "Unknown state"
	}
}

// updateDirectorySelection handles input during directory selection
func (m Model) updateDirectorySelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "enter":
		if m.directoryPath != "" {
			// Validate directory exists
			if _, err := os.Stat(m.directoryPath); os.IsNotExist(err) {
				m.state = StateError
				m.errorMsg = fmt.Sprintf("Directory does not exist: %s", m.directoryPath)
				return m, nil
			}

			// Determine operation mode
			if err := m.determineOperationMode(); err != nil {
				m.state = StateError
				m.errorMsg = err.Error()
				return m, nil
			}

			m.state = StateSelectItems
			m.cursor = 0 // Reset cursor for item selection
		}
	case "backspace":
		if len(m.directoryPath) > 0 {
			m.directoryPath = m.directoryPath[:len(m.directoryPath)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.directoryPath += msg.String()
		}
	}
	return m, nil
}

// updateFormatSelection handles input during format selection
func (m Model) updateFormatSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.formats)-1 {
			m.cursor++
		}
	case "enter":
		m.selectedFormat = m.formats[m.cursor]
		m.state = StateProcessing
		return m, m.startProcessing()
	case "tab":
		m.deleteOriginal = !m.deleteOriginal
	}
	return m, nil
}

// updateProcessing handles input during processing
func (m Model) updateProcessing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}
	return m, nil
}

// updateComplete handles input when processing is complete
func (m Model) updateComplete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "enter":
		return m, tea.Quit
	case "r":
		// Restart the application
		return InitialModel(), nil
	}
	return m, nil
}

// startProcessing begins the directory processing
func (m Model) startProcessing() tea.Cmd {
	// First, count directories and set up initial progress
	return m.countDirectories()
}

// countDirectories counts the total directories or files to process
func (m Model) countDirectories() tea.Cmd {
	return func() tea.Msg {
		var items []string

		// Add selected items to process
		for item := range m.selectedItems {
			fullPath := filepath.Join(m.directoryPath, item)
			items = append(items, fullPath)
		}

		return DirectoryCountMsg{
			TotalDirs:   len(items),
			Directories: items,
		}
	}
}

// DirectoryCountMsg is sent when directory counting is complete
type DirectoryCountMsg struct {
	TotalDirs   int
	Directories []string
}

// processDirectories processes directories with progress updates
func (m Model) processDirectories(directories []string) tea.Cmd {
	return func() tea.Msg {
		// Start processing the first directory
		return ProcessDirectoryMsg{
			Directories:   directories,
			CurrentIndex:  0,
			CompletedDirs: []string{},
		}
	}
}

// ProcessDirectoryMsg is sent to process the next directory
type ProcessDirectoryMsg struct {
	Directories   []string
	CurrentIndex  int
	CompletedDirs []string
}

// ProcessingCompleteMsg is sent when processing is complete
type ProcessingCompleteMsg struct {
	CompletedDirs []string
	TotalDirs     int
}

// ProgressMsg is sent during processing to update progress
type ProgressMsg struct {
	CurrentDir     string
	CurrentDirNum  int
	TotalDirs      int
	ProcessedFiles int
	TotalFiles     int
	Message        string
}

// FileProcessedMsg is sent when a file is processed
type FileProcessedMsg struct {
	FileName    string
	FileType    string
	ConvertedTo string
}

// viewDirectorySelection renders the directory selection screen
func (m Model) viewDirectorySelection() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("📁 CBZ WebP Converter")

	instruction := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Enter the directory path to process:")

	input := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(0, 1).
		Render(m.directoryPath + "█")

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Press Enter to continue, Ctrl+C or 'q' to quit")

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			title,
			"",
			instruction,
			"",
			input,
			"",
			help,
		),
	)
}

// viewFormatSelection renders the format selection screen
func (m Model) viewFormatSelection() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("📁 CBZ WebP Converter")

	directory := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf("Directory: %s", m.directoryPath))

	instruction := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Select archive format:")

	var formatOptions []string
	for i, format := range m.formats {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		formatOptions = append(formatOptions, fmt.Sprintf("%s %s", cursor, format))
	}

	formats := lipgloss.JoinVertical(lipgloss.Left, formatOptions...)

	deleteOption := " "
	if m.deleteOriginal {
		deleteOption = "✓"
	}
	deleteText := lipgloss.NewStyle().
		Render(fmt.Sprintf("%s Delete original files after conversion", deleteOption))

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Use ↑/↓ to navigate, Tab to toggle delete option, Enter to start, Ctrl+C or 'q' to quit")

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			title,
			"",
			directory,
			"",
			instruction,
			"",
			formats,
			"",
			deleteText,
			"",
			help,
		),
	)
}

// viewProcessing renders the processing screen
func (m Model) viewProcessing() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("🔄 Processing...")

	// Overall progress
	overallProgress := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf("Directories: %d/%d", m.currentDir, m.totalDirs))

	// Current directory info
	currentDirInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("220")).
		Render(fmt.Sprintf("Current: %s", m.currentDirName))

	// Progress bar
	progressBar := m.renderProgressBar()

	// Status message
	statusMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(m.processingMsg)

	// Help text
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Press Ctrl+C or 'q' to quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		overallProgress,
		currentDirInfo,
		"",
		progressBar,
		"",
		statusMsg,
		"",
		help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// renderProgressBar creates a visual progress bar
func (m Model) renderProgressBar() string {
	if m.totalDirs == 0 {
		return ""
	}

	progress := float64(m.currentDir) / float64(m.totalDirs)
	barWidth := 30
	filledWidth := int(progress * float64(barWidth))

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"

	percentage := int(progress * 100)

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Render(fmt.Sprintf("%s %d%%", bar, percentage))
}

// viewComplete renders the completion screen
func (m Model) viewComplete() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("46")).
		Render("✅ Processing Complete!")

	summary := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf("Successfully processed %d directories", len(m.completedDirs)))

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Press Enter or 'q' to quit, 'r' to restart")

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			title,
			"",
			summary,
			"",
			help,
		),
	)
}

// updateItemSelection handles input during item selection
func (m Model) updateItemSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			// Adjust start index if cursor moves above visible area
			if m.cursor < m.itemStartIndex {
				m.itemStartIndex = m.cursor
			}
		}
	case "down", "j":
		if m.cursor < len(m.availableItems)-1 {
			m.cursor++
			// Adjust start index if cursor moves below visible area
			if m.cursor >= m.itemStartIndex+m.itemsPerPage {
				m.itemStartIndex = m.cursor - m.itemsPerPage + 1
			}
		}
	case "space", " ":
		// Toggle selection of current item
		if m.cursor >= 0 && m.cursor < len(m.availableItems) {
			currentItem := m.availableItems[m.cursor]
			if m.selectedItems[currentItem] {
				delete(m.selectedItems, currentItem)
			} else {
				m.selectedItems[currentItem] = true
			}
		}
	case "enter":
		// Only proceed if at least one item is selected
		if len(m.selectedItems) > 0 {
			m.state = StateSelectFormat
		}
	case "a":
		// Select all items
		for _, item := range m.availableItems {
			m.selectedItems[item] = true
		}
	case "n":
		// Deselect all items
		m.selectedItems = make(map[string]bool)
	}
	return m, nil
}

// viewItemSelection renders the item selection screen
func (m Model) viewItemSelection() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("📁 CBZ WebP Converter")

	modeText := "Select directories to archive"
	if m.operationMode == ModeFiles {
		modeText = "Select files to include in archive"
	}

	instruction := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(modeText)

	// Calculate visible items
	endIndex := m.itemStartIndex + m.itemsPerPage
	if endIndex > len(m.availableItems) {
		endIndex = len(m.availableItems)
	}
	visibleItems := m.availableItems[m.itemStartIndex:endIndex]

	// Build the list of items
	var itemList []string
	for i, item := range visibleItems {
		cursor := " "
		if m.itemStartIndex+i == m.cursor {
			cursor = ">"
		}

		checkbox := "[ ]"
		if m.selectedItems[item] {
			checkbox = "[✓]"
		}

		itemText := fmt.Sprintf("%s %s %s", cursor, checkbox, item)
		if m.itemStartIndex+i == m.cursor {
			itemText = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Render(itemText)
		}
		itemList = append(itemList, itemText)
	}

	items := lipgloss.JoinVertical(lipgloss.Left, itemList...)

	// Show scrollbar if needed
	if len(m.availableItems) > m.itemsPerPage {
		scrollPosition := fmt.Sprintf("(%d/%d)", m.cursor+1, len(m.availableItems))
		items = lipgloss.JoinHorizontal(lipgloss.Top, items, "  ", scrollPosition)
	}

	selectedCount := fmt.Sprintf("Selected: %d/%d", len(m.selectedItems), len(m.availableItems))

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("↑/↓: Navigate • Space: Toggle • a: Select All • n: None • Enter: Continue • Ctrl+c/q: Quit")

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			title,
			"",
			instruction,
			"",
			items,
			"",
			selectedCount,
			"",
			help,
		),
	)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// viewError renders the error screen
func (m Model) viewError() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Render("❌ Error")

	error := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(m.errorMsg)

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Press Enter or 'q' to quit, 'r' to restart")

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			title,
			"",
			error,
			"",
			help,
		),
	)
}
