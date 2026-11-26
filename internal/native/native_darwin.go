//go:build darwin

package native

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
#include <stdint.h>

static int popoverVisible(uintptr_t p) {
  return ((NSWindow *)p).isVisible ? 1 : 0;
}

static int popoverFocused(uintptr_t p) {
  return ((NSWindow *)p).isKeyWindow ? 1 : 0;
}

static NSButton *findButton(NSView *v) {
  if ([v isKindOfClass:[NSButton class]]) return (NSButton *)v;
  for (NSView *s in v.subviews) {
    NSButton *b = findButton(s);
    if (b) return b;
  }
  return nil;
}

// Marca a imagem do item da barra de menus como "template" (adapta ao
// tema claro/escuro), como icon.setTemplateImage(true) no Electron.
static void trayMakeTemplate(void) {
  @try {
    for (NSWindow *win in [NSApp windows]) {
      if (![NSStringFromClass([win class]) containsString:@"StatusBar"]) continue;
      NSButton *b = findButton(win.contentView);
      if (b && b.image) {
        NSImage *img = [b.image copy];
        [img setSize:NSMakeSize(20.0, 20.0)];
        [img setTemplate:YES];
        b.image = img;
      }
    }
  } @catch (NSException *e) {
    NSLog(@"look-news: trayMakeTemplate: %@", e);
  }
}

static void appActivate(void) {
  [NSApp activateIgnoringOtherApps:YES];
}

// Não colocar ícone no Dock nem no Cmd+Tab.
static void popoverInit(uintptr_t p) {
  @try {
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];

    NSWindow *w = (NSWindow *)p;
    [w setHasShadow:YES];
    [w setLevel:NSPopUpMenuWindowLevel];
    [w setCollectionBehavior:NSWindowCollectionBehaviorCanJoinAllSpaces |
                             NSWindowCollectionBehaviorFullScreenAuxiliary];

    w.backgroundColor = [NSColor windowBackgroundColor];
    NSView *content = w.contentView;
    content.wantsLayer = YES;
    content.layer.cornerRadius = 12.0;
    content.layer.masksToBounds = YES;
  } @catch (NSException *e) {
    NSLog(@"look-news: popoverInit: %@", e);
  }
}

// Retorna a altura da Dock sempre como se ela estivesse visível.
// Quando a Dock está com auto-hide, visibleFrame só desconta ~4pt.
// Nesse caso lemos tilesize (+ largesize se magnification estiver ligado).
static CGFloat dockHeightAlways(NSScreen *screen) {
  if (!screen) return 0;

  NSRect frame = screen.frame;
  NSRect visible = screen.visibleFrame;

  CGFloat current = 0;

  // Dock embaixo
  if (visible.origin.y > frame.origin.y + 0.5) {
    current = visible.origin.y - frame.origin.y;
  }
  // Dock na esquerda
  else if (visible.origin.x > frame.origin.x + 0.5) {
    current = visible.origin.x - frame.origin.x;
  }
  // Dock na direita
  else if (visible.size.width < frame.size.width - 0.5) {
    current = frame.size.width - visible.size.width;
  }

  // Se a diferença for pequena → Dock está escondida
  if (current < 25.0) {
    NSUserDefaults *defaults = [[NSUserDefaults alloc] initWithSuiteName:@"com.apple.dock"];
    CGFloat tile = [defaults doubleForKey:@"tilesize"];
    if (tile < 16.0) tile = 48.0; // fallback seguro

    if ([defaults boolForKey:@"magnification"]) {
      CGFloat large = [defaults doubleForKey:@"largesize"];
      if (large > tile) {
        tile = (tile + large) * 0.5; // média razoável
      }
    }

    // padding vertical típico da Dock
    return tile + 16.0;
  }

  return current;
}

