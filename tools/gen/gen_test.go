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
	err := ApplyDir("../testdata")
	assert.NoError(t, err)
}

func TestParseFileName(t *testing.T) {
	result, err := parseFileName("2025-02-23T21-59-41_ダークソウル3実況を見る.webm")
	assert.NoError(t, err)
	assert.Equal(t, "ダークソウル3実況を見る", result.name)

	jst, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2025, time.February, 23, 21, 59, 41, 0, jst), result.start)
}
