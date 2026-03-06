package edge_tts

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// RemoveIncompatibleCharacters replaces unsupported characters with space.
func RemoveIncompatibleCharacters(text string) string {
	var builder strings.Builder
	builder.Grow(len(text))
	for _, char := range text {
		if (0 <= char && char <= 8) || (11 <= char && char <= 12) || (14 <= char && char <= 31) {
			builder.WriteRune(' ')
		} else {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func ConnectID() string {
	id := uuid.New().String()
	return strings.ReplaceAll(id, "-", "")
}

func findLastNewlineOrSpaceWithinLimit(text []byte, limit int) int {
	sub := text
	if len(text) > limit {
		sub = text[:limit]
	}
	idx := bytes.LastIndexByte(sub, '\n')
	if idx < 0 {
		idx = bytes.LastIndexByte(sub, ' ')
	}
	return idx
}

func findSafeUTF8SplitPoint(text []byte) int {
	splitAt := len(text)
	for splitAt > 0 {
		if utf8.Valid(text[:splitAt]) {
			return splitAt
		}
		splitAt--
	}
	return splitAt
}

func adjustSplitPointForXMLEntity(text []byte, splitAt int) int {
	for splitAt > 0 && bytes.Contains(text[:splitAt], []byte("&")) {
		ampersandIndex := bytes.LastIndex(text[:splitAt], []byte("&"))
		if bytes.Contains(text[ampersandIndex:splitAt], []byte(";")) {
			break
		}
		splitAt = ampersandIndex
	}
	return splitAt
}

// SplitTextByByteLength splits text into chunks prioritizing newlines, spaces, utf8 and escaping XML.
func SplitTextByByteLength(text []byte, byteLength int) [][]byte {
	if byteLength <= 0 {
		return nil
	}

	var chunks [][]byte
	for len(text) > byteLength {
		splitAt := findLastNewlineOrSpaceWithinLimit(text, byteLength)
		if splitAt < 0 {
			splitAt = findSafeUTF8SplitPoint(text[:byteLength])
		}
		splitAt = adjustSplitPointForXMLEntity(text, splitAt)
		if splitAt < 0 {
			splitAt = len(text)
		}

		chunk := bytes.TrimSpace(text[:splitAt])
		if len(chunk) > 0 {
			chunks = append(chunks, chunk)
		}
		if splitAt > 0 {
			text = text[splitAt:]
		} else {
			text = text[1:]
		}
	}
	remaining := bytes.TrimSpace(text)
	if len(remaining) > 0 {
		chunks = append(chunks, remaining)
	}
	return chunks
}

func MkSSML(tc TTSConfig, escapedText string) string {
	return fmt.Sprintf(
		`<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='en-US'>`+
			`<voice name='%s'>`+
			`<prosody pitch='%s' rate='%s' volume='%s'>`+
			`%s`+
			`</prosody>`+
			`</voice>`+
			`</speak>`, tc.Voice, tc.Pitch, tc.Rate, tc.Volume, escapedText)
}

func EscapeText(text string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(text))
	return buf.String()
}

func DateToString() string {
	// JS-style date string
	// e.g. "Thu Oct 01 2020 00:00:00 GMT+0000 (Coordinated Universal Time)"
	dt := time.Now().UTC()
	return dt.Format("Mon Jan 02 2006 15:04:05 GMT+0000 (Coordinated Universal Time)")
}

func SSMLHeadersPlusData(requestID, timestamp, ssml string) string {
	return fmt.Sprintf("X-RequestId:%s\r\nContent-Type:application/ssml+xml\r\nX-Timestamp:%sZ\r\nPath:ssml\r\n\r\n%s",
		requestID, timestamp, ssml)
}
