# edge-tts-go Architecture Design

## Overview
This project is a native Go port of the Python `edge-tts` module, allowing users to interact with Microsoft Edge's online text-to-speech service programmatically or via a CLI.

## Packages
The project follows standard Go layout:
* `/cmd/edge-tts`: Provides a command-line interface supporting flags identical to the original Python project (`--text`, `--file`, `--voice`, `--rate`, `--volume`, `--pitch`, `--write-media`, `--write-subtitles`, `--list-voices`).
* `/pkg/edge_tts`: Core library for text synthesis, communication, DRM auth logic, and text splitting.

### Core Modules (`/pkg/edge_tts`)
1. **drm (drm.go)**
   * Manages DRM operations including `Sec-MS-GEC` token generation and system clock skew compensation.
   * Exposes methods to adjust clock skew if encountering HTTP 403 on the websocket or voice list endpoint.
2. **communicate (communicate.go)**
   * Main abstraction boundary for users.
   * `Communicate` type takes parameters for SSML generation.
   * Interfaces to a WebSocket (`wss://`) utilizing `gorilla/websocket` or a similar standard lib to stream audio chunks and metadata (WordBoundary).
   * Parses binary WebSocket frames for mp3 audio and text frames for metadata.
3. **voices (voices.go)**
   * Handles downloading the list of available voices from the REST endpoint, using the DRM token and managing 403 retries.
4. **srt_composer (srt_composer.go)**
   * A unified module to collect metadata and dump it as SubRip text format (`.srt`).
5. **utils / config (utils.go, constants.go)**
   * Constants for HTTP/WSS headers, TrustClientToken.
   * Handlers for text sanitization, SSML chunking to respect 4096-byte limits while keeping valid UTF-8 and XML limits.

## Key Differences from Python
* Concurrency in downloading / writing files uses goroutines/channels instead of `asyncio`.
* Types are strictly defined structs for JSON metadata.
* `spf13/cobra` or standard `flag` will be used for CLI. Standard `flag` or `pflag` is preferred to minimize dependencies.
* `gorilla/websocket` is used since standard library doesn't include a complete WSS client.
