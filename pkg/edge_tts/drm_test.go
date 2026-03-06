package edge_tts

import (
	"strings"
	"testing"
)

func TestParseRFC2616Date(t *testing.T) {
	d := ParseRFC2616Date("Wed, 21 Oct 2015 07:28:00 GMT")
	if d == 0 {
		t.Errorf("failed to parse date")
	}

	d2 := ParseRFC2616Date("Invalid date")
	if d2 != 0 {
		t.Errorf("expected 0 for invalid date, got %f", d2)
	}
}

func TestGenerateSecMSGEC(t *testing.T) {
	gec := GenerateSecMSGEC()
	if len(gec) != 64 {
		t.Errorf("expected length 64 (sha256 hex string), got %d", len(gec))
	}
	if strings.ToUpper(gec) != gec {
		t.Errorf("expected upper case hex")
	}
}

func TestGenerateMUID(t *testing.T) {
	muid := GenerateMUID()
	if len(muid) != 32 {
		t.Errorf("expected length 32 for muid, got %d", len(muid))
	}
}

func TestHeadersWithMUID(t *testing.T) {
	h := map[string]string{
		"Key": "Value",
	}
	res := HeadersWithMUID(h)
	if res["Key"] != "Value" {
		t.Errorf("missing key")
	}
	if !strings.HasPrefix(res["Cookie"], "muid=") {
		t.Errorf("missing muid cookie")
	}
}

func TestAdjClockSkew(t *testing.T) {
	prev := GetUnixTimestamp()
	AdjClockSkewSeconds(500)
	next := GetUnixTimestamp()

	if next-prev < 500 {
		t.Errorf("clock skew not adjusted correctly")
	}
}
