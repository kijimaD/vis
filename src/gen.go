package vis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const timestampLayout = "2006-01-02T15-04-05"

type Info struct {
	Files []VideoInfo `json:"files"`
}

type VideoInfo struct {
	// パス
	Path string `json:"path"`
	// タイトル
	Name string `json:"name"`
	// 動画の秒数
	Duration float64 `json:"duration"`
	// 実際の経過秒数
	RealDurationLabel string    `json:"real_duration_label"`
	RealStart         time.Time `json:"start"`
	RealStartLabel    string    `json:"real_start_label"`
	RealEnd           time.Time `json:"end"`
	RealEndLabel      string    `json:"real_end_label"`
	// 人間が読みやすい単位に調整されたファイルサイズ
	Size string `json:"size"`
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
	path string
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

	pname := matches[0][re.SubexpIndex("name")]
	pdate := matches[0][re.SubexpIndex("date")]
	result := parseFileResult{
		path: fmt.Sprintf("./files/%s_%s.webm", pdate, pname),
		name: pname,
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

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.0f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
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
		if !strings.Contains(fi.Name(), ".webm") {
			return nil
		}

		vinfo := VideoInfo{}
		{
			vinfo.Size = formatSize(fi.Size())
		}
		{
			result, err := getFFProbeInfo(path)
			if err != nil {
				return err
			}
			vinfo.Duration = result.Duration
			vinfo.RealStart = result.Start
			vinfo.RealEnd = result.End
			vinfo.RealStartLabel = vinfo.RealStart.Format(timestampLayout)
			vinfo.RealEndLabel = vinfo.RealEnd.Format(timestampLayout)
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
			vinfo.Path = result.path
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