static void popoverPlace(uintptr_t p) {
  @try {
    NSWindow *w = (NSWindow *)p;
    NSPoint mouse = [NSEvent mouseLocation]; // origem canto inferior esquerdo

    // Tela sob o mouse
    NSScreen *scr = nil;
    for (NSScreen *s in [NSScreen screens]) {
      if (NSPointInRect(mouse, s.frame)) {
        scr = s;
        break;
      }
    }
    if (!scr) scr = [NSScreen mainScreen];

    NSRect frame = scr.frame;
    NSRect visible = scr.visibleFrame;
    NSRect wf = w.frame;

    CGFloat gap = 36;      // um pouco mais de espaço abaixo da barra
    CGFloat margin = 12;
    CGFloat dockH = dockHeightAlways(scr);

    // X: centralizado no mouse, limitado à área visível
    CGFloat x = mouse.x - wf.size.width / 2.0;
    x = MAX(NSMinX(visible) + margin, MIN(x, NSMaxX(visible) - wf.size.width - margin));

    CGFloat y;

    // Mouse na metade superior da tela → abre abaixo da barra de menus
    if (mouse.y > NSMidY(frame)) {
      // Usamos o topo da área visível (já desconta a menu bar)
      // e ainda empurramos um pouco mais para baixo
      y = NSMaxY(visible) - wf.size.height - gap;

      // Garante que nunca invada a área da Dock (mesmo quando escondida)
      CGFloat minY = NSMinY(frame) + dockH + 8;
      if (y < minY) {
        y = minY;
      }
    } else {
      // Mouse embaixo → abre acima da Dock
      y = NSMinY(visible) + gap;

      // Se a Dock estiver escondida, visible.minY é quase 0.
      // Forçamos a altura real da Dock.
      CGFloat minY = NSMinY(frame) + dockH + gap;
      if (y < minY) {
        y = minY;
      }
    }

    [w setFrameOrigin:NSMakePoint(x, y)];
    [NSApp activateIgnoringOtherApps:YES];
    [w makeKeyAndOrderFront:nil];
  } @catch (NSException *e) {
    NSLog(@"look-news: popoverPlace: %@", e);
  }
}
*/
import "C"

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

func nsWindowOf(ctx any) uintptr {
	if c, ok := ctx.(driver.MacWindowContext); ok {
		return c.NSWindow
	}
	return 0
}

func nativeVisible(ctx any) (bool, bool) {
	p := nsWindowOf(ctx)
	if p == 0 {
		return false, false
	}
	return C.popoverVisible(C.uintptr_t(p)) != 0, true
}

func nativeFocused(ctx any) bool {
	return C.popoverFocused(C.uintptr_t(nsWindowOf(ctx))) != 0
}

func nativeInit(ctx any) {
	if p := nsWindowOf(ctx); p != 0 {
		C.popoverInit(C.uintptr_t(p))
	}
}

func nativePlace(ctx any) {
	if p := nsWindowOf(ctx); p != 0 {
		C.popoverPlace(C.uintptr_t(p))
	}
}

// ActivateApp brings the app to the foreground for its secondary windows.
func ActivateApp() { C.appActivate() }

// MarkTrayIconTemplate lets macOS adapt the tray icon to the menu bar theme.
func MarkTrayIconTemplate() { C.trayMakeTemplate() }

// StartPopoverPositioner reproduces the Electron popover behavior:
//   - ao aparecer, posiciona junto ao tray (mac: abaixo da barra de menus;
//     Windows: acima da barra de tarefas);
//   - ao perder o foco (blur), esconde;
//   - clicar no ícone com o popup aberto fecha (sem reabrir em seguida).
//
// Tudo roda na thread principal (AppKit exige) via fyne.DoAndWait.
// As partes nativas ficam em popover_<os>.go.
const (
	focusGrace              = 400 * time.Millisecond // tempo para a janela ganhar foco antes de checar o blur
	toggleGrace             = 500 * time.Millisecond // clique no tray logo após um blur = "fechar"
	popoverIdlePollInterval = 120 * time.Millisecond
	popoverOpenPollInterval = 40 * time.Millisecond
)

func StartPopoverPositioner(w fyne.Window) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return
	}
	go func() {
		var (
			inited                bool
			wasVisible            bool
			shownAt, blurHiddenAt time.Time
		)
		timer := time.NewTimer(popoverIdlePollInterval)
		defer timer.Stop()
		for range timer.C {
			visibleNow := false
			fyne.DoAndWait(
				func() {
					nw.RunNative(
						func(ctx any) {
							visible, ok := nativeVisible(ctx)
							if !ok {
								return
							}
							if !inited {
								nativeInit(ctx)
								inited = true
							}

							if visible && !wasVisible {
								// O clique no ícone causou blur (esconde) e logo depois o
								// toggle do tray tentou mostrar de novo: trata como "fechar".
								if time.Since(blurHiddenAt) < toggleGrace {
									w.Hide()
									wasVisible = false
									visibleNow = false
									return
								}
								nativePlace(ctx)
								shownAt = time.Now()
							}

							if visible && time.Since(shownAt) > focusGrace && !nativeFocused(ctx) {
								w.Hide()
								blurHiddenAt = time.Now()
								visible = false
							}
							wasVisible = visible
							visibleNow = visible
						},
					)
				},
			)
			interval := popoverIdlePollInterval
			if visibleNow {
				interval = popoverOpenPollInterval
			}
			timer.Reset(interval)
		}
	}()
}
