package coordinator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Assifar-Karim/apollo/internal/coordinator"
	"github.com/Assifar-Karim/apollo/internal/db"
)

type artifactRepositoryMock struct {
	calls int
}

func (r *artifactRepositoryMock) CreateArtifact(name, artifactType, hash string, size int64) (db.Artifact, error) {
	r.calls += 1
	if name == "new-case" {
		return db.Artifact{
			Name: name,
			Type: artifactType,
			Size: size,
			Hash: hash,
		}, nil
	}
	return db.Artifact{}, nil
}

func (r *artifactRepositoryMock) FetchArtifacts() ([]db.Artifact, error) {
	// do nothing
	return nil, nil
}

func (r *artifactRepositoryMock) FetchArficatByName(name string) (*db.Artifact, error) {
	r.calls += 1
	if name == "fail-case" {
		return nil, errors.New("custom fetch error")
	}
	if name == "new-case" {
		return nil, nil
	}
	if name == "exist-case" {
		return &db.Artifact{
			Hash: "1c87d5ffba8bd8a4143f34f99beb33dfeb18031a545dc43647f21f4c4b9e99a3",
		}, nil
	}

	if name == "update-case" {
		return &db.Artifact{
			Hash: "old-hash",
		}, nil
	}
	return nil, nil
}

func (r *artifactRepositoryMock) DeleteArtifact(name string) (bool, error) {
	// do nothing
	r.calls += 1
	return true, nil
}

func (r *artifactRepositoryMock) UpdateArtifact(name, hash string, size int64) (db.Artifact, error) {
	r.calls += 1
	return db.Artifact{}, nil
}

func TestCreateArtifactWhenArtifactDataFetchFails(t *testing.T) {
	// Given
	filename := "fail-case"
	reader := strings.NewReader("dummy artifact data")
	t.Setenv("ARTIFACTS_PATH", os.TempDir())
	mockRepository := &artifactRepositoryMock{
		calls: 0,
	}
	artifactManager := coordinator.NewArtifactManager(mockRepository)

	// When
	_, err := artifactManager.CreateArtifact(filename, "type", reader.Size(), reader)

	// Then
	if mockRepository.calls != 1 && err == nil {
		t.Errorf("CreateArtifact was expected to fail but it didn't!")
	}
}

func TestCreateArtifactWhenArtifactDoesNotExist(t *testing.T) {
	// Given
	filename := "new-case"
	reader := strings.NewReader("dummy artifact data")
	t.Setenv("ARTIFACTS_PATH", os.TempDir())
	mockRepository := &artifactRepositoryMock{
		calls: 0,
	}
	artifactManager := coordinator.NewArtifactManager(mockRepository)

	// When
	_, err := artifactManager.CreateArtifact(filename, "type", reader.Size(), reader)
	defer os.Remove(fmt.Sprintf("%s/%s", os.TempDir(), filename))

	// Then
	if mockRepository.calls != 2 && err != nil {
		t.Errorf("CreateArtifact failed with unexpected error!")
	}
}

func TestCreateArtifactWhenArtifactExistsWithSameHash(t *testing.T) {
	// Given
	filename := "exist-case"
	reader := strings.NewReader("dummy artifact data")
	t.Setenv("ARTIFACTS_PATH", os.TempDir())
	mockRepository := &artifactRepositoryMock{
		calls: 0,
	}
	artifactManager := coordinator.NewArtifactManager(mockRepository)

	// When
	artifact, err := artifactManager.CreateArtifact(filename, "type", reader.Size(), reader)

	// Then
	if artifact.Hash != "1c87d5ffba8bd8a4143f34f99beb33dfeb18031a545dc43647f21f4c4b9e99a3" && mockRepository.calls != 1 && err != nil {
		t.Errorf("Expected existing artifact but method failed!")
	}
}

func TestCreateArtifactWhenArtifactExistsWithDifferentHash(t *testing.T) {
	// Given
	filename := "update-case"
	reader := strings.NewReader("dummy artifact data")
	t.Setenv("ARTIFACTS_PATH", os.TempDir())
	mockRepository := &artifactRepositoryMock{
		calls: 0,
	}
	artifactManager := coordinator.NewArtifactManager(mockRepository)

	// When
	_, err := artifactManager.CreateArtifact(filename, "type", reader.Size(), reader)
	defer os.Remove(fmt.Sprintf("%s/%s", os.TempDir(), filename))

	// Then
	if err != nil && mockRepository.calls != 2 {
		t.Errorf("Expected artifact to be updated but method failed to do!")
	}
}

func TestDeleteArtifactWhenFileDoesNotExist(t *testing.T) {
	// Given
	filename := "temp_artifact.txt"
	file, err := os.CreateTemp("", filename)
	if err != nil {
		t.Fatalf("Couldn't create temp file %v during test initialization!", filename)
	}
	if err := os.Remove(file.Name()); err != nil {
		t.Fatalf("Couldn't delete randomly created temp for test!")
	}
	t.Setenv("ARTIFACTS_PATH", os.TempDir())
	mockRepository := &artifactRepositoryMock{
		calls: 0,
	}
	artifactManager := coordinator.NewArtifactManager(mockRepository)

	// When
	_, err = artifactManager.DeleteArtifact(file.Name())

	// Then
	if err == nil && mockRepository.calls != 0 {
		t.Errorf("File was deleted even though it wasn't supposed to exist!")
	}
}

func TestDeleteArtifactWhenFileExists(t *testing.T) {
	// Given
	filename := "temp_artifact.txt"
	file, err := os.CreateTemp("", filename)
	if err != nil {
		t.Fatalf("Couldn't create temp file %v during test initialization!", filename)
	}
	defer os.Remove(file.Name())
	t.Setenv("ARTIFACTS_PATH", os.TempDir())
	mockRepository := &artifactRepositoryMock{
		calls: 0,
	}
	artifactManager := coordinator.NewArtifactManager(mockRepository)

	// When
	artifactManager.DeleteArtifact(filepath.Base(file.Name()))

	// Then
	if _, err := os.Stat(file.Name()); !errors.Is(err, os.ErrNotExist) || mockRepository.calls != 1 {
		t.Errorf("File %s wasn't deleted by manager!", file.Name())
	}
}
