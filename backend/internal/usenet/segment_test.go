package usenet

import (
	"bytes"
	"io"
	"testing"

	"github.com/acomagu/bufpipe"
	"github.com/mnightingale/rapidyenc"
)

func encodeTestSegment(t *testing.T, payload []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	enc, err := rapidyenc.NewEncoder(&buf, rapidyenc.Meta{
		FileName:   "sample.bin",
		FileSize:   int64(len(payload)),
		PartSize:   int64(len(payload)),
		PartNumber: 1,
		TotalParts: 1,
	})
	if err != nil {
		t.Fatalf("failed to create encoder: %v", err)
	}
	if _, err := enc.Write(payload); err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("failed to close encoder: %v", err)
	}

	return buf.Bytes()
}

func TestSegmentGetReaderDecodesYEnc(t *testing.T) {
	payload := []byte{0x00, 0xFF, 0x10, 0x20, 0x7F, 0x80, 0xAA, 0xBB}
	encoded := encodeTestSegment(t, payload)

	reader, writer := bufpipe.New(nil)
	seg := &segment{
		Id:          "<sample@usenet>",
		Start:       0,
		End:         int64(len(payload) - 1),
		SegmentSize: int64(len(payload)),
		reader:      reader,
		writer:      writer,
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := writer.Write(encoded); err != nil {
			t.Errorf("failed to write encoded data: %v", err)
		}
		_ = writer.Close()
	}()

	got, err := io.ReadAll(seg.GetReader())
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if !bytes.Equal(payload, got) {
		t.Fatalf("decoded payload mismatch\nwant=%v\n got=%v", payload, got)
	}

	if err := seg.Close(); err != nil {
		t.Fatalf("segment.Close() error = %v", err)
	}

	<-done
}

func TestSegmentGetReaderHandlesDecodedInput(t *testing.T) {
	payload := []byte("Plain decoded bytes")

	reader, writer := bufpipe.New(nil)
	seg := &segment{
		Id:          "<decoded@usenet>",
		Start:       0,
		End:         int64(len(payload) - 1),
		SegmentSize: int64(len(payload)),
		reader:      reader,
		writer:      writer,
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := writer.Write(payload); err != nil {
			t.Errorf("failed to write decoded data: %v", err)
		}
		_ = writer.Close()
	}()

	got, err := io.ReadAll(seg.GetReader())
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if !bytes.Equal(payload, got) {
		t.Fatalf("decoded reader mismatch\nwant=%q\n got=%q", payload, got)
	}

	if err := seg.Close(); err != nil {
		t.Fatalf("segment.Close() error = %v", err)
	}

	<-done
}

func TestSegmentGetReaderRespectsOffsets(t *testing.T) {
	payload := []byte("Hello, NovaStream!")
	encoded := encodeTestSegment(t, payload)

	reader, writer := bufpipe.New(nil)
	start := int64(2)
	end := int64(len(payload) - 3)
	seg := &segment{
		Id:          "<slice@usenet>",
		Start:       start,
		End:         end,
		SegmentSize: int64(len(payload)),
		reader:      reader,
		writer:      writer,
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := writer.Write(encoded); err != nil {
			t.Errorf("failed to write encoded data: %v", err)
		}
		_ = writer.Close()
	}()

	got, err := io.ReadAll(seg.GetReader())
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	want := payload[start : end+1]
	if !bytes.Equal(want, got) {
		t.Fatalf("decoded slice mismatch\nwant=%q\n got=%q", want, got)
	}

	if err := seg.Close(); err != nil {
		t.Fatalf("segment.Close() error = %v", err)
	}

	<-done
}
