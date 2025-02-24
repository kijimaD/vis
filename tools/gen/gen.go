package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type VideoInfo struct {
	Name     string    `json:"name"`
	Duration float64   `json:"duration"`
	Start    time.Time `json:"start"`
	// 計算で得る大体の値
	End time.Time `json:"end"`
}

// FFProbeResult represents the JSON structure of ffprobe output
type FFProbeResult struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

// require ffprobe
func getWebMDuration(filePath string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("error running ffprobe: %w", err)
	}

	var result FFProbeResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return 0, fmt.Errorf("error parsing JSON: %w", err)
	}

	duration, err := strconv.ParseFloat(result.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting duration to float: %w", err)
	}

	return duration, nil
}

type parseResult struct {
	start time.Time
	name  string
}

func parseFileName(raw string) (parseResult, error) {
	const fileRegexp = `^(?P<date>\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2})_(?P<name>.+?)\.webm$`
	re, err := regexp.Compile(fileRegexp)
	if err != nil {
		return parseResult{}, err
	}
	matches := re.FindAllStringSubmatch(filepath.Base(raw), -1)
	if len(matches) < 1 {
		return parseResult{}, fmt.Errorf("invalid format: %s", raw)
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return parseResult{}, err
	}
	layout := "2006-01-02T15-04-05"
	parsedTime, err := time.ParseInLocation(layout, matches[0][re.SubexpIndex("date")], jst)
	if err != nil {
		return parseResult{}, err
	}

	result := parseResult{
		start: parsedTime,
		name:  matches[0][re.SubexpIndex("name")],
	}
	return result, nil
}

func ApplyDir(targetDir string) error {
	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		fmt.Println(path)
		duration, err := getWebMDuration(path)
		if err != nil {
			return err
		}
		fmt.Printf("WebM Duration: %.2f seconds\n", duration)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
