# HyprDisplay

![HyprDisplay Demo](https://github.com/yourusername/hyprdisplay/raw/main/demo.gif)

A TUI (Terminal User Interface) application for managing Hyprland display configurations. This tool allows you to visually position your monitors and generate the appropriate Hyprland commands to apply your layout.

## Features

- Visual representation of your connected monitors
- Interactive positioning using arrow keys
- Apply configurations directly to Hyprland
- Copy configurations to clipboard for manual application
- Simple, keyboard-driven interface

## Installation

### Prerequisites

- Go 1.16 or later
- Hyprland
- For clipboard functionality:
  - `wl-clipboard` package

### Building from source
```bash
# Clone the repository
git clone https://github.com/yourusername/hyprdisplay.git
cd hyprdisplay

# Build the application
go build

# Run the application
./hyprdisplay
```

## Usage

1. Launch the application:
   ```bash
   ./hyprdisplay
   ```

2. The interface will display your connected monitors as rectangles in a grid.

3. Navigation:
   - Use arrow keys (or h/j/k/l) to navigate between monitors
   - Press Enter or Space to select a monitor for movement
   - When a monitor is selected, use arrow keys to reposition it
   - Press Enter or Space again to deselect the monitor

4. Actions:
   - Press `a` to apply the configuration directly to Hyprland
   - Press `c` to copy the configuration commands to clipboard
   - Press `q` to quit the application

## Legend

- `░` - Regular monitor
- `▓` - Active monitor (currently focused)
- `█` - Selected monitor (being moved)

## How it works

HyprDisplay uses the `hyprctl monitors` command to detect your current monitor setup. It then provides a visual interface to adjust the positions of these monitors. When you're satisfied with the layout, you can either:

1. Apply the configuration directly using `hyprctl keyword monitor` commands
2. Copy these commands to your clipboard to run them manually or add to your Hyprland configuration

## Example

Display Layout:
┌────────────────────────────────────┐
│                                    │
│ ░░░░░░░░░░                         │
│ ░eDP-1░░░                          │
│ ░░░░░░░░░░                         │
│                                    │
│         ▓▓▓▓▓▓▓▓▓▓▓▓               │
│         ▓DP-2▓▓▓▓▓▓               │
│         ▓▓▓▓▓▓▓▓▓▓▓▓               │
│                                    │
│                   ████████████████ │
│                   █DP-4█████████  │
│                   ████████████████ │
│                                    │
└────────────────────────────────────┘

Legend: ░ = Monitor  ▓ = Active Monitor  █ = Selected for Movement

Monitors:
  eDP-1: Position(0,0) Size(1920×1080)
→ DP-2: Position(1920,0) Size(1920×1080)
* DP-4: Position(3840,0) Size(2560×1440)

↑/k move up · ↓/j move down · ←/h move left · →/l move right · enter/space select monitor · a apply configuration · c copy to clipboard · q/ctrl+c quit

## Configuration Commands

The generated Hyprland commands will look like:

```bash
hyprctl keyword monitor 'eDP-1,highres,0,0,1' && hyprctl keyword monitor 'DP-2,highres,1920,0,1' && hyprctl keyword monitor 'DP-4,highres,3840,0,1'
```
