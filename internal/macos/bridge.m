// macOS native bridge for Pausa.
//
// Threading model: every AppKit call is dispatched onto the main queue
// (synchronously when called from the main thread, asynchronously
// otherwise). User callbacks back into Go are invoked on the main queue.
//
// Defensive principles applied throughout:
//   * Every AppKit/UNUserNotificationCenter call is wrapped in @try/@catch
//     so an Obj-C exception never propagates into Go (where it becomes an
//     unrecoverable EXC_BAD_ACCESS).
//   * UNUserNotificationCenter requires a code-signed bundled app; we
//     detect unbundled execution (`wails dev`) and fall back to a no-op.
//   * AppKit calls that touch NSApplication require AppKit to be running.
//     We early-return if NSApp is nil (e.g. very-early startup).

#import <Cocoa/Cocoa.h>
#import <QuartzCore/QuartzCore.h>
#import <CoreGraphics/CoreGraphics.h>
#import <CoreAudio/CoreAudio.h>
#import <UserNotifications/UserNotifications.h>

#include "bridge.h"
#include <stdlib.h>
#include <string.h>
#include <dlfcn.h>

// Go-exported callbacks ------------------------------------------------------
extern void pausaStatusBarClicked(int tag);
extern void pausaNotificationAction(int action);
extern void pausaOverlayAction(int action);

// Helpers --------------------------------------------------------------------

static inline NSString *NSStr(const char *s) {
    if (s == NULL) return @"";
    NSString *out = [NSString stringWithUTF8String:s];
    return out ? out : @"";
}

static inline void onMain(dispatch_block_t block) {
    if ([NSThread isMainThread]) {
        @try { block(); } @catch (NSException *e) {
            NSLog(@"pausa: main-thread exception: %@", e);
        }
    } else {
        dispatch_async(dispatch_get_main_queue(), ^{
            @try { block(); } @catch (NSException *e) {
                NSLog(@"pausa: dispatched exception: %@", e);
            }
        });
    }
}

// Returns YES iff this process is running from a proper .app bundle (i.e.
// a code-signed install, not `wails dev`).
static BOOL isBundledApp(void) {
    NSBundle *b = [NSBundle mainBundle];
    if (b == nil) return NO;
    NSString *path = b.bundlePath ?: @"";
    return [path hasSuffix:@".app"];
}

// Click target object --------------------------------------------------------

@interface PausaMenuTarget : NSObject
- (void)menuItemSelected:(id)sender;
@end

@implementation PausaMenuTarget
- (void)menuItemSelected:(id)sender {
    if (![sender isKindOfClass:[NSMenuItem class]]) return;
    NSMenuItem *item = (NSMenuItem *)sender;
    pausaStatusBarClicked((int)item.tag);
}
@end

// Button target: routes overlay button clicks (Skip / Postpone) into Go.
@interface PausaOverlayTarget : NSObject
- (void)buttonClicked:(id)sender;
@end

@implementation PausaOverlayTarget
- (void)buttonClicked:(id)sender {
    if (![sender isKindOfClass:[NSButton class]]) return;
    NSButton *b = (NSButton *)sender;
    pausaOverlayAction((int)b.tag);
}
@end

// Notification delegate ------------------------------------------------------

@interface PausaNotifDelegate : NSObject <UNUserNotificationCenterDelegate>
@end

@implementation PausaNotifDelegate
- (void)userNotificationCenter:(UNUserNotificationCenter *)center
       willPresentNotification:(UNNotification *)notification
         withCompletionHandler:(void (^)(UNNotificationPresentationOptions))completionHandler {
    completionHandler(UNNotificationPresentationOptionBanner);
}

- (void)userNotificationCenter:(UNUserNotificationCenter *)center
didReceiveNotificationResponse:(UNNotificationResponse *)response
         withCompletionHandler:(void (^)(void))completionHandler {
    NSString *actionID = response.actionIdentifier ?: @"";
    int action = PAUSA_NOTIFY_ACTION_DEFAULT;
    if ([actionID isEqualToString:@"SKIP"])     action = PAUSA_NOTIFY_ACTION_SKIP;
    else if ([actionID isEqualToString:@"POSTPONE"]) action = PAUSA_NOTIFY_ACTION_POSTPONE;
    pausaNotificationAction(action);
    completionHandler();
}
@end

// Globals (strong refs under ARC) -------------------------------------------

static NSStatusItem               *gStatusItem    = nil;
static NSMenu                     *gStatusMenu    = nil;
static PausaMenuTarget            *gMenuTarget    = nil;
static PausaOverlayTarget         *gOverlayTarget = nil;
static PausaNotifDelegate         *gNotifDelegate = nil;
static NSMutableArray<NSWindow *> *gOverlays      = nil;
static BOOL                        gNotifReady    = NO;

// View tags used to find label/button subviews for updates later.
#define PAUSA_TAG_TIMER 100

// Status bar -----------------------------------------------------------------

void pausa_statusbar_init(const char *title) {
    NSString *t = NSStr(title);
    onMain(^{
        if (gStatusItem != nil) return;
        if (NSApp == nil) {
            NSLog(@"pausa: NSApp not ready, skipping statusbar init");
            return;
        }
        gMenuTarget = [[PausaMenuTarget alloc] init];
        gStatusItem = [[NSStatusBar systemStatusBar]
                       statusItemWithLength:NSVariableStatusItemLength];
        if (gStatusItem == nil) return;
        gStatusItem.button.title = t;
        gStatusMenu = [[NSMenu alloc] init];
        gStatusMenu.autoenablesItems = NO;
        gStatusItem.menu = gStatusMenu;
    });
}

void pausa_statusbar_set_title(const char *title) {
    NSString *t = NSStr(title);
    onMain(^{
        if (gStatusItem) gStatusItem.button.title = t;
    });
}

