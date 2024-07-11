package utils

import (
	"testing"
)

func TestFileSizeByte(t *testing.T) {
	expected := "1 B"
	actual := FileSize(1)
	if actual != expected {
		t.Fatalf("expected %v but got %v", expected, actual)
	}
}

func TestFileSizeKiloByte(t *testing.T) {
	expected := "1.0 KB"
	actual := FileSize(1024)
	if actual != expected {
		t.Fatalf("expected %v but got %v", expected, actual)
	}
}

func TestFileSizeMegaByte(t *testing.T) {
	expected := "1.0 MB"
	actual := FileSize(1024 * 1024)
	if actual != expected {
		t.Fatalf("expected %v but got %v", expected, actual)
	}
}

func TestFileSizeGigaByte(t *testing.T) {
	expected := "1.0 GB"
	actual := FileSize(1024 * 1024 * 1024)
	if actual != expected {
		t.Fatalf("expected %v but got %v", expected, actual)
	}
}

func TestFileSizeTeraByte(t *testing.T) {
	expected := "1.0 TB"
	actual := FileSize(1024 * 1024 * 1024 * 1024)
	if actual != expected {
		t.Fatalf("expected %v but got %v", expected, actual)
	}
}

func TestFileSizePetaByte(t *testing.T) {
	expected := "1.0 PB"
	actual := FileSize(1024 * 1024 * 1024 * 1024 * 1024)
	if actual != expected {
		t.Fatalf("expected %v but got %v", expected, actual)
	}
}

func TestFileSizeExaByte(t *testing.T) {
	expected := "1.0 EB"
	actual := FileSize(1024 * 1024 * 1024 * 1024 * 1024 * 1024)
	if actual != expected {
		t.Fatalf("expected %v but got %v", expected, actual)
	}
}

func TestFileSizeZero(t *testing.T) {
	expected := "0 B"
	actual := FileSize(0)
	if actual != expected {
		t.Fatalf("expected %v but got %v", expected, actual)
	}
}