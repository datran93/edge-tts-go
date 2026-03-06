package edge_tts

import (
	"fmt"
)

const (
	BaseURL            = "speech.platform.bing.com/consumer/speech/synthesize/readaloud"
	TrustedClientToken = "6A5AA1D4EAFF4E9FB37E23D68491D6F4"
	DefaultVoice       = "en-US-EmmaMultilingualNeural"

	ChromiumFullVersion  = "143.0.3650.75"
	ChromiumMajorVersion = "143"
)

var (
	WssURL          = fmt.Sprintf("wss://%s/edge/v1?TrustedClientToken=%s", BaseURL, TrustedClientToken)
	VoiceListURL    = fmt.Sprintf("https://%s/voices/list?trustedclienttoken=%s", BaseURL, TrustedClientToken)
	SecMSGECVersion = fmt.Sprintf("1-%s", ChromiumFullVersion)
)

var BaseHeaders = map[string]string{
	"User-Agent": fmt.Sprintf(
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36 Edg/%s.0.0.0",
		ChromiumMajorVersion, ChromiumMajorVersion),
	"Accept-Encoding": "gzip, deflate, br, zstd",
	"Accept-Language": "en-US,en;q=0.9",
}

var WssHeaders = func() map[string]string {
	h := map[string]string{
		"Pragma":        "no-cache",
		"Cache-Control": "no-cache",
		"Origin":        "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold",
	}
	for k, v := range BaseHeaders {
		h[k] = v
	}
	return h
}()

var VoiceHeaders = func() map[string]string {
	h := map[string]string{
		"Authority":        "speech.platform.bing.com",
		"Sec-CH-UA":        fmt.Sprintf(`" Not;A Brand";v="99", "Microsoft Edge";v="%s", "Chromium";v="%s"`, ChromiumMajorVersion, ChromiumMajorVersion),
		"Sec-CH-UA-Mobile": "?0",
		"Accept":           "*/*",
		"Sec-Fetch-Site":   "none",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Dest":   "empty",
	}
	for k, v := range BaseHeaders {
		h[k] = v
	}
	return h
}()

type TTSConfig struct {
	Voice  string
	Rate   string
	Volume string
	Pitch  string
}

func DefaultTTSConfig() TTSConfig {
	return TTSConfig{
		Voice:  DefaultVoice,
		Rate:   "+0%",
		Volume: "+0%",
		Pitch:  "+0Hz",
	}
}
