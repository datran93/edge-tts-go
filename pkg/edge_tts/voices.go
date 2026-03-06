package edge_tts

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Voice struct {
	Name           string   `json:"Name"`
	ShortName      string   `json:"ShortName"`
	Gender         string   `json:"Gender"`
	Locale         string   `json:"Locale"`
	SuggestedCodec string   `json:"SuggestedCodec"`
	FriendlyName   string   `json:"FriendlyName"`
	Status         string   `json:"Status"`
	VoiceTag       VoiceTag `json:"VoiceTag"`
}

type VoiceTag struct {
	ContentCategories  []string `json:"ContentCategories"`
	VoicePersonalities []string `json:"VoicePersonalities"`
}

// ListVoices fetches all available voices.
// It handles a 403 response to adjust DRM clock skew and retry.
func ListVoices() ([]Voice, error) {
	voices, err := fetchVoices()
	if err != nil {
		if stringsErr, ok := err.(*HTTPError); ok && stringsErr.StatusCode == 403 {
			// Try handling the skew issue and retry once.
			if errSkew := HandleClientResponseError(stringsErr.DateHeader); errSkew == nil {
				return fetchVoices()
			}
		}
		return nil, err
	}
	return voices, nil
}

type HTTPError struct {
	StatusCode int
	DateHeader string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

func fetchVoices() ([]Voice, error) {
	url := fmt.Sprintf("%s&Sec-MS-GEC=%s&Sec-MS-GEC-Version=%s", VoiceListURL, GenerateSecMSGEC(), SecMSGECVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range HeadersWithMUID(VoiceHeaders) {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			DateHeader: resp.Header.Get("Date"),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var voices []Voice
	if err := json.Unmarshal(body, &voices); err != nil {
		return nil, err
	}

	// Make sure nested struct allocations exist
	for i := range voices {
		if voices[i].VoiceTag.ContentCategories == nil {
			voices[i].VoiceTag.ContentCategories = []string{}
		}
		if voices[i].VoiceTag.VoicePersonalities == nil {
			voices[i].VoiceTag.VoicePersonalities = []string{}
		}
	}

	return voices, nil
}
