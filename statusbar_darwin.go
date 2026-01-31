//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

static NSStatusItem *statusItem = nil;
static NSMenu *statusMenu = nil;
static int lastClickedTag = -1;

@interface StatusBarDelegate : NSObject
- (void)menuItemClicked:(id)sender;
- (void)setupStatusBar:(NSString *)title;
@end

@implementation StatusBarDelegate

- (void)menuItemClicked:(id)sender {
    NSMenuItem *item = (NSMenuItem *)sender;
    lastClickedTag = (int)[item tag];
}

- (void)setupStatusBar:(NSString *)title {
    if (statusItem != nil) return;
    
    statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
    [statusItem retain];
    
    statusItem.button.title = title;
    
    statusMenu = [[NSMenu alloc] init];
    statusItem.menu = statusMenu;
}

@end

static StatusBarDelegate *delegate = nil;

void initStatusBarOnMainThread(const char *title) {
    if (delegate == nil) {
        delegate = [[StatusBarDelegate alloc] init];
    }
    
    NSString *nsTitle = [NSString stringWithUTF8String:title];
    
    if ([NSThread isMainThread]) {
        [delegate setupStatusBar:nsTitle];
    } else {
        dispatch_sync(dispatch_get_main_queue(), ^{
            [delegate setupStatusBar:nsTitle];
        });
    }
}

void addMenuItemOnMainThread(const char *title, int tag, int enabled) {
    if (statusMenu == nil || delegate == nil) return;
    
    NSString *nsTitle = [NSString stringWithUTF8String:title];
    
    void (^block)(void) = ^{
        NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:nsTitle
                                                      action:@selector(menuItemClicked:)
                                               keyEquivalent:@""];
        [item setTarget:delegate];
        [item setTag:tag];
        [item setEnabled:enabled != 0];
        [statusMenu addItem:item];
    };
    
    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_sync(dispatch_get_main_queue(), block);
    }
}

void addSeparatorOnMainThread() {
    if (statusMenu == nil) return;
    
    void (^block)(void) = ^{
        [statusMenu addItem:[NSMenuItem separatorItem]];
    };
    
    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_sync(dispatch_get_main_queue(), block);
    }
}

void updateMenuTitle(int tag, const char *title) {
    if (statusMenu == nil) return;
    
    NSString *nsTitle = [NSString stringWithUTF8String:title];
    
    dispatch_async(dispatch_get_main_queue(), ^{
        NSMenuItem *item = [statusMenu itemWithTag:tag];
        if (item != nil) {
            [item setTitle:nsTitle];
        }
    });
}

void setItemVisible(int tag, int visible) {
    if (statusMenu == nil) return;
    
    dispatch_async(dispatch_get_main_queue(), ^{
        NSMenuItem *item = [statusMenu itemWithTag:tag];
        if (item != nil) {
            [item setHidden:visible == 0];
        }
    });
}

int pollClickedTag() {
    int tag = lastClickedTag;
    lastClickedTag = -1;
    return tag;
}

void cleanupStatusBar() {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (statusItem != nil) {
            [[NSStatusBar systemStatusBar] removeStatusItem:statusItem];
            [statusItem release];
            statusItem = nil;
        }
    });
}

// Make main window appear on all spaces
void makeWindowOnAllSpaces() {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSApplication *app = [NSApplication sharedApplication];
        for (NSWindow *window in [app windows]) {
            // Set all windows (except status bar items) to appear on all spaces
            if (window.contentView != nil) {
                [window setCollectionBehavior:NSWindowCollectionBehaviorCanJoinAllSpaces | 
                                              NSWindowCollectionBehaviorFullScreenAuxiliary |
                                              NSWindowCollectionBehaviorManaged];
                [window setLevel:NSFloatingWindowLevel];
            }
        }
    });
}

// Switch to the space where a break should appear
void bringWindowToCurrentSpace() {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSApplication *app = [NSApplication sharedApplication];
        
        for (NSWindow *window in [app windows]) {
            if (window.contentView != nil && ![window.className containsString:@"NSStatusBar"]) {
                // First hide the window
                [window orderOut:nil];
                
                // Set to move to active space
                [window setCollectionBehavior:NSWindowCollectionBehaviorMoveToActiveSpace | 
                                              NSWindowCollectionBehaviorFullScreenAuxiliary];
                
                // Show it again - this should place it on current space
                [window makeKeyAndOrderFront:nil];
                
                // Activate the app
                [app activateIgnoringOtherApps:YES];
                break;
            }
        }
    });
}
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Menu item tags
const (
	TagNextBreak = iota + 1
	TagTakeBreak
	TagPauseBreaks
	TagResumeBreaks
	TagShowWindow
	TagPreferences
	TagQuit
)

