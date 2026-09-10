package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestChartPNGValidation(t *testing.T) {
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatal(err)
	}
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())
	got, err := chartPNG(dataURL)
	if err != nil || !bytes.Equal(got, encoded.Bytes()) {
		t.Fatalf("valid PNG rejected: %v", err)
	}
	for _, input := range []string{"data:text/html;base64,YQ==", "data:image/png;base64,??", "data:image/png;base64,YQ==", dataURL[:len(dataURL)-8]} {
		if _, err := chartPNG(input); err == nil {
			t.Fatal("invalid PNG accepted")
		}
	}
}

func TestSaveExportPreservesDestinationOnFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "results.csv")
	if err := os.WriteFile(path, []byte("existing results"), 0600); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("export interrupted")
	_, err := saveExport(path, func(w io.Writer) error {
		_, _ = io.WriteString(w, "partial")
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("lost export error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "existing results" {
		t.Fatalf("destination changed: %q, %v", data, err)
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 1 {
		t.Fatalf("temporary export leaked: %v", err)
	}
	if _, err := saveExport(path, func(w io.Writer) error { _, err := io.WriteString(w, "complete"); return err }); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	if string(data) != "complete" {
		t.Fatal("successful export was not saved")
	}
}
