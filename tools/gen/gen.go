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

const timestampLayout = "2006-01-02T15-04-05"

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
type FFProbeJSONResult struct {
	Format struct {
		Duration string `json:"duration"`
		Tags     struct {
			EndTime   string `json:"END_TIME"`
			StartTime string `json:"START_TIME"`
		}
	} `json:"format"`
}

type FFProbeResult struct {
	Duration float64
	Start    time.Time
	End      time.Time
}

// require ffprobe
func getFFProbeInfo(filePath string) (FFProbeResult, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return FFProbeResult{}, fmt.Errorf("error running ffprobe: %w", err)
	}

	var result FFProbeJSONResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return FFProbeResult{}, fmt.Errorf("error parsing JSON: %w", err)
	}

	ffpResult := FFProbeResult{}
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return FFProbeResult{}, err
	}
	{
		start, err := time.ParseInLocation(timestampLayout, result.Format.Tags.StartTime, jst)
		if err != nil {
			return FFProbeResult{}, err
		}
		ffpResult.Start = start
	}
	{
		end, err := time.ParseInLocation(timestampLayout, result.Format.Tags.EndTime, jst)
		if err != nil {
			return FFProbeResult{}, err
		}
		ffpResult.End = end
	}
	{
		duration, err := strconv.ParseFloat(result.Format.Duration, 64)
		if err != nil {
			return FFProbeResult{}, err
		}
		ffpResult.Duration = duration
	}

	return ffpResult, nil
}

type parseFileResult struct {
	name string
}

func parseFileName(raw string) (parseFileResult, error) {
	const fileRegexp = `^(?P<date>\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2})_(?P<name>.+?)\.webm$`
	re, err := regexp.Compile(fileRegexp)
	if err != nil {
		return parseFileResult{}, err
	}
	matches := re.FindAllStringSubmatch(filepath.Base(raw), -1)
	if len(matches) < 1 {
		return parseFileResult{}, fmt.Errorf("invalid format: %s", raw)
	}

	result := parseFileResult{
		name: matches[0][re.SubexpIndex("name")],
	}

	return result, nil
}

func formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
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
			result, err := getFFProbeInfo(path)
			if err != nil {
				return err
			}
			vinfo.Duration = result.Duration
			vinfo.RealStart = result.Start
			vinfo.RealEnd = result.End
		}
		{
			realDuration := vinfo.RealEnd.Sub(vinfo.RealStart)
			vinfo.RealDurationLabel = formatDuration(realDuration)
		}
		{
			result, err := parseFileName(filepath.Base(path))
			if err != nil {
				return err
			}
			vinfo.Name = result.name
		}
		info.Files = append(info.Files, vinfo)

		return nil
	})
	if err != nil {
		return Info{}, err
	}

	return info, nil
}
