// Package transcode provides video transcoding utilities backed by FFmpeg.
// FFmpeg must be installed on the host system.
package transcode

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// VideoInfo holds basic metadata about a video file.
type VideoInfo struct {
	Duration  float64 // seconds
	Width     int
	Height    int
	CodecName string
}

// ToMP4 transcodes any video file to MP4 (H.264 video + AAC audio) with the
// faststart flag so that the moov atom is placed at the beginning of the file,
// enabling progressive download / streaming.
//
// The caller is responsible for creating and cleaning up temporary files.
func ToMP4(ctx context.Context, inputPath, outputPath string) error {
	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		"-y",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg transcode: %w: %s", err, string(output))
	}
	return nil
}

// Probe returns basic video metadata using ffprobe.
func Probe(ctx context.Context, inputPath string) (*VideoInfo, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-select_streams", "v:0",
		inputPath,
	}

	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe: %w", err)
	}

	var result struct {
		Streams []struct {
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			CodecName string `json:"codec_name"`
			Duration  string `json:"duration"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("ffprobe parse: %w", err)
	}

	if len(result.Streams) == 0 {
		return nil, fmt.Errorf("ffprobe: no video stream found")
	}

	s := result.Streams[0]
	dur, _ := strconv.ParseFloat(strings.TrimSpace(s.Duration), 64)

	return &VideoInfo{
		Duration:  dur,
		Width:     s.Width,
		Height:    s.Height,
		CodecName: s.CodecName,
	}, nil
}
