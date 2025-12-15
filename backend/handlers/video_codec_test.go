package handlers

import "testing"

func TestShouldForceMp4CodecTransmux(t *testing.T) {
	testCases := []struct {
		name     string
		codec    string
		expected bool
	}{
		{name: "h264 stays direct", codec: "h264", expected: false},
		{name: "avc prefix stays direct", codec: "avc1.640028", expected: false},
		{name: "mpeg4 stays direct", codec: "MPEG4", expected: false},
		{name: "hevc requires transmux", codec: "hevc", expected: true},
		{name: "h265 requires transmux", codec: "H265", expected: true},
		{name: "vp9 requires transmux", codec: "vp9", expected: true},
		{name: "empty defaults to transmux", codec: "", expected: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldForceMp4CodecTransmux(tc.codec)
			if got != tc.expected {
				t.Fatalf("shouldForceMp4CodecTransmux(%q) = %t, want %t", tc.codec, got, tc.expected)
			}
		})
	}
}
