package main

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>

static void ykyActiveScreenFrame(double *x, double *w, double *h) {
	NSRect f = [NSScreen mainScreen].frame;
	*x = f.origin.x;
	*w = f.size.width;
	*h = f.size.height;
}
*/
import "C"

import "github.com/wailsapp/wails/v3/pkg/application"

// activeScreen returns the screen holding the frontmost window, so the banner appears where the user is working.
func activeScreen(app *application.App) *application.Screen {
	var x, w, h C.double
	C.ykyActiveScreenFrame(&x, &w, &h)
	for _, s := range app.Screen.GetAll() {
		if s.Bounds.X == int(x) && s.Bounds.Width == int(w) && s.Bounds.Height == int(h) {
			return s
		}
	}
	return nil
}
