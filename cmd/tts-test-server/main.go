package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	edge_tts "github.com/rany2/edge-tts-go/pkg/edge_tts"
)

//go:embed static
var staticFiles embed.FS

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	// Serve static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("failed to create sub filesystem: %v", err)
	}
	http.Handle("/", http.FileServer(http.FS(staticFS)))

	// TTS streaming endpoint
	http.HandleFunc("/api/tts", handleTTS)

	// List voices endpoint
	http.HandleFunc("/api/voices", handleVoices)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("🎵 TTS Test Server running at http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleTTS(w http.ResponseWriter, r *http.Request) {
	text := r.URL.Query().Get("text")
	if text == "" {
		http.Error(w, "missing 'text' query parameter", http.StatusBadRequest)
		return
	}

	voice := r.URL.Query().Get("voice")
	if voice == "" {
		voice = edge_tts.DefaultVoice
	}

	rate := r.URL.Query().Get("rate")
	if rate == "" {
		rate = "+0%"
	}

	// CORS headers for frontend fetch
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	config := edge_tts.TTSConfig{
		Voice:  voice,
		Rate:   rate,
		Volume: "+0%",
		Pitch:  "+0Hz",
	}

	communicate, err := edge_tts.NewCommunicate(text, config)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to init TTS: %v", err), http.StatusInternalServerError)
		return
	}

	// Set headers for chunked streaming
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	chunksChan, errChan := communicate.Stream()

	chunkCount := 0
	totalBytes := 0

	for {
		select {
		case streamErr, ok := <-errChan:
			if !ok {
				errChan = nil
			} else if streamErr != nil {
				log.Printf("TTS stream error: %v", streamErr)
				return
			}
		case chunk, ok := <-chunksChan:
			if !ok {
				log.Printf("TTS complete: %d chunks, %d bytes total", chunkCount, totalBytes)
				return
			}
			if chunk.Type == "audio" {
				n, writeErr := w.Write(chunk.Data)
				if writeErr != nil {
					log.Printf("Write error (client disconnected?): %v", writeErr)
					return
				}
				flusher.Flush()
				chunkCount++
				totalBytes += n
			}
		}
	}
}

func handleVoices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	voices, err := edge_tts.ListVoices()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list voices: %v", err), http.StatusInternalServerError)
		return
	}

	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range voices {
		if i > 0 {
			sb.WriteString(",")
		}
		fmt.Fprintf(&sb, `{"name":"%s","gender":"%s","locale":"%s"}`, v.ShortName, v.Gender, v.Locale)
	}
	sb.WriteString("]")

	w.Write([]byte(sb.String()))
}
