package vis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetWebMDuration(t *testing.T) {
	result, err := getFFProbeInfo("./testdata/2025-02-24T13-17-51_サンプル.webm")
	assert.NoError(t, err)

	jst, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)

	expect := FFProbeResult{
		Duration: 0.5,
		Start:    time.Date(2025, time.February, 24, 13, 17, 51, 0, jst),
		End:      time.Date(2025, time.February, 24, 13, 18, 15, 0, jst),
	}
	assert.Equal(t, expect, result)
}

func TestApplyDir(t *testing.T) {
	info, err := ApplyDir("./testdata")
	assert.NoError(t, err)

	jst, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)
	expect := Info{
		Files: []VideoInfo{
			VideoInfo{
				Path:              "./files/2025-02-24T13-17-51_サンプル.webm",
				Name:              "サンプル",
				Duration:          0.5,
				RealDurationLabel: "0時間0分24秒",
				RealStart:         time.Date(2025, time.February, 24, 13, 17, 51, 0, jst),
				RealEnd:           time.Date(2025, time.February, 24, 13, 18, 15, 0, jst),
				RealStartLabel:    "2025-02-24T13-17-51",
				RealEndLabel:      "2025-02-24T13-18-15",
				Size:              "56KB",
			},
		},
	}
	assert.Equal(t, expect, info)
}

func TestParseFileName(t *testing.T) {
	result, err := parseFileName("2025-02-23T21-59-41_ダークソウル3実況を見る.webm")
	assert.NoError(t, err)
	assert.Equal(t, "ダークソウル3実況を見る", result.name)
}

func TestDuration(t *testing.T) {
	assert.Equal(t, "0時間0分1秒", formatDuration(time.Second*1))
	assert.Equal(t, "0時間1分1秒", formatDuration(time.Minute*1+time.Second*1))
	assert.Equal(t, "1時間0分0秒", formatDuration(time.Hour*1))
}

func TestFormatSize(t *testing.T) {
	assert.Equal(t, "1B", formatSize(1))
	assert.Equal(t, "1KB", formatSize(1024))
	assert.Equal(t, "1KB", formatSize(1025)) // 切り捨て
	assert.Equal(t, "2KB", formatSize(2048))
	assert.Equal(t, "64KB", formatSize(1024*64))
	assert.Equal(t, "1MB", formatSize(1024*1024))
	assert.Equal(t, "64MB", formatSize(1024*1024*64))
}
