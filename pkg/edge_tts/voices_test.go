package edge_tts

import (
	"testing"
)

func TestVoices(t *testing.T) {
	voices, err := ListVoices()
	if err != nil {
		t.Fatalf("failed to list voices: %v", err)
	}

	if len(voices) == 0 {
		t.Errorf("expected to find voices")
	}
}
