package edge_tts

import (
	"fmt"
	"strings"
	"time"
)

type SubMaker struct {
	Offset   []int
	Duration []int
	Text     []string
}

func NewSubMaker() *SubMaker {
	return &SubMaker{}
}

func (s *SubMaker) Feed(chunk TTSChunk) {
	if chunk.Type == "WordBoundary" || chunk.Type == "SentenceBoundary" {
		s.Offset = append(s.Offset, chunk.Offset)
		s.Duration = append(s.Duration, chunk.Duration)
		s.Text = append(s.Text, chunk.Text)
	}
}

func formatTime(ticks int) string {
	d := time.Duration(ticks * 100) // 1 tick = 100 ns
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	sec := d / time.Second
	d -= sec * time.Second
	ms := d / time.Millisecond
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, sec, ms)
}

func (s *SubMaker) GetSRT() string {
	var builder strings.Builder
	for i := 0; i < len(s.Text); i++ {
		builder.WriteString(fmt.Sprintf("%d\n", i+1))
		startTime := formatTime(s.Offset[i])
		endTime := formatTime(s.Offset[i] + s.Duration[i])
		builder.WriteString(fmt.Sprintf("%s --> %s\n", startTime, endTime))
		builder.WriteString(s.Text[i] + "\n\n")
	}
	return builder.String()
}
