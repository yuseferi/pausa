//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework QuartzCore

#import <Cocoa/Cocoa.h>
#import <QuartzCore/QuartzCore.h>

// Get the number of screens
int getScreenCount() {
    return (int)[[NSScreen screens] count];
}

// Get screen info for a specific screen index
void getScreenFrame(int index, int *x, int *y, int *width, int *height) {
    NSArray *screens = [NSScreen screens];
    if (index >= 0 && index < [screens count]) {
        NSScreen *screen = screens[index];
        NSRect frame = [screen frame];
        *x = (int)frame.origin.x;
        *y = (int)frame.origin.y;
        *width = (int)frame.size.width;
        *height = (int)frame.size.height;
    }
}

// Static array to hold overlay windows
static NSMutableArray *overlayWindows = nil;

// Create simple overlay windows on all screens except the main one
void createSimpleOverlays(const char *title, const char *timer, const char *tip) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (overlayWindows == nil) {
            overlayWindows = [[NSMutableArray alloc] init];
        }
        
        // Close any existing overlay windows first
        for (NSWindow *window in overlayWindows) {
            [window close];
        }
        [overlayWindows removeAllObjects];
        
        NSArray *screens = [NSScreen screens];
        
        // Skip the first screen (main screen where Wails window is)
        for (int i = 1; i < [screens count]; i++) {
            NSScreen *screen = screens[i];
            NSRect frame = [screen frame];
            
            // Create a borderless window
            NSWindow *window = [[NSWindow alloc] initWithContentRect:frame
                                                           styleMask:NSWindowStyleMaskBorderless
                                                             backing:NSBackingStoreBuffered
                                                               defer:NO
                                                              screen:screen];
            
            [window setLevel:NSStatusWindowLevel + 1];  // Above everything
            [window setCollectionBehavior:NSWindowCollectionBehaviorCanJoinAllSpaces | NSWindowCollectionBehaviorFullScreenAuxiliary];
            
            // Create gradient background
            NSView *contentView = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, frame.size.width, frame.size.height)];
            contentView.wantsLayer = YES;
            
            CAGradientLayer *gradient = [CAGradientLayer layer];
            gradient.frame = contentView.bounds;
            gradient.colors = @[
                (id)[[NSColor colorWithRed:0.1 green:0.1 blue:0.18 alpha:1.0] CGColor],
                (id)[[NSColor colorWithRed:0.09 green:0.13 blue:0.24 alpha:1.0] CGColor],
                (id)[[NSColor colorWithRed:0.06 green:0.2 blue:0.38 alpha:1.0] CGColor]
            ];
            gradient.startPoint = CGPointMake(0, 1);
            gradient.endPoint = CGPointMake(1, 0);
            [contentView.layer addSublayer:gradient];
            
            // Create container view for labels
            CGFloat centerX = frame.size.width / 2;
            CGFloat centerY = frame.size.height / 2;
            
            // Title label (e.g., "Mini Break")
            NSTextField *titleLabel = [[NSTextField alloc] initWithFrame:NSMakeRect(centerX - 300, centerY + 100, 600, 60)];
            titleLabel.stringValue = [NSString stringWithUTF8String:title];
            titleLabel.alignment = NSTextAlignmentCenter;
            titleLabel.font = [NSFont systemFontOfSize:48 weight:NSFontWeightMedium];
            titleLabel.textColor = [NSColor whiteColor];
            titleLabel.backgroundColor = [NSColor clearColor];
            titleLabel.bordered = NO;
            titleLabel.editable = NO;
            titleLabel.selectable = NO;
            [contentView addSubview:titleLabel];
            
            // Timer label
            NSTextField *timerLabel = [[NSTextField alloc] initWithFrame:NSMakeRect(centerX - 300, centerY - 50, 600, 140)];
            timerLabel.stringValue = [NSString stringWithUTF8String:timer];
            timerLabel.alignment = NSTextAlignmentCenter;
            timerLabel.font = [NSFont systemFontOfSize:120 weight:NSFontWeightUltraLight];
            timerLabel.textColor = [NSColor whiteColor];
            timerLabel.backgroundColor = [NSColor clearColor];
            timerLabel.bordered = NO;
            timerLabel.editable = NO;
            timerLabel.selectable = NO;
            timerLabel.tag = 100;  // Tag for updating later
            [contentView addSubview:timerLabel];
            
            // Tip label
            NSTextField *tipLabel = [[NSTextField alloc] initWithFrame:NSMakeRect(centerX - 400, centerY - 180, 800, 80)];
            tipLabel.stringValue = [NSString stringWithUTF8String:tip];
            tipLabel.alignment = NSTextAlignmentCenter;
            tipLabel.font = [NSFont systemFontOfSize:24 weight:NSFontWeightRegular];
            tipLabel.textColor = [[NSColor whiteColor] colorWithAlphaComponent:0.85];
            tipLabel.backgroundColor = [NSColor clearColor];
            tipLabel.bordered = NO;
            tipLabel.editable = NO;
            tipLabel.selectable = NO;
            tipLabel.lineBreakMode = NSLineBreakByWordWrapping;
            tipLabel.maximumNumberOfLines = 3;
            [contentView addSubview:tipLabel];
            
            // Emoji
            NSTextField *emojiLabel = [[NSTextField alloc] initWithFrame:NSMakeRect(centerX - 50, centerY + 180, 100, 100)];
            emojiLabel.stringValue = @"🧘";
            emojiLabel.alignment = NSTextAlignmentCenter;
            emojiLabel.font = [NSFont systemFontOfSize:80];
            emojiLabel.backgroundColor = [NSColor clearColor];
            emojiLabel.bordered = NO;
            emojiLabel.editable = NO;
            emojiLabel.selectable = NO;
            [contentView addSubview:emojiLabel];
            
            [window setContentView:contentView];
            [window makeKeyAndOrderFront:nil];
            
            [overlayWindows addObject:window];
        }
    });
}

