package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Interactive browser profile configuration",
	Long:  `Scan for installed browsers and interactively select profiles to configure.`,
	RunE:  runConfigure,
}

func init() {
	rootCmd.AddCommand(configureCmd)
}

// BrowserProfile represents a detected browser profile.
type BrowserProfile struct {
	Browser  string // "Firefox", "Chrome", "Chromium"
	Name     string // profile name or path
	Path     string // full path
	Selected bool
}

// ScanBrowserProfiles detects installed browser profiles on the system.
func ScanBrowserProfiles() []BrowserProfile {
	var profiles []BrowserProfile

	home, err := os.UserHomeDir()
	if err != nil {
		return profiles
	}

	// Scan Firefox profiles
	profiles = append(profiles, scanFirefoxProfiles(home)...)

	// Scan Chrome/Chromium profiles
	profiles = append(profiles, scanChromeProfiles(home)...)

	return profiles
}

func scanFirefoxProfiles(home string) []BrowserProfile {
	var profiles []BrowserProfile

	var basePaths []string
	if runtime.GOOS == "darwin" {
		basePaths = []string{
			filepath.Join(home, "Library", "Application Support", "Firefox"),
		}
	} else {
		basePaths = []string{
			filepath.Join(home, ".mozilla", "firefox"),
			filepath.Join(home, "snap", "firefox", "common", ".mozilla", "firefox"),
			filepath.Join(home, ".var", "app", "org.mozilla.firefox", ".mozilla", "firefox"),
		}
	}

	for _, basePath := range basePaths {
		if _, err := os.Stat(basePath); err != nil {
			continue
		}

		// Read profiles.ini to find profile directories
		profilesIni := filepath.Join(basePath, "profiles.ini")
		if _, err := os.Stat(profilesIni); err != nil {
			// Try reading directories directly
			entries, err := os.ReadDir(basePath)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if entry.IsDir() && strings.Contains(entry.Name(), ".") {
					profilePath := filepath.Join(basePath, entry.Name())
					placesDB := filepath.Join(profilePath, "places.sqlite")
					if _, err := os.Stat(placesDB); err == nil {
						profiles = append(profiles, BrowserProfile{
							Browser:  "Firefox",
							Name:     entry.Name(),
							Path:     profilePath,
							Selected: true,
						})
					}
				}
			}
			continue
		}

		// Parse profiles.ini
		profileDirs := parseFirefoxProfilesIni(basePath, profilesIni)
		for _, pd := range profileDirs {
			placesDB := filepath.Join(pd.path, "places.sqlite")
			if _, err := os.Stat(placesDB); err == nil {
				profiles = append(profiles, BrowserProfile{
					Browser:  "Firefox",
					Name:     pd.name,
					Path:     pd.path,
					Selected: true,
				})
			}
		}
	}

	return profiles
}

type profileDir struct {
	name string
	path string
}

func parseFirefoxProfilesIni(basePath, iniPath string) []profileDir {
	var result []profileDir

	data, err := os.ReadFile(iniPath)
	if err != nil {
		return result
	}

	lines := strings.Split(string(data), "\n")
	var currentName, currentPath string
	var isRelative bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "[") {
			// Save previous profile if we have one
			if currentPath != "" {
				fullPath := currentPath
				if isRelative {
					fullPath = filepath.Join(basePath, currentPath)
				}
				name := currentName
				if name == "" {
					name = filepath.Base(fullPath)
				}
				result = append(result, profileDir{name: name, path: fullPath})
			}
			currentName = ""
			currentPath = ""
			isRelative = true
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			currentName = value
		case "Path":
			currentPath = value
		case "IsRelative":
			isRelative = value == "1"
		}
	}

	// Don't forget the last profile
	if currentPath != "" {
		fullPath := currentPath
		if isRelative {
			fullPath = filepath.Join(basePath, currentPath)
		}
		name := currentName
		if name == "" {
			name = filepath.Base(fullPath)
		}
		result = append(result, profileDir{name: name, path: fullPath})
	}

	return result
}

func scanChromeProfiles(home string) []BrowserProfile {
	var profiles []BrowserProfile

	type browserDef struct {
		name string
		dirs []string
	}

	var browsers []browserDef
	if runtime.GOOS == "darwin" {
		browsers = []browserDef{
			{
				name: "Chrome",
				dirs: []string{
					filepath.Join(home, "Library", "Application Support", "Google", "Chrome"),
				},
			},
			{
				name: "Chromium",
				dirs: []string{
					filepath.Join(home, "Library", "Application Support", "Chromium"),
				},
			},
		}
	} else {
		browsers = []browserDef{
			{
				name: "Chrome",
				dirs: []string{
					filepath.Join(home, ".config", "google-chrome"),
					filepath.Join(home, "snap", "google-chrome", "current", ".config", "google-chrome"),
					filepath.Join(home, ".var", "app", "com.google.Chrome", "config", "google-chrome"),
				},
			},
			{
				name: "Chromium",
				dirs: []string{
					filepath.Join(home, ".config", "chromium"),
					filepath.Join(home, "snap", "chromium", "common", "chromium"),
					filepath.Join(home, "snap", "chromium", "common", ".config", "chromium"),
					filepath.Join(home, ".var", "app", "org.chromium.Chromium", "config", "chromium"),
				},
			},
		}
	}

	for _, browser := range browsers {
		for _, dir := range browser.dirs {
			if _, err := os.Stat(dir); err != nil {
				continue
			}

			// Check for Default profile
			bookmarksPath := filepath.Join(dir, "Default", "Bookmarks")
			if _, err := os.Stat(bookmarksPath); err == nil {
				profiles = append(profiles, BrowserProfile{
					Browser:  browser.name,
					Name:     "Default",
					Path:     bookmarksPath,
					Selected: true,
				})
			}

			// Check for additional numbered profiles (Profile 1, Profile 2, etc.)
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), "Profile ") {
					bookmarksPath := filepath.Join(dir, entry.Name(), "Bookmarks")
					if _, err := os.Stat(bookmarksPath); err == nil {
						profiles = append(profiles, BrowserProfile{
							Browser:  browser.name,
							Name:     entry.Name(),
							Path:     bookmarksPath,
							Selected: true,
						})
					}
				}
			}
		}
	}

	return profiles
}

