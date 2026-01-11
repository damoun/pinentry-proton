// Package protocol implements the pinentry Assuan protocol.
package protocol

import (
	"fmt"
	"strings"
)

// UnescapeArg decodes percent-encoded arguments
func UnescapeArg(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '%' && i+2 < len(s) {
			// Decode hex
			var b byte
			if _, err := fmt.Sscanf(s[i+1:i+3], "%02x", &b); err == nil {
				result.WriteByte(b)
				i += 3
			} else {
				// If decoding fails, keep the literal '%'
				result.WriteByte(s[i])
				i++
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

// EscapeArg encodes special characters for responses
func EscapeArg(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '%' || c == '\n' || c == '\r' || c < 32 || c > 126 {
			fmt.Fprintf(&result, "%%%02X", c)
		} else {
			result.WriteByte(c)
		}
	}
	return result.String()
}

// PercentEncode encodes bytes for D response
func PercentEncode(data []byte) string {
	var result strings.Builder
	for _, b := range data {
		if b == '%' || b == '\n' || b == '\r' || b < 32 || b > 126 {
			fmt.Fprintf(&result, "%%%02X", b)
		} else {
			result.WriteByte(b)
		}
	}
	return result.String()
}