// Close all overlay windows
void closeAllOverlays() {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (overlayWindows != nil) {
            for (NSWindow *window in overlayWindows) {
                [window close];
            }
            [overlayWindows removeAllObjects];
        }
    });
}

// Update timer on overlay windows
void updateOverlayTimer(const char *timer) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (overlayWindows != nil) {
            NSString *timerStr = [NSString stringWithUTF8String:timer];
            for (NSWindow *window in overlayWindows) {
                NSView *contentView = [window contentView];
                for (NSView *subview in [contentView subviews]) {
                    if ([subview isKindOfClass:[NSTextField class]] && subview.tag == 100) {
                        ((NSTextField *)subview).stringValue = timerStr;
                    }
                }
            }
        }
    });
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// ScreenInfo contains information about a screen
type ScreenInfo struct {
	X      int
	Y      int
	Width  int
	Height int
}

// GetScreenCount returns the number of connected screens
func GetScreenCount() int {
	return int(C.getScreenCount())
}

// GetScreenInfo returns information about a specific screen
func GetScreenInfo(index int) ScreenInfo {
	var x, y, width, height C.int
	C.getScreenFrame(C.int(index), &x, &y, &width, &height)
	return ScreenInfo{
		X:      int(x),
		Y:      int(y),
		Width:  int(width),
		Height: int(height),
	}
}

// CreateOverlayWindows creates overlay windows on all secondary screens
func CreateOverlayWindows(breakType string, duration int, tip string) {
	var title string
	if breakType == "mini" {
		title = "Mini Break"
	} else {
		title = "Long Break"
	}
	
	mins := duration / 60
	secs := duration % 60
	timer := fmt.Sprintf("%02d:%02d", mins, secs)
	
	cTitle := C.CString(title)
	cTimer := C.CString(timer)
	cTip := C.CString(tip)
	defer C.free(unsafe.Pointer(cTitle))
	defer C.free(unsafe.Pointer(cTimer))
	defer C.free(unsafe.Pointer(cTip))
	
	C.createSimpleOverlays(cTitle, cTimer, cTip)
}

// CloseOverlayWindows closes all overlay windows
func CloseOverlayWindows() {
	C.closeAllOverlays()
}

// UpdateOverlayWindows updates the timer on overlay windows
func UpdateOverlayWindows(breakType string, timeLeft int, tip string) {
	mins := timeLeft / 60
	secs := timeLeft % 60
	timer := fmt.Sprintf("%02d:%02d", mins, secs)
	
	cTimer := C.CString(timer)
	defer C.free(unsafe.Pointer(cTimer))
	
	C.updateOverlayTimer(cTimer)
}