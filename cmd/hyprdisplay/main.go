package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Monitor represents a display
type Monitor struct {
	Name     string
	X        int
	Y        int
	Width    int
	Height   int
	Selected bool
}

// HyprMonitor represents the JSON structure returned by hyprctl
type HyprMonitor struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	// Add other fields as needed
}

// Model represents the application state
type Model struct {
	monitors     []Monitor
	activeIndex  int
	help         help.Model
	keys         keyMap
	quitting     bool
	windowWidth  int
	windowHeight int
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Select key.Binding
	Apply  key.Binding
	Copy   key.Binding
	Quit   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Left, k.Right, k.Select, k.Apply, k.Copy, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Select, k.Apply, k.Copy, k.Quit},
	}
}

func initialModel() Model {
	keys := keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", "space"),
			key.WithHelp("enter/space", "select monitor"),
		),
		Apply: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "apply configuration"),
		),
		Copy: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy to clipboard"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
	}

	return Model{
		monitors:    getMonitors(),
		activeIndex: 0,
		help:        help.New(),
		keys:        keys,
	}
}

func getMonitors() []Monitor {
	// Execute hyprctl monitors command
	cmd := exec.Command("hyprctl", "monitors", "-j")
	output, err := cmd.Output()
	
	// Default monitors in case of error
	defaultMonitors := []Monitor{
		{Name: "eDP-1", X: 0, Y: 0, Width: 1920, Height: 1080},
		{Name: "DP-2", X: 1920, Y: 0, Width: 1920, Height: 1080},
		{Name: "DP-4", X: 3840, Y: 0, Width: 2560, Height: 1440},
	}
	
	if err != nil {
		fmt.Println("Error running hyprctl:", err)
		return defaultMonitors
	}
	
	// Parse the JSON output
	var hyprMonitors []HyprMonitor
	err = json.Unmarshal(output, &hyprMonitors)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return defaultMonitors
	}
	
	// Convert to our Monitor type
	var monitors []Monitor
	for _, hm := range hyprMonitors {
		monitors = append(monitors, Monitor{
			Name:   hm.Name,
			X:      hm.X,
			Y:      hm.Y,
			Width:  hm.Width,
			Height: hm.Height,
		})
	}
	
	// If no monitors were found, use defaults
	if len(monitors) == 0 {
		return defaultMonitors
	}
	
	return monitors
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Up):
			if m.monitors[m.activeIndex].Selected {
				m.monitors[m.activeIndex].Y -= 10
			} else if m.activeIndex > 0 {
				m.activeIndex--
			}

		case key.Matches(msg, m.keys.Down):
			if m.monitors[m.activeIndex].Selected {
				m.monitors[m.activeIndex].Y += 10
			} else if m.activeIndex < len(m.monitors)-1 {
				m.activeIndex++
			}

		case key.Matches(msg, m.keys.Left):
			if m.monitors[m.activeIndex].Selected {
				m.monitors[m.activeIndex].X -= 10
			}

		case key.Matches(msg, m.keys.Right):
			if m.monitors[m.activeIndex].Selected {
				m.monitors[m.activeIndex].X += 10
			}

		case key.Matches(msg, m.keys.Select):
			m.monitors[m.activeIndex].Selected = !m.monitors[m.activeIndex].Selected

		case key.Matches(msg, m.keys.Apply):
			// Generate the command string
			cmdString := generateHyprlandCommand(m.monitors)
			
			// Return a command that will display the command and execute it
			return m, tea.Batch(
				tea.Printf("Applying configuration:\n%s", cmdString),
				applyConfiguration(m.monitors),
			)

		case key.Matches(msg, m.keys.Copy):
			// Generate the command string
			cmdString := generateHyprlandCommand(m.monitors)
			
			// Return a command that will copy to clipboard
			return m, tea.Batch(
				copyToClipboard(cmdString),
			)
		}

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.help.Width = msg.Width
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// Calculate a scale factor to fit all monitors on screen
	scale := calculateScale(m.monitors, m.windowWidth, m.windowHeight)

	// Render monitors
	var b strings.Builder
	b.WriteString("Hyprland Display Manager\n\n")

	// Draw a visual representation of monitors
	monitorView := renderMonitors(m.monitors, m.activeIndex, scale)
	b.WriteString(monitorView)
	b.WriteString("\n\n")

	// Show monitor details
	b.WriteString("Monitors:\n")
	for i, mon := range m.monitors {
		prefix := "  "
		if i == m.activeIndex {
			prefix = "→ "
		}
		if mon.Selected {
			prefix = "* "
		}
		b.WriteString(fmt.Sprintf("%s%s: Position(%d,%d) Size(%d×%d)\n", 
			prefix, mon.Name, mon.X, mon.Y, mon.Width, mon.Height))
	}

	b.WriteString("\n")
	b.WriteString(m.help.View(m.keys))
	return b.String()
}

