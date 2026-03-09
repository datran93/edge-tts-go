package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	edge_tts "github.com/rany2/edge-tts-go/pkg/edge_tts"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var text, file, voice, rate, volume, pitch, writeMedia, writeSubtitles string
	var listVoices bool
	var parseMarkdown bool

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.StringVar(&text, "t", "", "what TTS will say")
	flag.StringVar(&text, "text", "", "what TTS will say")
	flag.StringVar(&file, "f", "", "same as --text but read from file")
	flag.StringVar(&file, "file", "", "same as --text but read from file")
	flag.StringVar(&voice, "v", edge_tts.DefaultVoice, "voice for TTS")
	flag.StringVar(&voice, "voice", edge_tts.DefaultVoice, "voice for TTS")
	flag.BoolVar(&listVoices, "l", false, "lists available voices and exits")
	flag.BoolVar(&listVoices, "list-voices", false, "lists available voices and exits")
	flag.BoolVar(&parseMarkdown, "parse-markdown", false, "Parse markdown input to clean plain text for smoother speech")
	flag.StringVar(&rate, "rate", "+0%", "set TTS rate")
	flag.StringVar(&volume, "volume", "+0%", "set TTS volume")
	flag.StringVar(&pitch, "pitch", "+0Hz", "set TTS pitch")
	flag.StringVar(&writeMedia, "write-media", "", "send media output to file instead of stdout")
	flag.StringVar(&writeSubtitles, "write-subtitles", "", "send subtitle output to provided file instead of stderr")

	flag.Parse()

	if listVoices {
		printVoices()
		return nil
	}

	if file != "" {
		if file == "-" || file == "/dev/stdin" {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			text = string(b)
		} else {
			b, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read from file: %w", err)
			}
			text = string(b)
		}
	}

	if text == "" {
		return fmt.Errorf("no text provided")
	}

	if parseMarkdown {
		text = edge_tts.ParseMarkdownToText(text)
	}

	config := edge_tts.TTSConfig{
		Voice:  voice,
		Rate:   rate,
		Volume: volume,
		Pitch:  pitch,
	}

	communicate, err := edge_tts.NewCommunicate(text, config)
	if err != nil {
		return fmt.Errorf("failed to initialize communicate: %w", err)
	}

	var audioOutput io.Writer
	if writeMedia != "" {
		if writeMedia == "-" {
			audioOutput = os.Stdout
		} else {
			f, err := os.Create(writeMedia)
			if err != nil {
				return fmt.Errorf("failed to open media file: %w", err)
			}
			defer f.Close()
			audioOutput = f
		}
	} else {
		audioOutput = io.Discard
	}

	var subOutput io.Writer
	if writeSubtitles != "" {
		if writeSubtitles == "-" {
			subOutput = os.Stderr
		} else {
			f, err := os.Create(writeSubtitles)
			if err != nil {
				return fmt.Errorf("failed to open subtitle file: %w", err)
			}
			defer f.Close()
			subOutput = f
		}
	}

	subMaker := edge_tts.NewSubMaker()

	chunksChan, errChan := communicate.Stream()

	for {
		select {
		case streamErr, ok := <-errChan:
			if !ok {
				errChan = nil
			} else if streamErr != nil {
				return fmt.Errorf("stream error: %w", streamErr)
			}
		case chunk, ok := <-chunksChan:
			if !ok {
				chunksChan = nil
				goto DONE
			}
			if chunk.Type == "audio" {
				audioOutput.Write(chunk.Data)
			} else if chunk.Type == "WordBoundary" || chunk.Type == "SentenceBoundary" {
				subMaker.Feed(chunk)
			}
		}
	}

DONE:
	if subOutput != nil {
		io.WriteString(subOutput, subMaker.GetSRT())
	}
	return nil
}

func printVoices() {
	voices, err := edge_tts.ListVoices()
	if err != nil {
		log.Fatalf("failed to list voices: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tGender\tContentCategories\tVoicePersonalities")
	fmt.Fprintln(w, "----\t------\t-----------------\t------------------")

	for _, v := range voices {
		cats := strings.Join(v.VoiceTag.ContentCategories, ", ")
		pers := strings.Join(v.VoiceTag.VoicePersonalities, ", ")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", v.ShortName, v.Gender, cats, pers)
	}
	w.Flush()
}
