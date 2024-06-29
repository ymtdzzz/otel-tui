//go:build !cgo

package component

import (
	"log"
)

func initClipboard() {
	log.Printf("Clipboard can't be used on this platform so clipboard-related feature is disabled. \nSee: https://github.com/golang-design/clipboard/issues/57")
}
