//go:build js

package native

import "syscall/js"

// OpenURL opens the address in a new tab of the browser the app runs in.
func OpenURL(url string) error {
	js.Global().Call("open", url, "_blank", "noopener")
	return nil
}
