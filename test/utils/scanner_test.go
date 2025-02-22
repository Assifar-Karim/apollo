package utils

import (
	"bufio"
	"os"
	"testing"

	"github.com/Assifar-Karim/apollo/internal/utils"
)

func readFile(path string) (*bufio.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return bufio.NewReader(file), nil
}

func TestModifiedScannerWithCRLFCorpusAndFullLines(t *testing.T) {
	// Given
	path := "data/crlf_corpus_1.txt"
	reader, err := readFile(path)
	if err != nil {
		t.Errorf("Couldn't read file %s -> %v", path, err)
	}
	scanner := utils.NewScanner(reader)
	expectedResult := "Line 1\nLine 2\n"

	// When
	res := ""
	for scanner.Scan() {
		res += scanner.Text()
	}

	// Then
	if res != expectedResult {
		t.Errorf("Expected %s but found %s", expectedResult, res)
	}
}

func TestModifiedScannerWithCRLFCorpusAndIncompleteLines(t *testing.T) {
	// Given
	path := "data/crlf_corpus_2.txt"
	reader, err := readFile(path)
	if err != nil {
		t.Errorf("Couldn't read file %s -> %v", path, err)
	}
	scanner := utils.NewScanner(reader)
	expectedResult := "Line 1\nLin"

	// When
	res := ""
	for scanner.Scan() {
		res += scanner.Text()
	}

	// Then
	if res != expectedResult {
		t.Errorf("Expected %s but found %s", expectedResult, res)
	}
}

func TestModifiedScannerWithLFCorpusAndFullLines(t *testing.T) {
	// Given
	path := "data/lf_corpus_1.txt"
	reader, err := readFile(path)
	if err != nil {
		t.Errorf("Couldn't read file %s -> %v", path, err)
	}
	scanner := utils.NewScanner(reader)
	expectedResult := "Line 1\nLine 2\n"

	// When
	res := ""
	for scanner.Scan() {
		res += scanner.Text()
	}

	// Then
	if res != expectedResult {
		t.Errorf("Expected %s but found %s", expectedResult, res)
	}
}

func TestModifiedScannerWithLFCorpusAndIncompleteLines(t *testing.T) {
	// Given
	path := "data/lf_corpus_2.txt"
	reader, err := readFile(path)
	if err != nil {
		t.Errorf("Couldn't read file %s -> %v", path, err)
	}
	scanner := utils.NewScanner(reader)
	expectedResult := "Line 1\nLin"

	// When
	res := ""
	for scanner.Scan() {
		res += scanner.Text()
	}

	// Then
	if res != expectedResult {
		t.Errorf("Expected %s but found %s", expectedResult, res)
	}
}