// makeFallbackStatusIcon draws a very simple template icon using vector
// paths. Used on older macOS versions where SF Symbols aren't available.
static NSImage *makeFallbackStatusIcon(NSInteger state) {
    NSRect r = NSMakeRect(0, 0, 18, 18);
    NSImage *img = [[NSImage alloc] initWithSize:r.size];
    [img lockFocus];
    [[NSColor blackColor] setFill];

    switch (state) {
        case 1: { // manual paused: pause bars in a circle
            NSBezierPath *circle = [NSBezierPath bezierPathWithOvalInRect:NSMakeRect(1.5, 1.5, 15, 15)];
            [circle setLineWidth:2.0];
            [[NSColor blackColor] setStroke];
            [circle stroke];
        } break;
        case 2: { // busy: simple video camera-ish shape
            NSBezierPath *body = [NSBezierPath bezierPathWithRoundedRect:NSMakeRect(3, 5, 8, 8) xRadius:1.5 yRadius:1.5];
            [body fill];
            NSBezierPath *tri = [NSBezierPath bezierPath];
            [tri moveToPoint:NSMakePoint(11, 7)];
            [tri lineToPoint:NSMakePoint(15, 5.5)];
            [tri lineToPoint:NSMakePoint(15, 12.5)];
            [tri closePath];
            [tri fill];
            img.template = YES;
            [img unlockFocus];
            return img;
        }
        case 3: { // break: timer/clock-ish outline
            NSBezierPath *circle = [NSBezierPath bezierPathWithOvalInRect:NSMakeRect(3, 3, 12, 12)];
            [circle setLineWidth:2.0];
            [[NSColor blackColor] setStroke];
            [circle stroke];
            NSBezierPath *hand = [NSBezierPath bezierPath];
            [hand moveToPoint:NSMakePoint(9, 9)];
            [hand lineToPoint:NSMakePoint(9, 5.5)];
            [hand moveToPoint:NSMakePoint(9, 9)];
            [hand lineToPoint:NSMakePoint(11.5, 10.5)];
            [hand setLineWidth:2.0];
            [hand stroke];
            img.template = YES;
            [img unlockFocus];
            return img;
        }
    }

    // default running/manual pause bars
    NSBezierPath *left = [NSBezierPath bezierPathWithRoundedRect:NSMakeRect(5, 3, 3.5, 12) xRadius:1.3 yRadius:1.3];
    NSBezierPath *right = [NSBezierPath bezierPathWithRoundedRect:NSMakeRect(10, 3, 3.5, 12) xRadius:1.3 yRadius:1.3];
    [left fill];
    [right fill];

    img.template = YES;
    [img unlockFocus];
    return img;
}

void pausa_statusbar_set_builtin_icon(int state) {
    NSInteger st = state;
    onMain(^{
        @try {
            if (gStatusItem == nil) return;

            NSImage *img = nil;
            if ([NSImage respondsToSelector:@selector(imageWithSystemSymbolName:accessibilityDescription:)]) {
                NSString *symbol = @"pause.fill";
                switch (st) {
                    case 1: symbol = @"pause.circle.fill"; break;
                    case 2: symbol = @"video.fill"; break;
                    case 3: symbol = @"timer"; break;
                }
                img = [NSImage imageWithSystemSymbolName:symbol accessibilityDescription:nil];
                if (img != nil) {
                    img = [img copy];
                    img.size = NSMakeSize(18, 18);
                    img.template = YES;
                }
            }
            if (img == nil) {
                img = makeFallbackStatusIcon(st);
            }
            gStatusItem.button.image = img;
            gStatusItem.button.title = @"";
            gStatusItem.button.imagePosition = NSImageOnly;
        } @catch (NSException *e) {
            NSLog(@"pausa: set_builtin_icon exception: %@", e);
        }
    });
}

void pausa_statusbar_set_image_data(const void *data, int len) {
    if (data == NULL || len <= 0) {
        onMain(^{
            if (gStatusItem) gStatusItem.button.image = nil;
        });
        return;
    }
    // Copy the bytes into an NSData on the calling thread (the cgo buffer
    // may be freed once this function returns) and dispatch the AppKit
    // mutation onto the main queue.
    NSData *bytes = [NSData dataWithBytes:data length:len];
    onMain(^{
        @try {
            if (gStatusItem == nil) return;
            NSImage *img = [[NSImage alloc] initWithData:bytes];
            if (img == nil) {
                NSLog(@"pausa: status image: failed to decode PNG");
                return;
            }
            // Render at 18pt logical size — Apple's menu-bar icon target
            // for non-control items. NSImage will pick the right backing
            // representation for the current display scale.
            img.size = NSMakeSize(18, 18);
            // Template image = alpha-only; macOS supplies the color so
            // the icon adapts to light/dark mode automatically.
            img.template = YES;
            gStatusItem.button.image = img;
            // Don't show both image and title; clear the title so the
            // image isn't shoved sideways by stale text.
            gStatusItem.button.title = @"";
            // Image-leading layout looks correct in macOS menu bar.
            gStatusItem.button.imagePosition = NSImageOnly;
        } @catch (NSException *e) {
            NSLog(@"pausa: set_image_data exception: %@", e);
        }
    });
}

void pausa_statusbar_add_item(const char *title, int tag, int enabled) {
    NSString *t = NSStr(title);
    int e = enabled;
    int tg = tag;
    onMain(^{
        if (gStatusMenu == nil || gMenuTarget == nil) return;
        NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:t
                                                      action:@selector(menuItemSelected:)
                                               keyEquivalent:@""];
        item.target = gMenuTarget;
        item.tag = tg;
        item.enabled = (e != 0);
        [gStatusMenu addItem:item];
    });
}

void pausa_statusbar_add_separator(void) {
    onMain(^{
        if (gStatusMenu) [gStatusMenu addItem:[NSMenuItem separatorItem]];
    });
}