func renderMonitors(monitors []Monitor, activeIndex int, scale float64) string {
	var b strings.Builder
	
	// Find the boundaries of the monitor layout
	minX, minY, maxX, maxY := 0, 0, 0, 0
	for _, mon := range monitors {
		if mon.X < minX {
			minX = mon.X
		}
		if mon.Y < minY {
			minY = mon.Y
		}
		if mon.X+mon.Width > maxX {
			maxX = mon.X + mon.Width
		}
		if mon.Y+mon.Height > maxY {
			maxY = mon.Y + mon.Height
		}
	}
	
	// Create a 2D grid to represent the display area
	// Scale down the actual dimensions to fit in terminal
	gridWidth := int(float64(maxX-minX) * scale / 20) + 10
	gridHeight := int(float64(maxY-minY) * scale / 40) + 5
	
	if gridWidth < 20 {
		gridWidth = 20
	}
	if gridHeight < 10 {
		gridHeight = 10
	}
	
	// Create an empty grid
	grid := make([][]string, gridHeight)
	for i := range grid {
		grid[i] = make([]string, gridWidth)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}
	
	// Place monitors on the grid
	for idx, mon := range monitors {
		// Calculate monitor position on grid
		startX := int(float64(mon.X-minX) * scale / 20)
		startY := int(float64(mon.Y-minY) * scale / 40)
		endX := startX + int(float64(mon.Width) * scale / 20)
		endY := startY + int(float64(mon.Height) * scale / 40)
		
		// Ensure we stay within grid bounds
		if endX >= gridWidth {
			endX = gridWidth - 1
		}
		if endY >= gridHeight {
			endY = gridHeight - 1
		}
		
		// Choose character based on monitor status
		char := "░"
		if idx == activeIndex {
			char = "▓"
		}
		if mon.Selected {
			char = "█"
		}
		
		// Draw the monitor on the grid
		for y := startY; y < endY && y < gridHeight; y++ {
			for x := startX; x < endX && x < gridWidth; x++ {
				grid[y][x] = char
			}
		}
		
		// Add monitor name in the center if there's space
		nameY := (startY + endY) / 2
		nameX := startX + 1
		name := mon.Name
		if nameX < gridWidth && nameY < gridHeight && endX-startX > len(name)+2 {
			for i, c := range name {
				if nameX+i < endX && nameX+i < gridWidth {
					grid[nameY][nameX+i] = string(c)
				}
			}
		}
	}
	
	// Render the grid
	b.WriteString("Display Layout:\n")
	b.WriteString("┌" + strings.Repeat("─", gridWidth) + "┐\n")
	
	for _, row := range grid {
		b.WriteString("│")
		for _, cell := range row {
			b.WriteString(cell)
		}
		b.WriteString("│\n")
	}
	
	b.WriteString("└" + strings.Repeat("─", gridWidth) + "┘\n")
	
	// Add a legend
	b.WriteString("\nLegend: ")
	b.WriteString("░ = Monitor  ")
	b.WriteString("▓ = Active Monitor  ")
	b.WriteString("█ = Selected for Movement\n")
	
	return b.String()
}

func calculateScale(monitors []Monitor, windowWidth, windowHeight int) float64 {
	// Find the total width and height of all monitors
	maxX, maxY := 0, 0
	for _, mon := range monitors {
		if mon.X+mon.Width > maxX {
			maxX = mon.X + mon.Width
		}
		if mon.Y+mon.Height > maxY {
			maxY = mon.Y + mon.Height
		}
	}
	
	// Calculate scale factor to fit in window
	// Adjust these values to control how much of the window is used
	scaleX := float64(windowWidth) / float64(maxX) * 0.5
	scaleY := float64(windowHeight) / float64(maxY) * 0.3
	
	// Use the smaller scale to ensure everything fits
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}
	
	// Ensure minimum scale
	if scale < 0.1 {
		scale = 0.1
	}
	
	return scale
}

func applyConfiguration(monitors []Monitor) tea.Cmd {
	return func() tea.Msg {
		var commands []string
		
		for _, mon := range monitors {
			cmd := fmt.Sprintf("hyprctl keyword monitor '%s,highres,%d,%d,1'", 
				mon.Name, mon.X, mon.Y)
			commands = append(commands, cmd)
		}
		
		// Join all commands with &&
		fullCmd := strings.Join(commands, " && ")
		
		// Execute the command
		cmd := exec.Command("bash", "-c", fullCmd)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			return tea.Printf("Error applying configuration: %v\n%s", err, output)
		}
		
		return tea.Printf("Configuration applied successfully!")
	}
}

func generateHyprlandCommand(monitors []Monitor) string {
	var commands []string
	
	for _, mon := range monitors {
		cmd := fmt.Sprintf("hyprctl keyword monitor '%s,highres,%d,%d,1'", 
			mon.Name, mon.X, mon.Y)
		commands = append(commands, cmd)
	}
	
	return strings.Join(commands, " && ")
}

// Add this new function to copy the configuration to clipboard
func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		// Try different clipboard commands based on what might be available
		// First try xclip (X11)
		cmd := exec.Command("xclip", "-selection", "clipboard")
		stdin, err := cmd.StdinPipe()
		if err == nil {
			go func() {
				defer stdin.Close()
				fmt.Fprint(stdin, text)
			}()
			
			err = cmd.Run()
			if err == nil {
				return tea.Printf("Configuration copied to clipboard using xclip!")
			}
		}
		
		// Try wl-copy (Wayland)
		cmd = exec.Command("wl-copy")
		stdin, err = cmd.StdinPipe()
		if err == nil {
			go func() {
				defer stdin.Close()
				fmt.Fprint(stdin, text)
			}()
			
			err = cmd.Run()
			if err == nil {
				return tea.Printf("Configuration copied to clipboard using wl-copy!")
			}
		}
		
		// If all else fails, just show the command
		return tea.Printf("Could not copy to clipboard. Here's your configuration:\n%s", text)
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
} 