var statusBarApp *App

// SetupStatusBar initializes the macOS status bar
func SetupStatusBar(app *App) {
	statusBarApp = app

	// Run status bar setup in a goroutine to avoid blocking startup
	go func() {
		// Small delay to ensure the app is fully initialized
		time.Sleep(500 * time.Millisecond)
		
		// Initialize the status bar
		title := C.CString("⏸")
		C.initStatusBarOnMainThread(title)
		C.free(unsafe.Pointer(title))

		// Small delay for the status bar to be created
		time.Sleep(100 * time.Millisecond)

		// Add menu items
		addMenuItem("Next break in: --:--", TagNextBreak, false)
		C.addSeparatorOnMainThread()
		addMenuItem("Take a Break Now", TagTakeBreak, true)
		C.addSeparatorOnMainThread()
		addMenuItem("Pause Breaks", TagPauseBreaks, true)
		addMenuItem("Resume Breaks", TagResumeBreaks, true)
		setMenuItemVisible(TagResumeBreaks, false)
		C.addSeparatorOnMainThread()
		addMenuItem("Show Pausa", TagShowWindow, true)
		addMenuItem("Preferences...", TagPreferences, true)
		C.addSeparatorOnMainThread()
		addMenuItem("Quit Pausa", TagQuit, true)

		// Start update and click polling loops
		go updateLoop()
		go clickLoop()
	}()
}

func addMenuItem(title string, tag int, enabled bool) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	var e C.int = 0
	if enabled {
		e = 1
	}
	C.addMenuItemOnMainThread(cTitle, C.int(tag), e)
}

func setMenuItemVisible(tag int, visible bool) {
	var v C.int = 0
	if visible {
		v = 1
	}
	C.setItemVisible(C.int(tag), v)
}

func updateMenuItemTitle(tag int, title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.updateMenuTitle(C.int(tag), cTitle)
}

func clickLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		tag := int(C.pollClickedTag())
		if tag <= 0 || statusBarApp == nil || statusBarApp.ctx == nil {
			continue
		}

		switch tag {
		case TagTakeBreak:
			statusBarApp.TakeBreakNow()
		case TagPauseBreaks:
			statusBarApp.PauseBreaks()
		case TagResumeBreaks:
			statusBarApp.ResumeBreaks()
		case TagShowWindow:
			runtime.WindowShow(statusBarApp.ctx)
		case TagPreferences:
			runtime.WindowShow(statusBarApp.ctx)
			runtime.EventsEmit(statusBarApp.ctx, "openPreferences", nil)
		case TagQuit:
			runtime.Quit(statusBarApp.ctx)
		}
	}
}

func updateLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if statusBarApp == nil || statusBarApp.ctx == nil {
			continue
		}

		state := statusBarApp.GetState()
		timeLeft := statusBarApp.GetTimeUntilBreak()

		if timeLeft > 0 {
			mins := timeLeft / 60
			secs := timeLeft % 60

			var breakType string
			if state.NextBreakType == MiniBreak {
				breakType = "Mini"
			} else {
				breakType = "Long"
			}

			title := fmt.Sprintf("%s break in: %02d:%02d", breakType, mins, secs)
			updateMenuItemTitle(TagNextBreak, title)
		} else if state.IsOnBreak {
			updateMenuItemTitle(TagNextBreak, "Break in progress...")
		}

		// Update pause/resume visibility
		if state.IsPaused {
			setMenuItemVisible(TagPauseBreaks, false)
			setMenuItemVisible(TagResumeBreaks, true)
		} else {
			setMenuItemVisible(TagPauseBreaks, true)
			setMenuItemVisible(TagResumeBreaks, false)
		}
	}
}

// RemoveStatusBar cleans up the status bar
func RemoveStatusBar() {
	C.cleanupStatusBar()
}

// MakeWindowOnAllSpaces sets the window to appear on all macOS Spaces
func MakeWindowOnAllSpaces() {
	C.makeWindowOnAllSpaces()
}

// BringWindowToCurrentSpace moves window to current space and shows it
func BringWindowToCurrentSpace() {
	C.bringWindowToCurrentSpace()
}