void pausa_statusbar_update_item(int tag, const char *title) {
    NSString *t = NSStr(title);
    int tg = tag;
    onMain(^{
        if (gStatusMenu == nil) return;
        NSMenuItem *item = [gStatusMenu itemWithTag:tg];
        if (item) item.title = t;
    });
}

void pausa_statusbar_set_item_hidden(int tag, int hidden) {
    int tg = tag, h = hidden;
    onMain(^{
        if (gStatusMenu == nil) return;
        NSMenuItem *item = [gStatusMenu itemWithTag:tg];
        if (item) item.hidden = (h != 0);
    });
}

void pausa_statusbar_remove(void) {
    onMain(^{
        if (gStatusItem) {
            [[NSStatusBar systemStatusBar] removeStatusItem:gStatusItem];
            gStatusItem = nil;
            gStatusMenu = nil;
        }
    });
}

// Workspace ------------------------------------------------------------------

char *pausa_workspace_frontmost_app(void) {
    __block char *result = NULL;
    @try {
        NSRunningApplication *app = [[NSWorkspace sharedWorkspace] frontmostApplication];
        if (app == nil) return NULL;
        NSString *name = app.bundleIdentifier;
        if (name.length == 0) name = app.localizedName;
        if (name.length == 0) return NULL;
        const char *cstr = [name UTF8String];
        if (cstr == NULL) return NULL;
        result = strdup(cstr);
    } @catch (NSException *e) {
        NSLog(@"pausa: frontmostApplication exception: %@", e);
    }
    return result;
}

// Locate a running app by bundle ID first, then localized name.
static NSRunningApplication *findRunningApp(NSString *ident) {
    if (ident.length == 0) return nil;
    NSArray<NSRunningApplication *> *apps =
        [NSRunningApplication runningApplicationsWithBundleIdentifier:ident];
    if (apps.count > 0) return apps.firstObject;
    for (NSRunningApplication *app in [[NSWorkspace sharedWorkspace] runningApplications]) {
        if ([app.localizedName isEqualToString:ident]) return app;
    }
    return nil;
}

int pausa_workspace_activate(const char *identifier) {
    if (identifier == NULL) return 0;
    NSString *ident = NSStr(identifier);
    __block int ok = 0;
    @try {
        NSRunningApplication *app = findRunningApp(ident);
        if (app != nil) {
            ok = [app activateWithOptions:NSApplicationActivateIgnoringOtherApps] ? 1 : 0;
        }
    } @catch (NSException *e) {
        NSLog(@"pausa: activate exception: %@", e);
    }
    return ok;
}

void pausa_app_hide_self(void) {
    onMain(^{
        if (NSApp == nil) return;
        @try {
            // [NSApp hide:nil] is the AppKit equivalent of Cmd-H. It removes
            // Pausa from the screen entirely, which causes macOS to focus
            // whatever app comes next in the activation stack -- without
            // forcing a Space switch first.
            [NSApp hide:nil];
        } @catch (NSException *e) {
            NSLog(@"pausa: hide_self exception: %@", e);
        }
    });
}

int pausa_workspace_activate_url(const char *bundleIdentifier) {
    if (bundleIdentifier == NULL) return 0;
    NSString *ident = NSStr(bundleIdentifier);
    __block int ok = 0;
    @try {
        // Modern path: NSWorkspace can resolve a bundle ID to its URL and
        // activate it via openApplicationAtURL:. This routes through
        // LaunchServices and correctly handles fullscreen-Space apps.
        NSURL *url = [[NSWorkspace sharedWorkspace] URLForApplicationWithBundleIdentifier:ident];
        if (url != nil) {
            NSWorkspaceOpenConfiguration *config = [NSWorkspaceOpenConfiguration configuration];
            config.activates = YES;
            // openApplicationAtURL is asynchronous; we don't wait for the
            // completion handler since it can take 100ms+ on cold launch.
            [[NSWorkspace sharedWorkspace] openApplicationAtURL:url
                                                  configuration:config
                                              completionHandler:^(NSRunningApplication * _Nullable app,
                                                                  NSError * _Nullable error) {
                if (error != nil) {
                    NSLog(@"pausa: activate_url completion error: %@", error);
                }
            }];
            ok = 1;
            return ok;
        }
        // Fallback to the older API if URL lookup failed.
        NSRunningApplication *app = findRunningApp(ident);
        if (app != nil) {
            ok = [app activateWithOptions:NSApplicationActivateIgnoringOtherApps] ? 1 : 0;
        }
    } @catch (NSException *e) {
        NSLog(@"pausa: activate_url exception: %@", e);
    }
    return ok;
}

// Notifications --------------------------------------------------------------
//
// UNUserNotificationCenter requires the process to be inside a code-signed
// app bundle. In `wails dev` the binary lives in /tmp under an arbitrary
// name and the framework throws on access. We detect that case and treat
// notifications as a silent no-op so dev mode stays usable.

static void registerNotifCategoryIfNeeded(void) {
    if (gNotifReady) return;
    @try {
        UNNotificationAction *skip =
            [UNNotificationAction actionWithIdentifier:@"SKIP"
                                                 title:@"Skip"
                                               options:UNNotificationActionOptionDestructive];
        UNNotificationAction *postpone =
            [UNNotificationAction actionWithIdentifier:@"POSTPONE"
                                                 title:@"Postpone"
                                               options:UNNotificationActionOptionNone];
        UNNotificationCategory *cat =
            [UNNotificationCategory categoryWithIdentifier:@"PAUSA_BREAK"
                                                   actions:@[postpone, skip]
                                         intentIdentifiers:@[]
                                                   options:UNNotificationCategoryOptionNone];
        UNUserNotificationCenter *center = [UNUserNotificationCenter currentNotificationCenter];
        [center setNotificationCategories:[NSSet setWithObject:cat]];
        gNotifReady = YES;
    } @catch (NSException *e) {
        NSLog(@"pausa: notif category register exception: %@", e);
    }
}

