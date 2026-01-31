<p align="center">
  <img src="build/appicon.png" alt="Pausa Logo" width="128" height="128">
</p>

<h1 align="center">Pausa</h1>

<p align="center">
  <strong>Take mindful breaks, protect your health</strong>
</p>

<p align="center">
  A modern, beautiful break reminder app for macOS built with Go and Wails.
  <br>
  Inspired by <a href="https://hovancik.net/stretchly/">Stretchly</a> and <a href="https://www.dejal.com/timeout/">Time Out</a>.
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#building">Building</a> •
  <a href="#contributing">Contributing</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/platform-macOS-blue?style=flat-square" alt="Platform">
  <img src="https://img.shields.io/badge/built%20with-Go%20%2B%20Wails-00ADD8?style=flat-square" alt="Built with Go">
  <img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="License">
</p>

---

## ✨ Why Pausa?

As developers and knowledge workers, we spend countless hours staring at screens. This leads to:

- 👁️ **Eye strain** from continuous screen exposure
- 🪑 **Poor posture** from prolonged sitting
- 🧠 **Mental fatigue** from lack of breaks
- 💪 **Physical tension** in neck, shoulders, and back

**Pausa** gently reminds you to take regular breaks with beautiful, non-intrusive notifications and calming break screens that guide you through eye exercises, stretches, and breathing techniques.

---

## 🎯 Features

### 🕐 Smart Break Scheduling
- **Mini breaks** - Short 20-second breaks every 10 minutes for eye rest
- **Long breaks** - 5-minute breaks after every 3 mini breaks for stretching and movement
- Fully customizable intervals and durations

### 🎨 Beautiful Break Screens
- Stunning gradient backgrounds with floating particles
- Smooth animations and transitions
- Different themes for mini and long breaks
- Circular countdown timer with progress visualization

### 💆 Wellness Tips
- Curated exercise tips during breaks
- Categories: 👀 Eyes, 🙆 Stretch, 🚶 Move, 🌬️ Breathe
- Guided breathing animation for long breaks

### 🖥️ Multi-Monitor Support
- Overlay screens appear on all connected monitors
- Never miss a break reminder, no matter which screen you're using

### 📊 Menu Bar Integration
- Native macOS menu bar app (⏸ icon)
- Shows countdown to next break
- Quick access to take breaks, pause, or adjust settings

### ⚙️ Flexible Controls
- **Skip** - Skip the current break entirely
- **Postpone** - Delay breaks with configurable postpone times
- **Pause/Resume** - Temporarily disable break reminders
- **Take Break Now** - Trigger an immediate break

### 🌓 Theme Support
- Light, Dark, and System themes
- Respects your macOS appearance settings

### 🔔 Native Notifications
- macOS notification before breaks start
- Configurable notification timing

---

## 🚀 Installation

### Download Binary

