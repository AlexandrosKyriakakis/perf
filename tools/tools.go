package tools

import (
	"strings"

	"github.com/quickfixgo/quickfix"
)

const (
	soh  = "\x01"
	pipe = "|"
)

// String returns a human-readable FIX string.
func String(msgIn interface{}) string {
	switch msgIn := msgIn.(type) {
	case quickfix.Messagable:
		return strings.ReplaceAll(msgIn.ToMessage().String(), soh, pipe)
	default:
		return strings.ReplaceAll(msgIn.(string), soh, pipe)
	}
}
