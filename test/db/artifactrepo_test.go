package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/Assifar-Karim/apollo/internal/db"
)

func setupDB() (*sql.DB, string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}
	driver := "sqlite"
	dbName := fmt.Sprintf("%s/test.db", currentDir)
	database, err := db.New(driver, dbName, true)
	return database, dbName, err
}

func TestCreateArtifact(t *testing.T) {
	// Given
	name := "name"
	artifactType := "artifact-type"
	hash := "hash"
	size := int64(10)
	expectedResult := db.Artifact{
		Name: name,
		Type: artifactType,
		Size: size,
		Hash: hash,
	}
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	artifactRepo := db.NewSQLiteArtifactRepository(database)

	// When
	result, err := artifactRepo.CreateArtifact(name, artifactType, hash, size)
	if err != nil {
		t.Fatalf("Couldn't create artifact %v", err)
	}

	// Then
	if expectedResult.Name != result.Name ||
		expectedResult.Type != result.Type ||
		expectedResult.Size != result.Size ||
		expectedResult.Hash != result.Hash {
		t.Errorf("Expected %v but found %v", expectedResult, result)
	}
	row := database.QueryRow("SELECT name, type, size, hash FROM artifact WHERE name = name;")
	fetchedArtifact := db.Artifact{}
	err = row.Scan(&fetchedArtifact.Name, &fetchedArtifact.Type, &fetchedArtifact.Size, &fetchedArtifact.Hash)
	if errors.Is(err, sql.ErrNoRows) {
		t.Errorf("Expected %v to be saved on db but it wasn't!", expectedResult)
	}
	if expectedResult.Name != fetchedArtifact.Name ||
		expectedResult.Type != fetchedArtifact.Type ||
		expectedResult.Size != fetchedArtifact.Size ||
		expectedResult.Hash != fetchedArtifact.Hash {
		t.Errorf("Expected %v but found %v", expectedResult, fetchedArtifact)
	}
}

func TestFetchArtifactsWhenNoArtifacts(t *testing.T) {
	// Given
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	artifactRepo := db.NewSQLiteArtifactRepository(database)

	// When
	artifacts, err := artifactRepo.FetchArtifacts()
	if err != nil {
		t.Fatalf("Couldn't fetch from the db %v", err)
	}

	// Then
	if len(artifacts) != 0 {
		t.Errorf("Expected to find no artifact but found %v artifacts", len(artifacts))
	}
}

func TestFetchArtifactsWhenArtifactsExist(t *testing.T) {
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	artifactRepo := db.NewSQLiteArtifactRepository(database)
	artifact, err := artifactRepo.CreateArtifact("name", "artifact-type", "hash", 10)
	if err != nil {
		t.Fatalf("Couldn't create artifact for logic testing! %v", err)
	}

	// When
	artifacts, err := artifactRepo.FetchArtifacts()
	if err != nil {
		t.Fatalf("Couldn't fetch from the db %v", err)
	}

	// Then
	if len(artifacts) != 1 && (artifact.Name != artifacts[0].Name ||
		artifact.Type != artifacts[0].Type ||
		artifact.Hash != artifacts[0].Hash ||
		artifact.Size != artifacts[0].Size) {
		t.Errorf("Expected to find artifact %v but found %v", artifact, artifacts[0])
	}
}

func TestFetchArtifactByNameWhenArtifactDoesNotExist(t *testing.T) {
	// Given
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	artifactRepo := db.NewSQLiteArtifactRepository(database)

	// When
	artifact, err := artifactRepo.FetchArficatByName("name")

	// Then
	if artifact != nil && err != nil {
		t.Error("Expected to find no artifact by found one!")
	}
}

func TestFetchArtifactByNameWhenArtifactExists(t *testing.T) {
	// Given
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	artifactRepo := db.NewSQLiteArtifactRepository(database)
	artifact, err := artifactRepo.CreateArtifact("name", "artifact-type", "hash", 10)
	if err != nil {
		t.Fatalf("Couldn't create artifact for logic testing! %v", err)
	}

	// When
	result, err := artifactRepo.FetchArficatByName("name")
	if err != nil {
		t.Fatalf("Couldn't fetch from the db %v", err)
	}

	// Then
	if artifact.Name != result.Name ||
		artifact.Type != result.Type ||
		artifact.Hash != result.Hash ||
		artifact.Size != result.Size {
		t.Errorf("Expected to find artifact %v but found %v", artifact, result)
	}
}

func TestDeleteArtifactWhenArtifactDoesNotExist(t *testing.T) {
	// Given
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	artifactRepo := db.NewSQLiteArtifactRepository(database)

	// When
	isDeleted, err := artifactRepo.DeleteArtifact("name")
	if err != nil {
		t.Fatalf("The delete operation didn't work %v", err)
	}
	// Then
	if isDeleted {
		t.Error("Expected no record to be deleted!")
	}
}

func TestDeleteArtifactWhenArtifactExists(t *testing.T) {
	// Given
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	artifactRepo := db.NewSQLiteArtifactRepository(database)
	_, err = artifactRepo.CreateArtifact("name", "artifact-type", "hash", 10)
	if err != nil {
		t.Fatalf("Couldn't create artifact for logic testing! %v", err)
	}

	// When
	isDeleted, err := artifactRepo.DeleteArtifact("name")
	if err != nil {
		t.Fatalf("The delete operation didn't work %v", err)
	}

	// Then
	if !isDeleted {
		t.Error("Expected record to be deleted but no record was deleted!")
	}
}

func TestUpdateArtifactWhenArtifactDoesNotExist(t *testing.T) {
	// Given
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	artifactRepo := db.NewSQLiteArtifactRepository(database)

	// When
	artifact, err := artifactRepo.UpdateArtifact("name", "hash", 10)

	// Then
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("Expected to update no artifact but one was!: %v", artifact)
	}
}

func TestUpdateArtifactWhenArtifactExists(t *testing.T) {
	// Given
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	artifactRepo := db.NewSQLiteArtifactRepository(database)
	_, err = artifactRepo.CreateArtifact("name", "artifact-type", "hash", 10)
	if err != nil {
		t.Fatalf("Couldn't create artifact for logic testing! %v", err)
	}

	// When
	artifact, err := artifactRepo.UpdateArtifact("name", "new-hash", 15)
	if err != nil {
		t.Fatalf("The update operation didn't work %v", err)
	}

	// Then
	if artifact.Hash != "new-hash" || artifact.Size != 15 {
		t.Fatalf("Expected artifact data to be updated but it wasn't! %v", artifact)
	}
}
