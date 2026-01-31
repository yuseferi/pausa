package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// BreakType represents the type of break
type BreakType string

const (
	MiniBreak BreakType = "mini"
	LongBreak BreakType = "long"
)

// Config holds all application settings
type Config struct {
	// Break intervals
	MiniBreakInterval  int `json:"miniBreakInterval"`  // minutes between mini breaks
	MiniBreakDuration  int `json:"miniBreakDuration"`  // seconds for mini break
	LongBreakInterval  int `json:"longBreakInterval"`  // number of mini breaks before long break
	LongBreakDuration  int `json:"longBreakDuration"`  // seconds for long break
	PostponeTimeMini   int `json:"postponeTimeMini"`   // minutes to postpone mini break
	PostponeTimeLong   int `json:"postponeTimeLong"`   // minutes to postpone long break

	// Feature toggles
	ShowExerciseTips     bool `json:"showExerciseTips"`
	MonitorIdleTime      bool `json:"monitorIdleTime"`
	IdleTimeThreshold    int  `json:"idleTimeThreshold"` // minutes of idle before pausing
	NotifyBeforeBreak    bool `json:"notifyBeforeBreak"`
	NotifyTimeMini       int  `json:"notifyTimeMini"`    // seconds before mini break
	NotifyTimeLong       int  `json:"notifyTimeLong"`    // seconds before long break
	ShowOnAllMonitors    bool `json:"showOnAllMonitors"`
	FullScreenBreak      bool `json:"fullScreenBreak"`
	StartAtLogin         bool `json:"startAtLogin"`
	Theme                string `json:"theme"` // "light", "dark", "system"
	Language             string `json:"language"`

	// State
	BreaksPaused bool `json:"breaksPaused"`
}

// BreakState holds the current break state
type BreakState struct {
	NextBreakTime    time.Time `json:"nextBreakTime"`
	NextBreakType    BreakType `json:"nextBreakType"`
	MiniBreakCount   int       `json:"miniBreakCount"`
	IsOnBreak        bool      `json:"isOnBreak"`
	CurrentBreakType BreakType `json:"currentBreakType"`
	BreakTimeLeft    int       `json:"breakTimeLeft"` // seconds
	IsPaused         bool      `json:"isPaused"`
}

// ExerciseTip represents a tip to show during breaks
type ExerciseTip struct {
	ID       int    `json:"id"`
	Text     string `json:"text"`
	Category string `json:"category"` // "eyes", "stretch", "move", "breathe"
}

// App struct
type App struct {
	ctx             context.Context
	config          Config
	state           BreakState
	breakTimer      *time.Timer
	exerciseTips    []ExerciseTip
	configPath      string
	previousAppName string // Track the app that was active before break
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		exerciseTips: getDefaultExerciseTips(),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.loadConfig()
	a.initializeState()
	a.startBreakTimer()
	
	// Initialize the status bar (macOS only)
	SetupStatusBar(a)
}

// domReady is called when the DOM is ready
func (a *App) domReady(ctx context.Context) {
	// DOM is ready, we can now interact with the window
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.breakTimer != nil {
		a.breakTimer.Stop()
	}
	a.saveConfig()
}

// GetConfig returns the current configuration
func (a *App) GetConfig() Config {
	return a.config
}

// SaveConfig saves the configuration
func (a *App) SaveConfig(config Config) error {
	a.config = config
	return a.saveConfig()
}

// GetState returns the current break state
func (a *App) GetState() BreakState {
	a.state.BreakTimeLeft = int(time.Until(a.state.NextBreakTime).Seconds())
	if a.state.BreakTimeLeft < 0 {
		a.state.BreakTimeLeft = 0
	}
	return a.state
}

// TakeBreakNow triggers an immediate break
func (a *App) TakeBreakNow() {
	if !a.state.IsOnBreak {
		a.StartBreak(a.state.NextBreakType)
	}
}

// StartBreak initiates a break
func (a *App) StartBreak(breakType BreakType) {
	// Store the currently active app before switching to Pausa
	a.savePreviousApp()
	
	a.state.IsOnBreak = true
	a.state.CurrentBreakType = breakType

	var duration int
	if breakType == MiniBreak {
		duration = a.config.MiniBreakDuration
	} else {
		duration = a.config.LongBreakDuration
	}
	a.state.BreakTimeLeft = duration

	// Hide window first, then show - this helps with space switching
	runtime.WindowHide(a.ctx)
	time.Sleep(50 * time.Millisecond)
	
	// Move window to current space
	BringWindowToCurrentSpace()
	time.Sleep(50 * time.Millisecond)

	// First, bring the app to the foreground using AppleScript
	a.bringToFront()
	
	// Small delay to let the app activate
	time.Sleep(100 * time.Millisecond)
	
	// Set always on top
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	
	// Show and focus the window
	runtime.WindowShow(a.ctx)
	
	// Make fullscreen if enabled
	if a.config.FullScreenBreak {
		runtime.WindowFullscreen(a.ctx)
	} else {
		// Maximize the window if not fullscreen
		runtime.WindowMaximise(a.ctx)
	}

	runtime.EventsEmit(a.ctx, "breakStarted", map[string]interface{}{
		"type":     breakType,
		"duration": duration,
	})
}

