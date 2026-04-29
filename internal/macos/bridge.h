// Shared Objective-C / C declarations for the macOS bridge.
// All functions are safe to call from any Go goroutine; AppKit work is
// dispatched onto the main queue internally.

#ifndef PAUSA_BRIDGE_H
#define PAUSA_BRIDGE_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// Status bar -----------------------------------------------------------------

// Initializes the menu-bar item with the given title (UTF-8). Idempotent.
// Pass NULL or "" if you plan to set an image instead.
void pausa_statusbar_init(const char *title);

// Replaces the menu-bar item title. Safe to call frequently.
void pausa_statusbar_set_title(const char *title);

// Replaces the menu-bar item icon with a built-in native vector template
// glyph.
//   0 = running (pause bars)
//   1 = manual paused
//   2 = busy auto-paused (video/meeting)
//   3 = break in progress
void pausa_statusbar_set_builtin_icon(int state);

// Replaces the menu-bar item's icon with a PNG loaded from `data` of
// length `len`. The image is rendered as a template image (auto-tinted by
// macOS to match the menu-bar appearance — black on light, white on dark,
// blue when highlighted). Pass len=0 to clear the image.
void pausa_statusbar_set_image_data(const void *data, int len);

// Adds a menu item identified by `tag`. Selecting it invokes the registered
// Go click callback (see pausa_register_status_click). Pass enabled=0 to
// add a disabled (informational) item.
void pausa_statusbar_add_item(const char *title, int tag, int enabled);

// Adds a separator.
void pausa_statusbar_add_separator(void);

// Updates an existing item's title.
void pausa_statusbar_update_item(int tag, const char *title);

// Sets visibility for an existing item.
void pausa_statusbar_set_item_hidden(int tag, int hidden);

// Removes the menu-bar item entirely.
void pausa_statusbar_remove(void);

// Workspace ------------------------------------------------------------------

// Returns the localized name of the frontmost application as a newly
// allocated UTF-8 C string. Caller must free(). Returns NULL on failure.
char *pausa_workspace_frontmost_app(void);

// Activates the application with the given bundle identifier or localized
// name (tries bundleID first). Returns 1 on success, 0 on failure.
int pausa_workspace_activate(const char *identifier);

// Hides the Pausa application from the macOS window manager (equivalent to
// Cmd-H), causing macOS to switch focus to whatever app is next in the
// activation stack. Use this in combination with pausa_workspace_activate
// to reliably return to a fullscreen-Space app when the break ends.
void pausa_app_hide_self(void);

// Restores the previous application using the modern NSWorkspace API.
// Unlike pausa_workspace_activate, this works correctly when the target
// app is in a fullscreen Space — macOS will switch Spaces to follow it.
// Returns 1 on success, 0 on failure.
int pausa_workspace_activate_url(const char *bundleIdentifier);

// Notifications --------------------------------------------------------------

// Requests notification authorization. Asynchronous; safe to call any time.
void pausa_notify_request_auth(void);

// Posts a notification with optional Skip/Postpone action buttons. When the
// user clicks an action, the registered Go callback is invoked with one of
// the PAUSA_NOTIFY_ACTION_* tags.
void pausa_notify_post(const char *title, const char *body, int withActions);

#define PAUSA_NOTIFY_ACTION_DEFAULT  1
#define PAUSA_NOTIFY_ACTION_SKIP     2
#define PAUSA_NOTIFY_ACTION_POSTPONE 3

// Idle -----------------------------------------------------------------------

// Returns the number of seconds since the last user input event (mouse or
// keyboard). Returns 0 on error.
double pausa_idle_seconds(void);

// Busy detection -------------------------------------------------------------

// Returns 1 if the system's default audio input device is currently being
// used by some process (any open mic stream). This catches video calls in
// Zoom, Meet, Teams, Discord, browser-based meetings, voice memos, etc.
// Uses CoreAudio's HAL — no permission prompts, no microphone access of
// our own. Returns 0 if no input device is active or on error.
int pausa_busy_microphone_active(void);

// Returns 1 if the macOS Now Playing system reports that audio or video is
// currently playing somewhere — Chrome / Safari / YouTube / Spotify /
// QuickTime / etc. all publish to this. Uses the (private) MediaRemote
// framework; loads it via dlsym so we degrade gracefully if Apple removes
// or changes the API. Returns 0 if nothing is playing or detection failed.
int pausa_busy_now_playing_active(void);

// Returns 1 if the system's default audio output device is currently being
// used by some process. This acts as a fallback for browsers/apps that do
// not publish to the macOS Now Playing API. On its own this can be noisy
// (notification sounds, etc.), so the Go side debounces it before treating
// it as a true busy/media signal.
int pausa_busy_output_active(void);

// Windows --------------------------------------------------------------------

// Returns the number of attached screens.
int pausa_screen_count(void);

// Brings the main app window to the current Space and makes it key.
void pausa_window_bring_to_current_space(void);

// Sets the app's activation policy to "accessory" (menu-bar app, no Dock
// icon). This is required for the break overlay to appear on top of
// fullscreen apps — regular apps cannot float over fullscreen Spaces.
// Idempotent.
void pausa_app_set_accessory(void);

// Creates an NSPanel-based break overlay. Panels use
// NSWindowStyleMaskNonactivatingPanel + a high window level + the right
// collection behavior, which is the only reliable way to render UI on top
// of fullscreen-app Spaces (regular NSWindows can't).
//
//   kind             "short" or "long" (long enables the breathing guide)
//   title            display title ("Short Break" / "Long Break")
//   timer            initial countdown text ("00:20")
//   tip              exercise tip text
//   hexAccent        hex color used for the gradient
//   showActions      whether to render Skip / Postpone buttons
//   fullscreen       1 = cover entire screen edge-to-edge,
//                    0 = centered ~720x520 panel
//   currentScreenOnly 1 = overlay only the screen with the mouse cursor,
//                     0 = overlay every connected screen
//
// Action buttons invoke pausaOverlayAction (Go-exported) when clicked.
void pausa_overlays_create(const char *kind, const char *title,
                           const char *timer, const char *tip,
                           const char *hexAccent, int showActions,
                           int fullscreen, int currentScreenOnly);

// Updates the timer text on every overlay window (1Hz).
void pausa_overlays_update_timer(const char *timer);

// Closes all overlay windows.
void pausa_overlays_close(void);

// Action codes passed to pausaOverlayAction.
#define PAUSA_OVERLAY_SKIP     1
#define PAUSA_OVERLAY_POSTPONE 2

#ifdef __cplusplus
}
#endif

#endif // PAUSA_BRIDGE_H
