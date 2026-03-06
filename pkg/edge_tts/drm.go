package edge_tts

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

const (
	WinEpoch = 11644473600
	SToNs    = 1e9
)

var (
	clockSkewSeconds float64
	skewMutex        sync.RWMutex
)

// AdjClockSkewSeconds adjusts the clock skew in seconds in case system clock is off.
func AdjClockSkewSeconds(skew float64) {
	skewMutex.Lock()
	defer skewMutex.Unlock()
	clockSkewSeconds += skew
}

// GetUnixTimestamp gets the current timestamp in Unix format with clock skew correction.
func GetUnixTimestamp() float64 {
	skewMutex.RLock()
	skew := clockSkewSeconds
	skewMutex.RUnlock()

	return float64(time.Now().UnixNano())/1e9 + skew
}

// ParseRFC2616Date parses an RFC 2616 date string into a Unix timestamp.
func ParseRFC2616Date(dateStr string) float64 {
	// e.g. "Mon, 02 Jan 2006 15:04:05 MST" in Go format
	t, err := time.Parse(time.RFC1123, dateStr)
	if err != nil {
		return 0
	}
	return float64(t.UnixNano()) / 1e9
}

// HandleClientResponseError adjusts the clock skew based on the server date.
func HandleClientResponseError(serverDate string) error {
	if serverDate == "" {
		return fmt.Errorf("no server date in headers")
	}

	serverDateParsed := ParseRFC2616Date(serverDate)
	if serverDateParsed == 0 {
		return fmt.Errorf("failed to parse server date: %s", serverDate)
	}

	clientDate := GetUnixTimestamp()
	AdjClockSkewSeconds(serverDateParsed - clientDate)
	return nil
}

// GenerateSecMSGEC generates the Sec-MS-GEC token value.
// It generates a token value based on the current time in Windows file time format
// adjusted for clock skew, and rounded down to the nearest 5 minutes.
// The token is hashed using SHA256 and returned as an uppercased hex digest.
func GenerateSecMSGEC() string {
	ticks := GetUnixTimestamp()
	ticks += WinEpoch
	ticks -= math.Mod(ticks, 300)
	ticks *= SToNs / 100

	strToHash := fmt.Sprintf("%.0f%s", ticks, TrustedClientToken)
	hash := sha256.Sum256([]byte(strToHash))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// GenerateMUID generates a random MUID (16 bytes hex).
func GenerateMUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

// HeadersWithMUID returns a copy of the given headers with the MUID header added.
func HeadersWithMUID(headers map[string]string) map[string]string {
	combined := make(map[string]string)
	for k, v := range headers {
		combined[k] = v
	}
	combined["Cookie"] = fmt.Sprintf("muid=%s;", GenerateMUID())
	return combined
}
