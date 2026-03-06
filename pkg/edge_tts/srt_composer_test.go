package edge_tts

import (
	"strings"
	"testing"
)

func TestSRTComposer(t *testing.T) {
	sm := NewSubMaker()

	sm.Feed(TTSChunk{
		Type:     "WordBoundary",
		Offset:   1234567, // 123ms
		Duration: 5000000, // 500ms
		Text:     "Hello",
	})

	sm.Feed(TTSChunk{
		Type:     "WordBoundary",
		Offset:   8000000,
		Duration: 2000000,
		Text:     "world",
	})

	srt := sm.GetSRT()

	// Check content
	if !strings.Contains(srt, "1\n") || !strings.Contains(srt, "2\n") {
		t.Errorf("missing entry numbers")
	}
	if !strings.Contains(srt, "Hello\n\n") || !strings.Contains(srt, "world\n\n") {
		t.Errorf("missing text")
	}

	formatted := formatTime(36000000000 + 6000000000 + 10000000 + 1230000) // 1h 10m 1s 123ms
	expected := "01:10:01,123"
	if formatted != expected {
		t.Errorf("expected %s, got %s", expected, formatted)
	}
}
