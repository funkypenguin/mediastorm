package handlers

import (
	"bytes"
	"io"
	"testing"
)

func TestAlignMatroskaClusterFindsPattern(t *testing.T) {
	garbage := bytes.Repeat([]byte{0xAA}, 1024)
	cluster := append([]byte{0x1F, 0x43, 0xB6, 0x75}, []byte("cluster-data")...)
	payload := append(garbage, cluster...)

	alignedReader, dropped, err := alignMatroskaCluster(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		t.Fatalf("expected successful alignment, got error: %v", err)
	}
	if dropped != int64(len(garbage)) {
		t.Fatalf("expected dropped bytes=%d, got %d", len(garbage), dropped)
	}

	header := make([]byte, len(cluster))
	if _, err := io.ReadFull(alignedReader, header); err != nil {
		t.Fatalf("failed reading aligned data: %v", err)
	}
	if !bytes.Equal(header[:4], []byte{0x1F, 0x43, 0xB6, 0x75}) {
		t.Fatalf("expected cluster to start with ID, got %x", header[:4])
	}
}

func TestAlignMatroskaClusterMissingPatternReturnsBuffer(t *testing.T) {
	data := bytes.Repeat([]byte{0xBB}, 2048)

	alignedReader, dropped, err := alignMatroskaCluster(bytes.NewReader(data), 512)
	if err == nil {
		t.Fatalf("expected error when cluster pattern missing")
	}
	if dropped != 0 {
		t.Fatalf("expected dropped bytes 0 when pattern missing, got %d", dropped)
	}

	result, readErr := io.ReadAll(alignedReader)
	if readErr != nil {
		t.Fatalf("failed to read data from aligned reader: %v", readErr)
	}
	if !bytes.Equal(result, data) {
		t.Fatalf("expected reader to return original data when pattern missing")
	}
}
