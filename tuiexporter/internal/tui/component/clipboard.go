//go:build cgo

package component

import (
	"log"

	"golang.design/x/clipboard"
)

func initClipboard() {
	if err := clipboard.Init(); err != nil {
		log.Printf("Clipboard can't be used on this platform so clipboard-related feature is disabled. err: %v", err)
	} else {
		log.Println("Clipboard initialization succeeded")
	}
}
