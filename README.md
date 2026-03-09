# edge-tts-go

`edge-tts-go` is a native Golang port of the popular [edge-tts](https://github.com/rany2/edge-tts) Python library. It allows you to use Microsoft Edge's online text-to-speech service directly from your Go code or via a highly performant command-line interface.

Because it is written in Go, it features extremely fast concurrency, zero Python runtime dependencies, and the ability to be compiled into a single static binary for any platform.

## Features
* **No API Key Required**: Fully utilizes the Microsoft Edge text-to-speech API anonymously using the latest DRM token schemes.
* **CLI & Package**: Can be used both as a CLI tool and imported into other Go apps as `pkg/edge_tts`.
* **Subtitle Generation**: Automatically generates SubRip Subtitle (`.srt`) files with precise word-boundary timestamps.
* **Audio adjustments**: Support overriding pitch, volume, and rate of the output audio.

## Installation / Building

Make sure you have Go 1.21+ installed on your machine.

Clone the repository and build the binary:

```bash
git clone https://github.com/datran93/edge-tts-go.git
cd edge-tts-go
go build -o bin/edge-tts ./cmd/edge-tts
```

*This will output the compiled binary to `bin/edge-tts`.*

## Command Line Usage

The CLI arguments match the original `edge-tts` tool closely. 

### Basic Usage
Convert a string of text to an mp3 file and corresponding subtitle file:

```bash
./bin/edge-tts --text 'Hello, world! This is a test from Golang.' --write-media hello.mp3 --write-subtitles hello.srt
```

### Read from a file
Instead of providing text inline, you can feed a text file:

```bash
./bin/edge-tts -f my_book.txt --write-media book.mp3
```

### Changing the Voice
You can change the voice used by the text-to-speech service by using the `--voice` option. The `--list-voices` option can be used to list all available voices.

**List voices:**
```bash
./bin/edge-tts --list-voices
```

**Use a specific voice:**
```bash
./bin/edge-tts --voice vi-VN-HoaiMyNeural --text 'Chào mừng bạn đến với phiên bản Golang siêu tốc của thư viện edge-tts' --write-media vi_test.mp3
```

### Adjusting Rate, Volume, and Pitch
You can change the rate, volume, and pitch of the generated speech. 

```bash
./bin/edge-tts --rate="-50%" --text "Speaking slowly." --write-media slow.mp3
./bin/edge-tts --volume="+50%" --text "Speaking louder." --write-media loud.mp3
./bin/edge-tts --pitch="-10Hz" --text "Deep voice." --write-media deep.mp3
```

### Direct Playback (Piping)
If you do not want to save the mp3 file to disk and just want to listen to the synthesized speech immediately, you can use the `-` option with `--write-media` to output binary media to standard output (stdout). You can pipe this directly into a media player like `mpv` or `ffplay`.

```bash
./bin/edge-tts --text "Hello, playing this live from the terminal!" --write-media - | mpv -
```

### Parsing Markdown
If you are passing a large markdown document and want a smooth, uninterrupted reading experience without reading code blocks, headers, or blockquotes, use the `--parse-markdown` flag. It strips markdown artifacts and replaces all line breaks with spaces so the TTS engine speaks fluidly.

```bash
./bin/edge-tts --parse-markdown -f README.md --write-media docs.mp3
```

## Using as a Go Module

You can easily embed `edge-tts-go` into your own Go applications. 

```go
package main

import (
	"fmt"
	"log"

	"github.com/datran93/edge-tts-go/pkg/edge_tts"
)

func main() {
	text := "Hello from Go!"
	
	// Set configuration
	config := edge_tts.TTSConfig{
		Voice:  "en-US-AriaNeural",
		Rate:   "+0%",
		Volume: "+0%",
		Pitch:  "+0Hz",
	}

	// Initialize the communicator
	comm, err := edge_tts.NewCommunicate(text, config)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Stream handles audio data strings and metadata back to you
	chunksChan, errChan := comm.Stream()
	
	subMaker := edge_tts.NewSubMaker()

	for {
		select {
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else if err != nil {
				log.Fatalf("Stream error: %v", err)
			}
		case chunk, ok := <-chunksChan:
			if !ok {
				chunksChan = nil
				goto DONE
			}
			
			if chunk.Type == "audio" {
				// Process or write mp3 bytes -> chunk.Data
				fmt.Printf("Received %d bytes of audio\n", len(chunk.Data))
			} else if chunk.Type == "WordBoundary" {
				// Feed metadata for Subtitles
				subMaker.Feed(chunk)
			}
		}
	}

DONE:
	fmt.Println("Completed!")
	fmt.Println("SRT Data:", subMaker.GetSRT())
}
```

## Disclaimer
This project is an unofficial port of the Python `edge-tts` and relies on the Microsoft Edge Read Aloud API. Microsoft may change the API or the DRM token protocols at any point which may impact functionality. 
