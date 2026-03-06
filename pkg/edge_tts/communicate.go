package edge_tts

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type TTSChunk struct {
	Type     string
	Data     []byte
	Offset   int
	Duration int
	Text     string
}

type Communicate struct {
	Config             TTSConfig
	Texts              [][]byte
	OffsetComp         int
	LastDurationOffset int
}

func NewCommunicate(text string, config TTSConfig) (*Communicate, error) {
	if text == "" {
		return nil, errors.New("text is empty")
	}

	escapedCleaned := EscapeText(RemoveIncompatibleCharacters(text))
	escapedSplit := SplitTextByByteLength([]byte(escapedCleaned), 4096)
	return &Communicate{
		Config: config,
		Texts:  escapedSplit,
	}, nil
}

// Stream streams TTS chunks into the provided channel.
func (c *Communicate) Stream() (<-chan TTSChunk, <-chan error) {
	chunksChan := make(chan TTSChunk)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunksChan)
		defer close(errChan)

		for _, textChunk := range c.Texts {
			// Process each text chunk, with 1 retry for 403.
			err := c.processChunkWithRetry(textChunk, chunksChan)
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	return chunksChan, errChan
}

func (c *Communicate) processChunkWithRetry(textChunk []byte, chunksChan chan<- TTSChunk) error {
	err := c.processChunk(textChunk, chunksChan)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 403 {
			// Attempt to handle clock skew based on the server date and retry
			if errSkew := HandleClientResponseError(httpErr.DateHeader); errSkew == nil {
				return c.processChunk(textChunk, chunksChan)
			}
		}
		return err
	}
	return nil
}

func (c *Communicate) processChunk(textChunk []byte, chunksChan chan<- TTSChunk) error {
	reqUrl := fmt.Sprintf("%s&ConnectionId=%s&Sec-MS-GEC=%s&Sec-MS-GEC-Version=%s",
		WssURL, ConnectID(), GenerateSecMSGEC(), SecMSGECVersion)

	headers := http.Header{}
	for k, v := range HeadersWithMUID(WssHeaders) {
		headers.Set(k, v)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	// Note: We bypass URL parsing for wss to pass exact query encoded by fmt.Sprintf.
	u, _ := url.Parse(reqUrl)
	conn, resp, err := dialer.Dial(u.String(), headers)
	if err != nil {
		if resp != nil {
			return &HTTPError{
				StatusCode: resp.StatusCode,
				DateHeader: resp.Header.Get("Date"),
			}
		}
		return err
	}
	defer conn.Close()

	// Send command request
	// In Python: sentenceBoundaryEnabled: false, wordBoundaryEnabled: true (if using WordBoundary)
	cmdReq := fmt.Sprintf("X-Timestamp:%s\r\nContent-Type:application/json; charset=utf-8\r\nPath:speech.config\r\n\r\n%s",
		DateToString(),
		`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":"false","wordBoundaryEnabled":"true"},"outputFormat":"audio-24khz-48kbitrate-mono-mp3"}}}}`)

	if err := conn.WriteMessage(websocket.TextMessage, []byte(cmdReq)); err != nil {
		return fmt.Errorf("failed to send command request: %w", err)
	}

	// Send SSML request
	ssml := MkSSML(c.Config, string(textChunk))
	ssmlReq := SSMLHeadersPlusData(ConnectID(), DateToString(), ssml)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(ssmlReq)); err != nil {
		return fmt.Errorf("failed to send SSML request: %w", err)
	}

	audioWasReceived := false

	// Read messages
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				return err
			}
			break
		}

		if msgType == websocket.TextMessage {
			// Parse headers
			split := bytes.SplitN(msg, []byte("\r\n\r\n"), 2)
			headersTxt := split[0]
			data := bytes.TrimSpace(split[1])

			path := extractHeaderPath(headersTxt)
			if path == "audio.metadata" {
				// Parse metadata
				metaChunks, err := c.parseMetadata(data)
				if err != nil {
					return err
				}
				for _, chunk := range metaChunks {
					chunksChan <- chunk
					c.LastDurationOffset = chunk.Offset + chunk.Duration
				}
			} else if path == "turn.end" {
				c.OffsetComp = c.LastDurationOffset + 8750000 // padding
				break
			}
		} else if msgType == websocket.BinaryMessage {
			if len(msg) < 2 {
				return fmt.Errorf("received binary message too short to contain header length")
			}
			headerLen := binary.BigEndian.Uint16(msg[:2])
			if int(headerLen+2) > len(msg) {
				return fmt.Errorf("binary header length exceeds message length")
			}
			headersTxt := msg[2 : 2+headerLen]
			data := msg[2+headerLen:]

			path := extractHeaderPath(headersTxt)
			if path != "audio" {
				return fmt.Errorf("binary message path must be audio")
			}
			if len(data) == 0 {
				continue
			}
			audioWasReceived = true
			chunksChan <- TTSChunk{
				Type: "audio",
				Data: data,
			}
		}
	}

	if !audioWasReceived {
		return fmt.Errorf("no audio was received. please verify that your parameters are correct")
	}

	return nil
}

func (c *Communicate) parseMetadata(data []byte) ([]TTSChunk, error) {
	var metaResp struct {
		Metadata []struct {
			Type string `json:"Type"`
			Data struct {
				Offset   int `json:"Offset"`
				Duration int `json:"Duration"`
				Text     struct {
					Text string `json:"Text"`
				} `json:"text"`
			} `json:"Data"`
		} `json:"Metadata"`
	}

	if err := json.Unmarshal(data, &metaResp); err != nil {
		return nil, err
	}

	var chunks []TTSChunk
	for _, m := range metaResp.Metadata {
		if m.Type == "WordBoundary" || m.Type == "SentenceBoundary" {
			chunks = append(chunks, TTSChunk{
				Type:     m.Type,
				Offset:   m.Data.Offset + c.OffsetComp,
				Duration: m.Data.Duration,
				Text:     html.UnescapeString(m.Data.Text.Text),
			})
		}
	}

	return chunks, nil
}

func extractHeaderPath(headersTxt []byte) string {
	for _, line := range bytes.Split(headersTxt, []byte("\r\n")) {
		if bytes.HasPrefix(line, []byte("Path:")) {
			return string(bytes.TrimSpace(bytes.TrimPrefix(line, []byte("Path:"))))
		}
	}
	return ""
}
