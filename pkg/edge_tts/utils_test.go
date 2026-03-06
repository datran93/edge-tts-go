package edge_tts

import (
	"strings"
	"testing"
)

func TestRemoveIncompatibleCharacters(t *testing.T) {
	input := "Hello\x0bWorld" // \x0b is vertical tab
	expected := "Hello World"
	result := RemoveIncompatibleCharacters(input)
	if result != expected {
		t.Errorf("expected %q but got %q", expected, result)
	}
}

func TestSplitTextByByteLength(t *testing.T) {
	text := []byte("Hello world, this is a very long string that should be split at a space or newline safely.")
	chunks := SplitTextByByteLength(text, 20)
	if len(chunks) == 0 {
		t.Fatalf("expected chunks, got 0")
	}

	for _, c := range chunks {
		if len(c) > 20 {
			t.Errorf("chunk %q exceeds 20 bytes", c)
		}
	}
}

func TestAdjustSplitPointForXMLEntity(t *testing.T) {
	text := []byte("This &amp; that")
	split := adjustSplitPointForXMLEntity(text, 8) // splits between & and ;
	if split != 5 {                                // should split at &
		t.Errorf("expected split at 5, got %d", split)
	}
}

func TestMkSSML(t *testing.T) {
	cfg := DefaultTTSConfig()
	escapedText := "Hello &amp; World"
	ssml := MkSSML(cfg, escapedText)
	if !strings.Contains(ssml, escapedText) {
		t.Errorf("ssml does not contain escaped text")
	}
	if !strings.Contains(ssml, cfg.Voice) {
		t.Errorf("ssml does not contain voice")
	}
}

func TestEscapeText(t *testing.T) {
	res := EscapeText("A & B < C")
	expected := "A &amp; B &lt; C"
	if res != expected {
		t.Errorf("expected %q, got %q", expected, res)
	}
}

func TestDateToString(t *testing.T) {
	d := DateToString()
	if !strings.Contains(d, "GMT+0000") {
		t.Errorf("date %q doesn't contain GMT+0000", d)
	}
}

func TestExtractHeaderPath(t *testing.T) {
	hdrs := []byte("Content-Type: application/json\r\nPath: audio.metadata\r\n")
	p := extractHeaderPath(hdrs)
	if p != "audio.metadata" {
		t.Errorf("expected audio.metadata, got %q", p)
	}
}