void pausa_notify_request_auth(void) {
    if (!isBundledApp()) {
        NSLog(@"pausa: not running from .app bundle, notifications disabled");
        return;
    }
    onMain(^{
        @try {
            UNUserNotificationCenter *center = [UNUserNotificationCenter currentNotificationCenter];
            if (gNotifDelegate == nil) {
                gNotifDelegate = [[PausaNotifDelegate alloc] init];
                center.delegate = gNotifDelegate;
            }
            registerNotifCategoryIfNeeded();
            [center requestAuthorizationWithOptions:
                (UNAuthorizationOptionAlert | UNAuthorizationOptionSound)
                completionHandler:^(BOOL granted, NSError * _Nullable error) {
                    (void)granted; (void)error;
                }];
        } @catch (NSException *e) {
            NSLog(@"pausa: requestAuth exception: %@", e);
        }
    });
}

void pausa_notify_post(const char *title, const char *body, int withActions) {
    if (!isBundledApp()) return;
    NSString *t = NSStr(title);
    NSString *b = NSStr(body);
    BOOL actions = (withActions != 0);
    onMain(^{
        @try {
            UNUserNotificationCenter *center = [UNUserNotificationCenter currentNotificationCenter];
            if (gNotifDelegate == nil) {
                gNotifDelegate = [[PausaNotifDelegate alloc] init];
                center.delegate = gNotifDelegate;
            }
            registerNotifCategoryIfNeeded();

            UNMutableNotificationContent *content = [[UNMutableNotificationContent alloc] init];
            content.title = t;
            content.body = b;
            if (actions) content.categoryIdentifier = @"PAUSA_BREAK";

            UNNotificationRequest *req =
                [UNNotificationRequest requestWithIdentifier:[[NSUUID UUID] UUIDString]
                                                     content:content
                                                     trigger:nil];
            [center addNotificationRequest:req withCompletionHandler:nil];
        } @catch (NSException *e) {
            NSLog(@"pausa: notify post exception: %@", e);
        }
    });
}

// Idle -----------------------------------------------------------------------

double pausa_idle_seconds(void) {
    @try {
        CFTimeInterval secs =
            CGEventSourceSecondsSinceLastEventType(kCGEventSourceStateHIDSystemState,
                                                   kCGAnyInputEventType);
        if (secs < 0) return 0.0;
        return (double)secs;
    } @catch (NSException *e) {
        NSLog(@"pausa: idle exception: %@", e);
        return 0.0;
    }
}

// Busy detection -------------------------------------------------------------
//
// We use two complementary signals:
//   1. CoreAudio HAL `kAudioDevicePropertyDeviceIsRunningSomewhere` on the
//      default input device — true when ANY process holds the mic open.
//      This catches every video-call client (Zoom, Meet, Teams, Discord,
//      browser-based meetings, voice memos, system dictation, ...) and
//      requires no microphone permission for Pausa itself.
//
//   2. Apple's private MediaRemote framework, via dlsym. The function
//      MRMediaRemoteGetNowPlayingApplicationIsPlaying(dispatch_queue, block)
//      asynchronously yields a BOOL indicating whether ANY app on the
//      system is currently playing audio/video. Catches YouTube, Netflix,
//      Vimeo, Spotify, Apple Music, QuickTime, Music app, even browser
//      tabs that publish to <audio>/<video>. Loaded at runtime so we
//      degrade gracefully if Apple ever yanks the symbol.

static AudioObjectID defaultInputDevice(void) {
    AudioObjectID dev = kAudioObjectUnknown;
    AudioObjectPropertyAddress addr = {
        .mSelector = kAudioHardwarePropertyDefaultInputDevice,
        .mScope    = kAudioObjectPropertyScopeGlobal,
        .mElement  = kAudioObjectPropertyElementMain,
    };
    UInt32 size = (UInt32)sizeof(dev);
    if (AudioObjectGetPropertyData(kAudioObjectSystemObject, &addr,
                                   0, NULL, &size, &dev) != noErr) {
        return kAudioObjectUnknown;
    }
    return dev;
}

static AudioObjectID defaultOutputDevice(void) {
    AudioObjectID dev = kAudioObjectUnknown;
    AudioObjectPropertyAddress addr = {
        .mSelector = kAudioHardwarePropertyDefaultOutputDevice,
        .mScope    = kAudioObjectPropertyScopeGlobal,
        .mElement  = kAudioObjectPropertyElementMain,
    };
    UInt32 size = (UInt32)sizeof(dev);
    if (AudioObjectGetPropertyData(kAudioObjectSystemObject, &addr,
                                   0, NULL, &size, &dev) != noErr) {
        return kAudioObjectUnknown;
    }
    return dev;
}

int pausa_busy_microphone_active(void) {
    @try {
        AudioObjectID dev = defaultInputDevice();
        if (dev == kAudioObjectUnknown) return 0;

        UInt32 running = 0;
        UInt32 size = (UInt32)sizeof(running);
        AudioObjectPropertyAddress addr = {
            .mSelector = kAudioDevicePropertyDeviceIsRunningSomewhere,
            .mScope    = kAudioObjectPropertyScopeInput,
            .mElement  = kAudioObjectPropertyElementMain,
        };
        if (AudioObjectGetPropertyData(dev, &addr, 0, NULL, &size, &running) != noErr) {
            return 0;
        }
        return running ? 1 : 0;
    } @catch (NSException *e) {
        NSLog(@"pausa: mic_active exception: %@", e);
        return 0;
    }
}

// Cached MediaRemote symbol + framework handle. Loaded on first call.
typedef void (*MRGetNowPlayingIsPlayingFn)(dispatch_queue_t queue,
                                           void (^handler)(Boolean isPlaying));
