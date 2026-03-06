package edge_tts

import (
	"testing"
)

func TestCommunicate(t *testing.T) {
	config := DefaultTTSConfig()
	comm, err := NewCommunicate("hello", config)
	if err != nil {
		t.Fatalf("failed to create communicate: %v", err)
	}

	chunksChan, errChan := comm.Stream()
	audioFound := false
	metaFound := false

	for {
		select {
		case err, ok := <-errChan:
			if ok && err != nil {
				t.Fatalf("stream error: %v", err)
			}
			errChan = nil
		case chunk, ok := <-chunksChan:
			if !ok {
				chunksChan = nil
				goto DONE
			}
			if chunk.Type == "audio" {
				audioFound = true
			} else if chunk.Type == "WordBoundary" {
				metaFound = true
			}
		}
	}
DONE:
	if !audioFound {
		t.Errorf("expected audio chunks")
	}
	if !metaFound {
		t.Errorf("expected WordBoundary metadata chunks")
	}
}