// bringToFront uses AppleScript to bring the app to the foreground
func (a *App) bringToFront() {
	// Use osascript with reopen to bring to current space
	script := `
		tell application "pausa"
			reopen
			activate
		end tell
	`
	exec.Command("osascript", "-e", script).Run()
}

// sendNotification sends a macOS notification
func (a *App) sendNotification(title, message string) {
	script := `display notification "` + message + `" with title "` + title + `"`
	exec.Command("osascript", "-e", script).Run()
}

// EndBreak ends the current break
func (a *App) EndBreak() {
	if a.state.IsOnBreak {
		a.state.IsOnBreak = false

		// Close overlay windows on secondary monitors
		CloseOverlayWindows()

		// Restore window to normal state
		runtime.WindowSetAlwaysOnTop(a.ctx, false)
		runtime.WindowUnfullscreen(a.ctx)
		runtime.WindowUnmaximise(a.ctx)

		// Update mini break count
		if a.state.CurrentBreakType == MiniBreak {
			a.state.MiniBreakCount++
			if a.state.MiniBreakCount >= a.config.LongBreakInterval {
				a.state.NextBreakType = LongBreak
				a.state.MiniBreakCount = 0
			} else {
				a.state.NextBreakType = MiniBreak
			}
		} else {
			a.state.NextBreakType = MiniBreak
			a.state.MiniBreakCount = 0
		}

		// Schedule next break
		a.scheduleNextBreak()

		runtime.EventsEmit(a.ctx, "breakEnded", nil)
	}
}

// savePreviousApp stores the currently active app name before break starts
func (a *App) savePreviousApp() {
	// Get the frontmost app that's not Pausa
	script := `
		tell application "System Events"
			set frontApp to first process whose frontmost is true
			return name of frontApp
		end tell
	`
	output, err := exec.Command("osascript", "-e", script).Output()
	if err == nil {
		appName := string(output)
		// Trim newline
		if len(appName) > 0 && appName[len(appName)-1] == '\n' {
			appName = appName[:len(appName)-1]
		}
		if appName != "pausa" && appName != "" {
			a.previousAppName = appName
		}
	}
}

// hideAndReturnToPreviousApp hides Pausa and returns to the previous active app
func (a *App) hideAndReturnToPreviousApp() {
	// Exit fullscreen first using keyboard shortcut
	exitFullscreenScript := `
		tell application "System Events"
			keystroke "f" using {control down, command down}
		end tell
	`
	exec.Command("osascript", "-e", exitFullscreenScript).Run()
	
	// Small delay to let fullscreen exit
	time.Sleep(200 * time.Millisecond)
	
	// Hide the Pausa window
	runtime.WindowHide(a.ctx)
	
	// Return to the previous app
	if a.previousAppName != "" {
		// Activate the saved previous app
		script := fmt.Sprintf(`
			tell application "%s"
				activate
			end tell
		`, a.previousAppName)
		exec.Command("osascript", "-e", script).Run()
	}
}

// SkipBreak skips the current break
func (a *App) SkipBreak() {
	a.EndBreak()
	a.hideAndReturnToPreviousApp()
	runtime.EventsEmit(a.ctx, "breakSkipped", nil)
}