static MRGetNowPlayingIsPlayingFn gMRGetNowPlaying = NULL;
static dispatch_once_t gMRLoadOnce;
static BOOL gMRLoadFailed = NO;

static void loadMediaRemoteIfNeeded(void) {
    dispatch_once(&gMRLoadOnce, ^{
        // Try the canonical path first; fall back to dlopen by name in
        // case Apple ever moves it.
        const char *paths[] = {
            "/System/Library/PrivateFrameworks/MediaRemote.framework/MediaRemote",
            "MediaRemote.framework/MediaRemote",
            NULL,
        };
        void *handle = NULL;
        for (int i = 0; paths[i] != NULL; i++) {
            handle = dlopen(paths[i], RTLD_LAZY);
            if (handle != NULL) break;
        }
        if (handle == NULL) {
            NSLog(@"pausa: MediaRemote unavailable: %s", dlerror());
            gMRLoadFailed = YES;
            return;
        }
        gMRGetNowPlaying = (MRGetNowPlayingIsPlayingFn)
            dlsym(handle, "MRMediaRemoteGetNowPlayingApplicationIsPlaying");
        if (gMRGetNowPlaying == NULL) {
            NSLog(@"pausa: MediaRemote symbol missing: %s", dlerror());
            gMRLoadFailed = YES;
        }
    });
}

int pausa_busy_now_playing_active(void) {
    loadMediaRemoteIfNeeded();
    if (gMRLoadFailed || gMRGetNowPlaying == NULL) return 0;

    // The MediaRemote API is asynchronous: we ask it on a background queue
    // and the framework calls our block back. We bridge it to a synchronous
    // result via a semaphore with a short timeout so the caller (Go's
    // scheduler tick) doesn't block more than ~50ms.
    __block BOOL result = NO;
    dispatch_semaphore_t sem = dispatch_semaphore_create(0);
    @try {
        gMRGetNowPlaying(dispatch_get_global_queue(QOS_CLASS_USER_INITIATED, 0),
                         ^(Boolean isPlaying) {
            result = isPlaying;
            dispatch_semaphore_signal(sem);
        });
        // 100ms is generous; in practice MediaRemote replies in <5ms.
        dispatch_time_t timeout = dispatch_time(DISPATCH_TIME_NOW, 100 * NSEC_PER_MSEC);
        if (dispatch_semaphore_wait(sem, timeout) != 0) {
            // Timed out — assume not playing rather than block forever.
            return 0;
        }
        return result ? 1 : 0;
    } @catch (NSException *e) {
        NSLog(@"pausa: now_playing exception: %@", e);
        return 0;
    }
}

int pausa_busy_output_active(void) {
    @try {
        AudioObjectID dev = defaultOutputDevice();
        if (dev == kAudioObjectUnknown) return 0;

        UInt32 running = 0;
        UInt32 size = (UInt32)sizeof(running);
        AudioObjectPropertyAddress addr = {
            .mSelector = kAudioDevicePropertyDeviceIsRunningSomewhere,
            .mScope    = kAudioObjectPropertyScopeOutput,
            .mElement  = kAudioObjectPropertyElementMain,
        };
        if (AudioObjectGetPropertyData(dev, &addr, 0, NULL, &size, &running) != noErr) {
            return 0;
        }
        return running ? 1 : 0;
    } @catch (NSException *e) {
        NSLog(@"pausa: output_active exception: %@", e);
        return 0;
    }
}

// Windows --------------------------------------------------------------------

int pausa_screen_count(void) {
    __block int n = 1;
    @try {
        n = (int)[[NSScreen screens] count];
    } @catch (NSException *e) {
        NSLog(@"pausa: screen_count exception: %@", e);
    }
    return n;
}

// Returns the Wails main window. Walks NSApp.windows and skips overlays we
// own and any auxiliary windows (panels, popovers).
static NSWindow *findMainWindow(void) {
    if (NSApp == nil) return nil;
    NSWindow *best = nil;
    for (NSWindow *w in [NSApp windows]) {
        if (w.contentView == nil) continue;
        if (gOverlays != nil && [gOverlays containsObject:w]) continue;
        // Prefer the keyWindow if it qualifies, else the first eligible.
        if (w.isKeyWindow) return w;
        if (best == nil) best = w;
    }
    return best;
}

void pausa_window_bring_to_current_space(void) {
    onMain(^{
        if (NSApp == nil) return;
        @try {
            NSWindow *w = findMainWindow();
            if (w == nil) return;
            w.collectionBehavior =
                NSWindowCollectionBehaviorMoveToActiveSpace |
                NSWindowCollectionBehaviorFullScreenAuxiliary;
            [w makeKeyAndOrderFront:nil];
            [NSApp activateIgnoringOtherApps:YES];
        } @catch (NSException *e) {
            NSLog(@"pausa: bring_to_current_space exception: %@", e);
        }
    });
}