// ConfigFile represents the configuration that will be written.
type ConfigFile struct {
	Plugins        map[string]interface{} `json:"plugins,omitempty"`
	DefaultBrowser string                 `json:"defaultBrowser"`
}

func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "marks", "config.json"), nil
	default:
		return filepath.Join(home, ".config", "marks", "config.json"), nil
	}
}

// TUI Model

type configModel struct {
	profiles []BrowserProfile
	cursor   int
	done     bool
	saved    bool
	err      error
	width    int
}

type savedMsg struct {
	path string
	err  error
}

func initialModel(profiles []BrowserProfile) configModel {
	return configModel{
		profiles: profiles,
		cursor:   0,
		width:    80,
	}
}

func (m configModel) Init() tea.Cmd {
	return nil
}

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.profiles)-1 {
				m.cursor++
			}

		case " ":
			if m.cursor < len(m.profiles) {
				m.profiles[m.cursor].Selected = !m.profiles[m.cursor].Selected
			}

		case "enter":
			return m, m.saveConfig

		case "a":
			// Select all
			for i := range m.profiles {
				m.profiles[i].Selected = true
			}

		case "n":
			// Deselect all
			for i := range m.profiles {
				m.profiles[i].Selected = false
			}
		}

	case savedMsg:
		m.done = true
		m.saved = true
		m.err = msg.err
		return m, tea.Quit
	}

	return m, nil
}

func (m configModel) saveConfig() tea.Msg {
	cfgPath, err := configFilePath()
	if err != nil {
		return savedMsg{err: err}
	}

	// Build config from selected profiles
	cfg := ConfigFile{
		Plugins:        make(map[string]interface{}),
		DefaultBrowser: "firefox",
	}

	// Use the first selected Firefox profile
	for _, p := range m.profiles {
		if p.Selected && p.Browser == "Firefox" {
			cfg.Plugins["firefox"] = map[string]string{
				"profile_path": p.Path,
			}
			break
		}
	}

	// Use the first selected Chrome/Chromium profile
	for _, p := range m.profiles {
		if p.Selected && (p.Browser == "Chrome" || p.Browser == "Chromium") {
			cfg.Plugins["chrome"] = map[string]string{
				"profile_path": p.Path,
			}
			cfg.DefaultBrowser = "chrome"
			break
		}
	}

	// Write config
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		return savedMsg{err: err}
	}

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return savedMsg{err: err}
	}

	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return savedMsg{err: err}
	}

	return savedMsg{path: cfgPath}
}

func (m configModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("120"))

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	if m.done {
		if m.err != nil {
			return fmt.Sprintf("\n  ✗ Error saving config: %v\n", m.err)
		}
		if m.saved {
			return m.renderSummary()
		}
		return "\n  Configuration cancelled.\n"
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("🔖 marks — Browser Profile Configuration"))
	sb.WriteString("\n\n")

	if len(m.profiles) == 0 {
		sb.WriteString("  No browser profiles found.\n")
		sb.WriteString(helpStyle.Render("  Press q to quit."))
		return sb.String()
	}

	currentBrowser := ""
	for i, p := range m.profiles {
		// Group header
		if p.Browser != currentBrowser {
			if currentBrowser != "" {
				sb.WriteString("\n")
			}
			currentBrowser = p.Browser
			sb.WriteString(fmt.Sprintf("  %s\n", lipgloss.NewStyle().Bold(true).Render(currentBrowser)))
		}

		cursor := "  "
		if m.cursor == i {
			cursor = cursorStyle.Render("▸ ")
		}

		checkbox := "○"
		nameStr := dimStyle.Render(p.Name)
		if p.Selected {
			checkbox = selectedStyle.Render("●")
			nameStr = selectedStyle.Render(p.Name)
		}

		pathStr := dimStyle.Render(truncatePath(p.Path, 50))
		sb.WriteString(fmt.Sprintf("  %s %s %s  %s\n", cursor, checkbox, nameStr, pathStr))
	}

	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("  ↑/↓ navigate • space toggle • a all • n none • enter save • q quit"))

	return sb.String()
}

func (m configModel) renderSummary() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("120"))

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("  ✓ Configuration saved!"))
	sb.WriteString("\n\n")

	selected := 0
	for _, p := range m.profiles {
		if p.Selected {
			selected++
			sb.WriteString(fmt.Sprintf("  • %s: %s (%s)\n", p.Browser, p.Name, p.Path))
		}
	}

	if selected == 0 {
		sb.WriteString("  No profiles selected.\n")
	}

	cfgPath, _ := configFilePath()
	sb.WriteString(fmt.Sprintf("\n  Config written to: %s\n", cfgPath))

	return sb.String()
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "…" + path[len(path)-maxLen+1:]
}

func runConfigure(cmd *cobra.Command, args []string) error {
	profiles := ScanBrowserProfiles()

	model := initialModel(profiles)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %v", err)
	}

	m, ok := finalModel.(configModel)
	if ok && m.err != nil {
		return m.err
	}

	return nil
}