// PostponeBreak postpones the current or upcoming break
func (a *App) PostponeBreak() {
	// Determine postpone time based on break type
	var postponeMinutes int
	var breakType BreakType
	
	if a.state.IsOnBreak {
		breakType = a.state.CurrentBreakType
	} else {
		breakType = a.state.NextBreakType
	}
	
	if breakType == MiniBreak {
		postponeMinutes = a.config.PostponeTimeMini
	} else {
		postponeMinutes = a.config.PostponeTimeLong
	}

	// If we're on a break, end it first and restore window
	if a.state.IsOnBreak {
		a.state.IsOnBreak = false
		
		// Close overlay windows on secondary monitors
		CloseOverlayWindows()
		
		// Restore window to normal state
		runtime.WindowSetAlwaysOnTop(a.ctx, false)
		runtime.WindowUnfullscreen(a.ctx)
		runtime.WindowUnmaximise(a.ctx)
		
		// Schedule the postponed break
		a.state.NextBreakTime = time.Now().Add(time.Duration(postponeMinutes) * time.Minute)
		a.state.NextBreakType = breakType
		a.restartBreakTimer()
		
		// Hide and return to previous app
		a.hideAndReturnToPreviousApp()
		
		runtime.EventsEmit(a.ctx, "breakEnded", nil)
	} else {
		// Just postpone the upcoming break
		a.state.NextBreakTime = a.state.NextBreakTime.Add(time.Duration(postponeMinutes) * time.Minute)
		a.restartBreakTimer()
	}

	runtime.EventsEmit(a.ctx, "breakPostponed", map[string]interface{}{
		"nextBreakTime": a.state.NextBreakTime,
		"postponedBy":   postponeMinutes,
	})
}

// PauseBreaks pauses all breaks
func (a *App) PauseBreaks() {
	a.state.IsPaused = true
	a.config.BreaksPaused = true
	if a.breakTimer != nil {
		a.breakTimer.Stop()
	}
	runtime.EventsEmit(a.ctx, "breaksPaused", nil)
}

// ResumeBreaks resumes breaks
func (a *App) ResumeBreaks() {
	a.state.IsPaused = false
	a.config.BreaksPaused = false
	a.scheduleNextBreak()
	runtime.EventsEmit(a.ctx, "breaksResumed", nil)
}

// ResetBreaks resets the break schedule
func (a *App) ResetBreaks() {
	a.state.MiniBreakCount = 0
	a.state.NextBreakType = MiniBreak
	a.scheduleNextBreak()
	runtime.EventsEmit(a.ctx, "breaksReset", nil)
}

// GetExerciseTip returns a random exercise tip
func (a *App) GetExerciseTip(category string) ExerciseTip {
	var tips []ExerciseTip
	for _, tip := range a.exerciseTips {
		if category == "" || tip.Category == category {
			tips = append(tips, tip)
		}
	}
	if len(tips) == 0 {
		return ExerciseTip{Text: "Take a moment to relax.", Category: "breathe"}
	}
	return tips[time.Now().UnixNano()%int64(len(tips))]
}

// GetTimeUntilBreak returns seconds until next break
func (a *App) GetTimeUntilBreak() int {
	return int(time.Until(a.state.NextBreakTime).Seconds())
}

// IsFirstLaunch checks if this is the first time the app is launched
func (a *App) IsFirstLaunch() bool {
	_, err := os.Stat(a.configPath)
	return os.IsNotExist(err)
}

// MarkFirstLaunchComplete marks the first launch as complete
func (a *App) MarkFirstLaunchComplete() {
	a.saveConfig()
}

// Private methods

func (a *App) loadConfig() {
	homeDir, _ := os.UserHomeDir()
	a.configPath = filepath.Join(homeDir, ".config", "pausa", "config.json")

	// Set defaults
	a.config = getDefaultConfig()

	// Try to load existing config
	data, err := os.ReadFile(a.configPath)
	if err == nil {
		json.Unmarshal(data, &a.config)
	}
}

