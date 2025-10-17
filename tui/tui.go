package tui

import (
	"fmt"
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
	StateSelectFormat
	StateProcessing
	StateComplete
	StateError
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
}

// InitialModel returns the initial state of the application
func InitialModel() Model {
	return Model{
		state:          StateSelectDirectory,
		formats:        []string{"CBZ (ZIP)", "CBR (RAR)", "CB7Z (7Z)"},
		selectedFormat: "CBZ (ZIP)",
		deleteOriginal: false,
		cursor:         0,
	}
}

// Init implements the tea.Model interface
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements the tea.Model interface
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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

		// Process this directory
		dirPath := msg.Directories[msg.CurrentIndex]
		format := strings.ToLower(strings.Split(m.selectedFormat, " ")[0])
		parentDir := filepath.Dir(dirPath)
		dirName := filepath.Base(dirPath)
		archivePath := filepath.Join(parentDir, dirName+"."+format)

		// Create the archive (silent version)
		err := CreateSilentZipArchive(dirPath, archivePath)
		completedDirs := msg.CompletedDirs
		if err == nil {
			completedDirs = append(completedDirs, dirPath)
			// Delete the directory if flag is set
			if m.deleteOriginal {
				os.RemoveAll(dirPath)
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
			m.state = StateSelectFormat
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

// countDirectories counts the total directories to process
func (m Model) countDirectories() tea.Cmd {
	return func() tea.Msg {
		totalDirs := 0
		var directories []string

		filepath.Walk(m.directoryPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && path != m.directoryPath {
				totalDirs++
				directories = append(directories, path)
			}
			return nil
		})

		return DirectoryCountMsg{
			TotalDirs:   totalDirs,
			Directories: directories,
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