void pausa_app_set_accessory(void) {
    onMain(^{
        if (NSApp == nil) return;
        @try {
            // NSApplicationActivationPolicyAccessory removes the Pausa Dock
            // icon AND lets us float over fullscreen apps. Without this,
            // macOS refuses to show our windows on a fullscreen Space.
            if ([NSApp activationPolicy] != NSApplicationActivationPolicyAccessory) {
                [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
                NSLog(@"pausa: switched to accessory activation policy");
            }
        } @catch (NSException *e) {
            NSLog(@"pausa: set_accessory exception: %@", e);
        }
    });
}

// Color helpers ------------------------------------------------------------

static NSColor *colorFromHex(NSString *hex, CGFloat alpha) {
    if (hex.length == 0) return [NSColor whiteColor];
    if ([hex hasPrefix:@"#"]) hex = [hex substringFromIndex:1];
    if (hex.length != 6) return [NSColor whiteColor];
    unsigned int v = 0;
    [[NSScanner scannerWithString:hex] scanHexInt:&v];
    CGFloat r = ((v >> 16) & 0xFF) / 255.0;
    CGFloat g = ((v >>  8) & 0xFF) / 255.0;
    CGFloat b = ( v        & 0xFF) / 255.0;
    return [NSColor colorWithRed:r green:g blue:b alpha:alpha];
}

// makeStyledLabel returns a non-bordered, non-editable, transparent NSTextField.
static NSTextField *makeStyledLabel(NSRect frame, NSString *text, CGFloat fontSize,
                                    NSFontWeight weight, NSColor *color) {
    NSTextField *l = [[NSTextField alloc] initWithFrame:frame];
    l.stringValue = text ?: @"";
    l.alignment   = NSTextAlignmentCenter;
    l.font        = [NSFont systemFontOfSize:fontSize weight:weight];
    l.textColor   = color;
    l.backgroundColor = [NSColor clearColor];
    l.bordered    = NO;
    l.editable    = NO;
    l.selectable  = NO;
    l.lineBreakMode = NSLineBreakByWordWrapping;
    l.maximumNumberOfLines = 0;
    return l;
}

// makeOverlayButton returns a borderless rounded button suitable for the
// translucent overlay UI.
static NSButton *makeOverlayButton(NSRect frame, NSString *title, int tag,
                                   id target, BOOL primary) {
    NSButton *b = [[NSButton alloc] initWithFrame:frame];
    b.title = title;
    b.target = target;
    b.action = @selector(buttonClicked:);
    b.tag = tag;
    b.bordered = NO;
    b.wantsLayer = YES;
    b.layer.cornerRadius = 12.0;
    b.layer.backgroundColor = primary
        ? [[NSColor whiteColor] colorWithAlphaComponent:0.25].CGColor
        : [[NSColor whiteColor] colorWithAlphaComponent:0.10].CGColor;
    NSMutableAttributedString *attr =
        [[NSMutableAttributedString alloc] initWithString:title];
    [attr addAttributes:@{
        NSForegroundColorAttributeName: [NSColor whiteColor],
        NSFontAttributeName: [NSFont systemFontOfSize:15 weight:NSFontWeightSemibold],
    } range:NSMakeRange(0, title.length)];
    b.attributedTitle = attr;
    return b;
}

// Build the overlay content view. Layout adapts to the panel size: large
// fonts and big spacing for fullscreen panels, tighter compact layout for
// small (~720x520) windowed mode.
static NSView *buildOverlayContent(NSRect frame, NSString *kind, NSString *title,
                                   NSString *timer, NSString *tip,
                                   NSString *accentHex, BOOL showActions,
                                   id buttonTarget) {
    NSView *root = [[NSView alloc] initWithFrame:frame];
    root.wantsLayer = YES;

    // Gradient background derived from accent color.
    NSColor *base = colorFromHex(accentHex, 1.0);
    CGFloat r = 0, g = 0, b = 0, a = 0;
    [[base colorUsingColorSpace:[NSColorSpace sRGBColorSpace]]
        getRed:&r green:&g blue:&b alpha:&a];
    BOOL longBreak = [kind isEqualToString:@"long"];
    CAGradientLayer *grad = [CAGradientLayer layer];
    grad.frame = root.bounds;
    if (longBreak) {
        grad.colors = @[
            (id)[NSColor colorWithRed:0.486 green:0.227 blue:0.929 alpha:1.0].CGColor,
            (id)[NSColor colorWithRed:0.353 green:0.157 blue:0.851 alpha:1.0].CGColor,
            (id)[NSColor colorWithRed:0.243 green:0.105 blue:0.752 alpha:1.0].CGColor,
        ];
    } else {
        grad.colors = @[
            (id)[NSColor colorWithRed:r * 0.55 green:g * 0.55 blue:b * 0.7 alpha:1.0].CGColor,
            (id)[NSColor colorWithRed:r * 0.30 green:g * 0.40 blue:b * 0.55 alpha:1.0].CGColor,
            (id)[NSColor colorWithRed:r * 0.15 green:g * 0.25 blue:b * 0.45 alpha:1.0].CGColor,
        ];
    }
    grad.startPoint = CGPointMake(0, 1);
    grad.endPoint   = CGPointMake(1, 0);
    [root.layer addSublayer:grad];

    // Decide layout proportions based on panel height. Anything under
    // 800pt tall uses the compact layout (typical for non-fullscreen).
    BOOL compact = (frame.size.height < 800);

    CGFloat cx = frame.size.width  / 2;
    CGFloat cy = frame.size.height / 2;

    // Layout constants per mode -------------------------------------------
    CGFloat titleSize     = compact ? 12  : 14;
    CGFloat titleOffsetY  = compact ? 110 : 220;
    CGFloat timerSize     = compact ? 96  : 160;
    CGFloat timerHeight   = compact ? 130 : 200;
    CGFloat timerOffsetY  = compact ? -30 : -80;
    CGFloat subOffsetY    = compact ? -55 : -110;
    CGFloat tipOffsetY    = compact ? -130 : -200;
    CGFloat tipFont       = compact ? 14  : 20;
    CGFloat tipWidth      = compact ? MIN(640, frame.size.width - 60) : 700;
    CGFloat tipHeight     = compact ? 50  : 60;
    CGFloat circleSize    = compact ? 64  : 100;
    CGFloat circleOffsetY = compact ? -190 : -320;
    CGFloat btnY          = compact ? 40  : 80;
    CGFloat btnW          = compact ? 120 : 140;
    CGFloat btnH          = compact ? 38  : 44;

    // Title badge ---------------------------------------------------------
    NSTextField *titleLabel = makeStyledLabel(
        NSMakeRect(cx - 200, cy + titleOffsetY, 400, 32),
        [title uppercaseString], titleSize, NSFontWeightSemibold,
        [[NSColor whiteColor] colorWithAlphaComponent:0.85]);
    [root addSubview:titleLabel];

    // Timer ---------------------------------------------------------------
    NSTextField *timerLabel = makeStyledLabel(
        NSMakeRect(cx - 300, cy + timerOffsetY, 600, timerHeight),
        timer, timerSize, NSFontWeightUltraLight, [NSColor whiteColor]);
    timerLabel.tag = PAUSA_TAG_TIMER;
    [root addSubview:timerLabel];

    // "REMAINING" sub-label ----------------------------------------------
    NSTextField *subLabel = makeStyledLabel(
        NSMakeRect(cx - 200, cy + subOffsetY, 400, 22),
        @"REMAINING", 10, NSFontWeightMedium,
        [[NSColor whiteColor] colorWithAlphaComponent:0.6]);
    [root addSubview:subLabel];

    // Tip -----------------------------------------------------------------
    if (tip.length > 0) {
        NSTextField *tipLabel = makeStyledLabel(
            NSMakeRect(cx - tipWidth / 2, cy + tipOffsetY, tipWidth, tipHeight),
            tip, tipFont, NSFontWeightRegular,
            [[NSColor whiteColor] colorWithAlphaComponent:0.85]);
        [root addSubview:tipLabel];
    }

    // Breathing guide for long breaks ------------------------------------
    if (longBreak) {
        NSView *circle = [[NSView alloc]
            initWithFrame:NSMakeRect(cx - circleSize / 2,
                                     cy + circleOffsetY,
                                     circleSize, circleSize)];
        circle.wantsLayer = YES;
        circle.layer.cornerRadius = circleSize / 2;
        circle.layer.backgroundColor =
            [[NSColor whiteColor] colorWithAlphaComponent:0.15].CGColor;

        CABasicAnimation *pulse = [CABasicAnimation animationWithKeyPath:@"transform.scale"];
        pulse.fromValue = @1.0;
        pulse.toValue   = @1.4;
        pulse.duration  = 4.0;
        pulse.autoreverses = YES;
        pulse.repeatCount  = HUGE_VALF;
        pulse.timingFunction = [CAMediaTimingFunction
            functionWithName:kCAMediaTimingFunctionEaseInEaseOut];
        [circle.layer addAnimation:pulse forKey:@"breathe"];

        if (!compact) {
            NSTextField *breatheLabel = makeStyledLabel(
                NSMakeRect(0, circleSize / 2 - 12, circleSize, 20),
                @"Breathe", 10, NSFontWeightSemibold,
                [[NSColor whiteColor] colorWithAlphaComponent:0.8]);
            [circle addSubview:breatheLabel];
        }
        [root addSubview:circle];
    }

    // Action buttons (Postpone, Skip) ------------------------------------
    if (showActions) {
        CGFloat gap = 16;
        NSButton *postpone = makeOverlayButton(
            NSMakeRect(cx - btnW - gap / 2, btnY, btnW, btnH),
            @"Postpone", PAUSA_OVERLAY_POSTPONE, buttonTarget, YES);
        NSButton *skip = makeOverlayButton(
            NSMakeRect(cx + gap / 2, btnY, btnW, btnH),
            @"Skip", PAUSA_OVERLAY_SKIP, buttonTarget, NO);
        [root addSubview:postpone];
        [root addSubview:skip];
    }

    return root;
}

// Compact-mode overlay dimensions (fullscreen=0). Centered on each screen.
#define PAUSA_COMPACT_WIDTH  720.0
#define PAUSA_COMPACT_HEIGHT 520.0

// activeScreen returns the NSScreen containing the mouse cursor, falling
// back to mainScreen / first screen.
static NSScreen *activeScreen(void) {
    NSPoint p = [NSEvent mouseLocation];
    for (NSScreen *s in [NSScreen screens]) {
        if (NSPointInRect(p, s.frame)) return s;
    }
    NSScreen *m = [NSScreen mainScreen];
    if (m != nil) return m;
    NSArray<NSScreen *> *all = [NSScreen screens];
    return all.count > 0 ? all[0] : nil;
}

// frameForScreen returns the panel frame on a given screen, either edge
// to edge (fullscreen) or a centered compact rect.
static NSRect frameForScreen(NSScreen *screen, BOOL fullscreen) {
    NSRect screenFrame = screen.frame;
    if (fullscreen) return screenFrame;

    CGFloat w = PAUSA_COMPACT_WIDTH;
    CGFloat h = PAUSA_COMPACT_HEIGHT;
    return NSMakeRect(NSMidX(screenFrame) - w / 2,
                      NSMidY(screenFrame) - h / 2,
                      w, h);
}

void pausa_overlays_create(const char *kind, const char *title,
                           const char *timer, const char *tip,
                           const char *hexAccent, int showActions,
                           int fullscreen, int currentScreenOnly) {
    NSString *kindStr   = NSStr(kind);
    NSString *titleStr  = NSStr(title);
    NSString *timerStr  = NSStr(timer);
    NSString *tipStr    = NSStr(tip);
    NSString *accentStr = NSStr(hexAccent);
    BOOL actions     = (showActions != 0);
    BOOL fs          = (fullscreen != 0);
    BOOL currentOnly = (currentScreenOnly != 0);

    onMain(^{
        if (NSApp == nil) {
            NSLog(@"pausa: overlays_create: NSApp nil");
            return;
        }
        @try {
            if (gOverlayTarget == nil) gOverlayTarget = [[PausaOverlayTarget alloc] init];
            if (gOverlays == nil) gOverlays = [NSMutableArray array];
            for (NSWindow *w in gOverlays) [w close];
            [gOverlays removeAllObjects];

            // Decide the set of target screens based on the toggle.
            NSArray<NSScreen *> *targetScreens;
            if (currentOnly) {
                NSScreen *s = activeScreen();
                targetScreens = s ? @[s] : @[];
            } else {
                targetScreens = [NSScreen screens];
            }
            NSLog(@"pausa: overlays_create: %lu target screen(s) (fullscreen=%d, currentOnly=%d)",
                  (unsigned long)targetScreens.count, fs, currentOnly);

            for (NSUInteger i = 0; i < targetScreens.count; i++) {
                NSScreen *screen = targetScreens[i];
                NSRect screenFrame = screen.frame;
                NSRect panelFrame  = frameForScreen(screen, fs);

                NSLog(@"pausa: screen %lu frame=(%g,%g %gx%g) panel=(%g,%g %gx%g)",
                      (unsigned long)i,
                      screenFrame.origin.x, screenFrame.origin.y,
                      screenFrame.size.width, screenFrame.size.height,
                      panelFrame.origin.x, panelFrame.origin.y,
                      panelFrame.size.width, panelFrame.size.height);

                // Build the panel at (0,0) sized to the panel frame; we
                // anchor it explicitly below. Creating the window directly
                // at the global frame can cause it to land on the wrong
                // screen when multiple displays have separate Spaces.
                NSRect contentRect = NSMakeRect(0, 0,
                                                panelFrame.size.width,
                                                panelFrame.size.height);

                // NSPanel + NonactivatingPanel + ScreenSaver level + the
                // right collection behavior is the only combination that
                // reliably overlays a fullscreen-app Space on macOS.
                NSPanel *win = [[NSPanel alloc]
                    initWithContentRect:contentRect
                              styleMask:(NSWindowStyleMaskBorderless |
                                         NSWindowStyleMaskNonactivatingPanel)
                                backing:NSBackingStoreBuffered
                                  defer:NO
                                 screen:screen];
                win.level = NSScreenSaverWindowLevel;
                win.collectionBehavior =
                    NSWindowCollectionBehaviorCanJoinAllSpaces |
                    NSWindowCollectionBehaviorStationary |
                    NSWindowCollectionBehaviorFullScreenAuxiliary |
                    NSWindowCollectionBehaviorIgnoresCycle;
                win.hidesOnDeactivate         = NO;
                win.releasedWhenClosed        = NO;
                win.floatingPanel             = YES;
                win.becomesKeyOnlyIfNeeded    = YES;
                win.worksWhenModal            = YES;
                win.hasShadow                 = !fs; // shadow only in compact mode
                win.movableByWindowBackground = NO;
                win.ignoresMouseEvents        = NO;

                if (fs) {
                    // Edge-to-edge: solid background, no shadow, no rounding.
                    win.opaque = YES;
                    win.backgroundColor = [NSColor blackColor];
                } else {
                    // Compact: transparent panel; the content view's
                    // rounded layer provides the visible surface.
                    win.opaque = NO;
                    win.backgroundColor = [NSColor clearColor];
                }

                // CRITICAL: explicitly place the window at its global
                // panel frame. Without this, NSPanel may snap to (0,0)
                // global coords (= the primary screen) regardless of the
                // `screen:` constructor arg, especially when the target
                // screen has its own active Space.
                [win setFrame:panelFrame display:NO];

                // Build content sized to the *panel*, not the screen, so
                // labels and buttons are laid out correctly in compact mode.
                NSView *root = buildOverlayContent(
                    NSMakeRect(0, 0, panelFrame.size.width, panelFrame.size.height),
                    kindStr, titleStr, timerStr, tipStr, accentStr,
                    actions, gOverlayTarget);

                if (!fs) {
                    // Compact mode: clip the content to a rounded rect so
                    // the panel reads as a card rather than a full screen.
                    root.layer.cornerRadius  = 24.0;
                    root.layer.masksToBounds = YES;
                }

                win.contentView = root;

                [win orderFrontRegardless];

                // Verify the window actually landed on the requested
                // screen. If macOS re-parented it (e.g. because the target
                // screen's current Space rejected the panel), force the
                // frame again after the order operation has settled.
                if (win.screen != screen) {
                    NSLog(@"pausa: panel drifted from screen %lu to %@, re-anchoring",
                          (unsigned long)i, win.screen);
                    [win setFrame:panelFrame display:YES animate:NO];
                }

                [gOverlays addObject:win];
                NSLog(@"pausa: overlay panel placed on screen %lu, final=(%g,%g %gx%g) %@",
                      (unsigned long)i,
                      win.frame.origin.x, win.frame.origin.y,
                      win.frame.size.width, win.frame.size.height,
                      (win.screen == screen ? @"OK" : @"MISMATCH"));
            }
            NSLog(@"pausa: %lu overlays active", (unsigned long)gOverlays.count);
        } @catch (NSException *e) {
            NSLog(@"pausa: overlays_create exception: %@", e);
        }
    });
}

void pausa_overlays_update_timer(const char *timer) {
    NSString *t = NSStr(timer);
    onMain(^{
        if (gOverlays == nil) return;
        @try {
            for (NSWindow *win in gOverlays) {
                for (NSView *v in win.contentView.subviews) {
                    if ([v isKindOfClass:[NSTextField class]] && v.tag == PAUSA_TAG_TIMER) {
                        ((NSTextField *)v).stringValue = t;
                    }
                }
            }
        } @catch (NSException *e) {
            NSLog(@"pausa: overlay update exception: %@", e);
        }
    });
}

void pausa_overlays_close(void) {
    onMain(^{
        if (gOverlays == nil) return;
        @try {
            for (NSWindow *w in gOverlays) [w close];
            [gOverlays removeAllObjects];
        } @catch (NSException *e) {
            NSLog(@"pausa: overlays close exception: %@", e);
        }
    });
}