func (a *App) saveConfig() error {
	dir := filepath.Dir(a.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(a.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(a.configPath, data, 0644)
}

func (a *App) initializeState() {
	a.state = BreakState{
		NextBreakType:  MiniBreak,
		MiniBreakCount: 0,
		IsOnBreak:      false,
		IsPaused:       a.config.BreaksPaused,
	}
}

func (a *App) scheduleNextBreak() {
	var intervalMinutes int
	if a.state.NextBreakType == MiniBreak {
		intervalMinutes = a.config.MiniBreakInterval
	} else {
		intervalMinutes = a.config.MiniBreakInterval * a.config.LongBreakInterval
	}

	a.state.NextBreakTime = time.Now().Add(time.Duration(intervalMinutes) * time.Minute)
	a.restartBreakTimer()
}

func (a *App) startBreakTimer() {
	if !a.state.IsPaused {
		a.scheduleNextBreak()
	}
}

func (a *App) restartBreakTimer() {
	if a.breakTimer != nil {
		a.breakTimer.Stop()
	}

	duration := time.Until(a.state.NextBreakTime)
	if duration < 0 {
		duration = time.Second
	}

	a.breakTimer = time.AfterFunc(duration, func() {
		a.triggerBreak()
	})

	// Schedule notification before break
	if a.config.NotifyBeforeBreak {
		var notifyTime int
		var breakName string
		if a.state.NextBreakType == MiniBreak {
			notifyTime = a.config.NotifyTimeMini
			breakName = "Mini Break"
		} else {
			notifyTime = a.config.NotifyTimeLong
			breakName = "Long Break"
		}

		notifyDuration := duration - time.Duration(notifyTime)*time.Second
		if notifyDuration > 0 {
			time.AfterFunc(notifyDuration, func() {
				// Send macOS native notification
				a.sendNotification("Pausa - "+breakName+" Coming", 
					fmt.Sprintf("Your break will start in %d seconds", notifyTime))
				
				runtime.EventsEmit(a.ctx, "breakNotification", map[string]interface{}{
					"type":       a.state.NextBreakType,
					"inSeconds": notifyTime,
				})
			})
		}
	}
}

func (a *App) triggerBreak() {
	if a.state.IsPaused {
		return
	}

	// Send notification that break is starting now
	var breakName string
	if a.state.NextBreakType == MiniBreak {
		breakName = "Mini Break"
	} else {
		breakName = "Long Break"
	}
	a.sendNotification("Pausa - "+breakName, "Time to take a break!")

	runtime.EventsEmit(a.ctx, "breakTime", map[string]interface{}{
		"type": a.state.NextBreakType,
	})

	a.StartBreak(a.state.NextBreakType)
}

func getDefaultConfig() Config {
	return Config{
		MiniBreakInterval:  10,  // 10 minutes
		MiniBreakDuration:  20,  // 20 seconds
		LongBreakInterval:  3,   // after 3 mini breaks
		LongBreakDuration:  300, // 5 minutes
		PostponeTimeMini:   2,   // 2 minutes
		PostponeTimeLong:   5,   // 5 minutes

		ShowExerciseTips:     true,
		MonitorIdleTime:      true,
		IdleTimeThreshold:    5, // 5 minutes
		NotifyBeforeBreak:    true,
		NotifyTimeMini:       10,  // 10 seconds
		NotifyTimeLong:       30,  // 30 seconds
		ShowOnAllMonitors:    true,
		FullScreenBreak:      true,  // Enable fullscreen by default
		StartAtLogin:         false,
		Theme:                "system",
		Language:             "en",
		BreaksPaused:         false,
	}
}

func getDefaultExerciseTips() []ExerciseTip {
	return []ExerciseTip{
		// Eyes
		{ID: 1, Text: "Close your eyes and take a few deep breaths.", Category: "eyes"},
		{ID: 2, Text: "Look at something 20 feet away for 20 seconds.", Category: "eyes"},
		{ID: 3, Text: "Blink rapidly for a few seconds to refresh your eyes.", Category: "eyes"},
		{ID: 4, Text: "Roll your eyes in circles, first clockwise then counter-clockwise.", Category: "eyes"},
		{ID: 5, Text: "Cup your palms over your eyes and relax for 30 seconds.", Category: "eyes"},

		// Stretch
		{ID: 6, Text: "Stretch your arms above your head and hold for 10 seconds.", Category: "stretch"},
		{ID: 7, Text: "Roll your shoulders backward 5 times, then forward 5 times.", Category: "stretch"},
		{ID: 8, Text: "Tilt your head to each side, holding for 10 seconds each.", Category: "stretch"},
		{ID: 9, Text: "Interlace your fingers and stretch your arms in front of you.", Category: "stretch"},
		{ID: 10, Text: "Stand up and do a gentle twist to each side.", Category: "stretch"},
		{ID: 11, Text: "Stretch your wrists by extending your arm and gently pulling fingers back.", Category: "stretch"},
		{ID: 12, Text: "Stand up and do a lunge. Hold for 10 seconds, then do the other leg.", Category: "stretch"},

		// Move
		{ID: 13, Text: "Stand up and walk around for a minute.", Category: "move"},
		{ID: 14, Text: "Do 10 jumping jacks to get your blood flowing.", Category: "move"},
		{ID: 15, Text: "March in place for 30 seconds.", Category: "move"},
		{ID: 16, Text: "Take a short walk to get some water.", Category: "move"},
		{ID: 17, Text: "Do 5 squats to activate your legs.", Category: "move"},

		// Breathe
		{ID: 18, Text: "Take 5 slow, deep breaths. Inhale for 4 seconds, exhale for 6.", Category: "breathe"},
		{ID: 19, Text: "Practice box breathing: inhale 4s, hold 4s, exhale 4s, hold 4s.", Category: "breathe"},
		{ID: 20, Text: "Close your eyes and focus on your breathing for 30 seconds.", Category: "breathe"},
		{ID: 21, Text: "Take a deep breath and slowly exhale, releasing all tension.", Category: "breathe"},
	}
}