package model

import (
	"di0build/internal/config"
	"di0build/pkg/logger"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	currentNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff87af"))
	errorDescStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	doneMark         = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d787")).SetString("✔")
	failMark         = lipgloss.NewStyle().Foreground(lipgloss.Color("#CC0000")).SetString("✗")
)

type Model struct {
	Cfg          *config.Config
	CurrentIndex int
	Phase        Phase
	
	InstallPackgesWithErr bool
	CreateSymlinksWithErr bool

	IsQuitting bool
	Width      int
	Height     int
	Spinner    spinner.Model
	Progress   progress.Model
}

func initModel(cfg *config.Config) Model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#5f5fff"))
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return Model{
		Cfg:      cfg,
		Phase:    Packages,
		Spinner:  s,
		Progress: p,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.Printf("Installing packages..."),
		installNextPackage(&m),
		m.Spinner.Tick,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width, m.Height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.Progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.Progress = newModel
		}
		return m, cmd
	case InstallMsg:
		pkg := m.Cfg.Packages[m.CurrentIndex]
		m.CurrentIndex++

		text := fmt.Sprintf(" %s %s", doneMark, pkg)
		if msg.Err != nil {
			if !m.InstallPackgesWithErr {
				m.InstallPackgesWithErr = true
			}
			err := strings.TrimSpace(fmt.Sprintf("%v", msg.Err))
			text = fmt.Sprintf(" %s %s\n    %v",
				failMark, pkg, errorDescStyle.Render(err))
		}

		if m.CurrentIndex >= len(m.Cfg.Packages) {
			m.CurrentIndex = 0
			m.Phase = Symlinks
			progressCmd := m.Progress.SetPercent(0)
			doneCmd := tea.Printf("Done! Packages installed")
			if m.InstallPackgesWithErr {
				doneCmd = tea.Printf("Done! Packages installed, but with errors")
			}
			return m, tea.Sequence(
				tea.Printf("%s", text),
				doneCmd,
				tea.Printf("\nCreating symlinks..."),
				progressCmd,
				createNextSymlink(&m),
			)
		}

		progressCmd := m.Progress.SetPercent(
			float64(m.CurrentIndex) / float64(len(m.Cfg.Packages)))

		return m, tea.Batch(
			progressCmd,
			tea.Printf("%s", text),
			installNextPackage(&m),
		)
	case SymlinkMsg:
		link := m.Cfg.Symlinks[m.CurrentIndex]
		m.CurrentIndex++

		text := fmt.Sprintf(" %s %s -> %s", doneMark, link.From, link.To)
		if msg.Err != nil {
			if !m.CreateSymlinksWithErr {
				m.CreateSymlinksWithErr = true
			}
			err := fmt.Sprintf("%v", msg.Err)
			text = fmt.Sprintf(" %s %s -> %s\n    %v",
				failMark, link.From, link.To, errorDescStyle.Render(err))
		}

		if m.CurrentIndex >= len(m.Cfg.Symlinks) {
			m.IsQuitting = true
			doneCmd := tea.Printf("Done! Symlinks created")
			if m.CreateSymlinksWithErr {
				doneCmd = tea.Printf("Done! Symlinks created, but with errors")
			}
			return m, tea.Sequence(
				tea.Printf("%s", text),
				doneCmd,
				tea.Quit,
			)
		}

		progressCmd := m.Progress.SetPercent(
			float64(m.CurrentIndex) / float64(len(m.Cfg.Symlinks)))

		return m, tea.Batch(
			progressCmd,
			tea.Printf("%s", text),
			createNextSymlink(&m),
		)
	}
	return m, nil
}

func (m Model) View() string {
	var n int
	switch m.Phase {
	case Packages:
		n = len(m.Cfg.Packages)
	case Symlinks:
		n = len(m.Cfg.Symlinks)
	}

	w := lipgloss.Width(fmt.Sprintf("%d", n))

	count := fmt.Sprintf(" %*d/%*d ", w, m.CurrentIndex, w, n)

	spin := " " + m.Spinner.View() + " "
	prog := m.Progress.View()
	cellsAvail := max(0, m.Width-lipgloss.Width(spin+prog+count))

	if m.IsQuitting {
		return ""
	}

	var info string
	switch m.Phase {
	case Packages:
		name := currentNameStyle.Render(m.Cfg.Packages[m.CurrentIndex])
		info = lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Installing " + name)
	case Symlinks:
		name := currentNameStyle.Render(m.Cfg.Symlinks[m.CurrentIndex].To)
		info = lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Linking " + name)
	}

	cellsRemaining := max(0, m.Width-lipgloss.Width(spin+info+prog+count))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + count
}

func installNextPackage(m *Model) tea.Cmd {
	pkg := m.Cfg.Packages[m.CurrentIndex]
	return func() tea.Msg {

		cmd := exec.Command("sudo", "pacman", "-S", "--noconfirm", pkg)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return InstallMsg{Err: fmt.Errorf("%s", string(output))}
		}
		return InstallMsg{Err: nil}
	}
}

func createNextSymlink(m *Model) tea.Cmd {
	link := m.Cfg.Symlinks[m.CurrentIndex]
	return func() tea.Msg {
		homeDir, _ := os.UserHomeDir()
		from := strings.Replace(link.From, "~", homeDir, 1)
		to := strings.Replace(link.To, "~", homeDir, 1)

		if err := os.MkdirAll(filepath.Dir(to), 0755); err != nil {
			return SymlinkMsg{Err: fmt.Errorf("%v", err)}
		}

		if err := os.Symlink(from, to); err != nil {
			return SymlinkMsg{Err: fmt.Errorf("%v", err)}
		}
		return SymlinkMsg{Err: nil}
	}
}

func RunInstaller(cfg *config.Config) {
	p := tea.NewProgram(initModel(cfg))
	if _, err := p.Run(); err != nil {
		logger.Fatal("Unable to start TUI platform: %v", err)
	}
}
