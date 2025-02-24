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

type Info struct {
	Files []VideoInfo `json: "files"`
}

type VideoInfo struct {
	// タイトル
	Name string `json:"name"`
	// 動画の秒数
	Duration float64 `json:"duration"`
	// 実際の時間を表示する
	// 例: 01:45:23
	RealDurationLabel string    `json:"start"`
	RealStart         time.Time `json:"start"`
	// 計算で得る大体の値
	RealEnd time.Time `json:"end"`
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

type parseFileResult struct {
	start time.Time
	name  string
}

func parseFile(raw string) (parseFileResult, error) {
	const fileRegexp = `^(?P<date>\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2})_(?P<name>.+?)\.webm$`
	re, err := regexp.Compile(fileRegexp)
	if err != nil {
		return parseFileResult{}, err
	}
	matches := re.FindAllStringSubmatch(filepath.Base(raw), -1)
	if len(matches) < 1 {
		return parseFileResult{}, fmt.Errorf("invalid format: %s", raw)
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return parseFileResult{}, err
	}
	layout := "2006-01-02T15-04-05"
	parsedTime, err := time.ParseInLocation(layout, matches[0][re.SubexpIndex("date")], jst)
	if err != nil {
		return parseFileResult{}, err
	}

	result := parseFileResult{
		start: parsedTime,
		name:  matches[0][re.SubexpIndex("name")],
	}

	return result, nil
}

func formatDuration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60

	return fmt.Sprintf("%d時間%d分%d秒", h, m, s)
}

func ApplyDir(targetDir string) (Info, error) {
	info := Info{
		Files: []VideoInfo{},
	}

	err := filepath.Walk(targetDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		vinfo := VideoInfo{}
		{
			duration, err := getWebMDuration(path)
			if err != nil {
				return err
			}
			vinfo.Duration = duration
		}
		{
			result, err := parseFile("2025-02-23T21-59-41_ダークソウル3実況を見る.webm")
			if err != nil {
				return err
			}
			vinfo.Name = result.name
			vinfo.RealStart = result.start
			// 2秒に1回撮ってるので
			vinfo.RealEnd = vinfo.RealStart.Add(time.Second * time.Duration(vinfo.Duration*60*2))
			// 2秒に1回撮ってるので
			vinfo.RealDurationLabel = formatDuration(int(vinfo.Duration * 60 * 2))
		}

		info.Files = append(info.Files, vinfo)

		return nil
	})
	if err != nil {
		return Info{}, err
	}

	return info, nil
}
