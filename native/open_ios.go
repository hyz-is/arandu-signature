//go:build ios

package native

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework UIKit -framework Foundation
#import <UIKit/UIKit.h>

// openURL asks UIKit to open the address, on the main queue where UIKit
// expects to be called. It answers 0 when the address parses, and the system
// decides the rest: Safari, or the app that claims the link.
static int openURL(const char *address) {
	NSString *text = [NSString stringWithUTF8String:address];
	NSURL *url = [NSURL URLWithString:text];
	if (url == nil) {
		return -1;
	}
	dispatch_async(dispatch_get_main_queue(), ^{
		[[UIApplication sharedApplication] openURL:url options:@{} completionHandler:nil];
	});
	return 0;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

// OpenURL hands the address to UIKit, which opens it in Safari or in the app
// that claims it. Ayra exposes no handle for this, and none is needed: the
// shared UIApplication is the process's.
func OpenURL(url string) error {
	address := C.CString(url)
	defer C.free(unsafe.Pointer(address))
	if C.openURL(address) != 0 {
		return errors.New("signature: not an address UIKit can open")
	}
	return nil
}