Download the latest release from the [Releases](https://github.com/yourusername/pausa/releases) page.

1. Download `pausa-macos-arm64.zip` (Apple Silicon) or `pausa-macos-amd64.zip` (Intel)
2. Unzip and drag `Pausa.app` to your Applications folder
3. Launch Pausa from Applications

### Build from Source

See the [Building](#building) section below.

---

## 💡 Usage

### Getting Started

1. Launch Pausa - it will appear in your menu bar as ⏸
2. Click the menu bar icon to see break status and options
3. Wait for your first break, or click "Take a Break Now"

### Menu Bar Options

| Option | Description |
|--------|-------------|
| **Next break in** | Shows countdown to next break |
| **Take a Break Now** | Triggers immediate break |
| **Pause Breaks** | Temporarily stops all breaks |
| **Resume Breaks** | Resumes break schedule |
| **Show Pausa** | Opens the main window |
| **Preferences** | Opens settings |
| **Quit Pausa** | Closes the application |

### During a Break

- **Wait** - Let the break complete naturally for full benefit
- **Postpone** - Delays the break by configured time
- **Skip** - Skips the break entirely (use sparingly!)

---

## ⚙️ Configuration

### Break Timing

| Setting | Default | Description |
|---------|---------|-------------|
| Mini Break Interval | 10 min | Time between mini breaks |
| Mini Break Duration | 20 sec | Length of mini breaks |
| Long Break Interval | 3 | Mini breaks before long break |
| Long Break Duration | 5 min | Length of long breaks |
| Mini Postpone Time | 2 min | Postpone time for mini breaks |
| Long Postpone Time | 5 min | Postpone time for long breaks |

### Notifications

| Setting | Default | Description |
|---------|---------|-------------|
| Notify Before Break | ✓ | Show notification before breaks |
| Mini Break Warning | 10 sec | Seconds before mini break |
| Long Break Warning | 30 sec | Seconds before long break |

### Display

| Setting | Default | Description |
|---------|---------|-------------|
| Fullscreen Breaks | ✓ | Show breaks in fullscreen mode |
| Show on All Monitors | ✓ | Display overlay on all screens |
| Show Exercise Tips | ✓ | Display wellness tips during breaks |

### Configuration File

Settings are stored in `~/.config/pausa/config.json`

---

## 🛠️ Building

### Prerequisites

- [Go 1.25+](https://golang.org/dl/)
- [Wails CLI v2](https://wails.io/docs/gettingstarted/installation)
- [Node.js 18+](https://nodejs.org/)
- Xcode Command Line Tools

### Build Steps

```bash
# Clone the repository
git clone https://github.com/yourusername/pausa.git
cd pausa

# Install Wails CLI (if not installed)
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build the application
wails build

# The built app will be in build/bin/pausa.app
```

### Development

```bash
# Run in development mode with hot reload
wails dev
```

---

## 🏗️ Tech Stack

- **Backend**: [Go](https://golang.org/) - Fast, reliable, and efficient
- **Frontend**: [Vue 3](https://vuejs.org/) - Progressive JavaScript framework
- **Framework**: [Wails v2](https://wails.io/) - Build desktop apps with Go and Web technologies
- **Styling**: Custom CSS with modern features (glassmorphism, animations)

### Project Structure

```
pausa/
├── app.go                    # Main application logic
├── main.go                   # Entry point
├── statusbar_darwin.go       # macOS menu bar integration
├── multimonitor_darwin.go    # Multi-monitor support
├── frontend/
│   ├── src/
│   │   ├── App.vue          # Root component
│   │   ├── components/
│   │   │   ├── BreakWindow.vue      # Break screen UI
│   │   │   ├── MainView.vue         # Main dashboard
│   │   │   ├── PreferencesModal.vue # Settings modal
│   │   │   └── WelcomeScreen.vue    # First-run setup
│   │   └── style.css        # Global styles
│   └── package.json
├── build/
│   └── appicon.png          # Application icon
└── README.md
```

---

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Ideas for Contributions

- [ ] Linux support
- [ ] Windows support
- [ ] Custom break sounds
- [ ] Break statistics and analytics
- [ ] Pomodoro mode integration
- [ ] iCloud sync for settings
- [ ] Localization (i18n)
- [ ] Custom exercise tips
- [ ] Idle time detection
- [ ] Do Not Disturb integration

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Inspired by [Stretchly](https://hovancik.net/stretchly/) and [Time Out](https://www.dejal.com/timeout/)
- Built with the amazing [Wails](https://wails.io/) framework
- Icons from native emoji for universal compatibility

---

## ⭐ Star History

If you find Pausa helpful, please consider giving it a ⭐ on GitHub!

---

<p align="center">
  Made with ❤️ for your health and wellbeing
</p>

<p align="center">
  <a href="https://github.com/yourusername/pausa/issues">Report Bug</a> •
  <a href="https://github.com/yourusername/pausa/issues">Request Feature</a>
</p>