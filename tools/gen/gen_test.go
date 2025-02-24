package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetWebMDuration(t *testing.T) {
	duration, err := getWebMDuration("../testdata/example.webm")
	assert.NoError(t, err)
	assert.Equal(t, 42.4, duration)
}

func TestApplyDir(t *testing.T) {
	info, err := ApplyDir("../testdata")
	assert.NoError(t, err)

	jst, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)
	expect := Info{
		Files: []VideoInfo{
			VideoInfo{
				Name:              "ダークソウル3実況を見る",
				Duration:          42.4,
				RealDurationLabel: "1時間24分48秒",
				RealStart:         time.Date(2025, time.February, 23, 21, 59, 41, 0, jst),
				RealEnd:           time.Date(2025, time.February, 23, 23, 24, 29, 0, jst),
			},
		},
	}
	assert.Equal(t, expect, info)
}

func TestParseFileName(t *testing.T) {
	result, err := parseFile("2025-02-23T21-59-41_ダークソウル3実況を見る.webm")
	assert.NoError(t, err)
	assert.Equal(t, "ダークソウル3実況を見る", result.name)

	jst, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2025, time.February, 23, 21, 59, 41, 0, jst), result.start)
}

func TestDuration(t *testing.T) {
	assert.Equal(t, "0時間0分1秒", formatDuration(1))
	assert.Equal(t, "0時間1分1秒", formatDuration(61))
	assert.Equal(t, "1時間0分0秒", formatDuration(3600))
}
