package db

import (
	"database/sql"
	"errors"

	"github.com/Assifar-Karim/apollo/internal/utils"
)

type ArtifactRepository interface {
	CreateArtifact(name, artifactType, hash string, size int64) (Artifact, error)
	FetchArtifacts() ([]Artifact, error)
	FetchArficatByName(name string) (*Artifact, error)
	DeleteArtifact(name string) (bool, error)
	UpdateArtifact(name, hash string, size int64) (Artifact, error)
}

type SQLiteArtifactRepository struct {
	db     *sql.DB
	logger *utils.Logger
}

func (r SQLiteArtifactRepository) CreateArtifact(name, artifactType, hash string, size int64) (Artifact, error) {
	query := "INSERT INTO artifact VALUES (?, ?, ?, ?);"
	r.logger.Trace(query)
	_, err := r.db.Exec(query, name, artifactType, size, hash)
	if err != nil {
		r.logger.Error(err.Error())
		return Artifact{}, err
	}
	return Artifact{
		Name: name,
		Type: artifactType,
		Size: size,
		Hash: hash,
	}, nil
}

func (r SQLiteArtifactRepository) FetchArtifacts() ([]Artifact, error) {
	query := "SELECT name, type, size, hash FROM artifact;"
	r.logger.Trace(query)
	rows, err := r.db.Query(query)
	if err != nil {
		r.logger.Error(err.Error())
		return []Artifact{}, err
	}
	defer rows.Close()
	artifacts := []Artifact{}
	for rows.Next() {
		artifact := Artifact{}
		err := rows.Scan(&artifact.Name, &artifact.Type, &artifact.Size, &artifact.Hash)
		if err != nil {
			r.logger.Error(err.Error())
			return []Artifact{}, err
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, nil
}

func (r SQLiteArtifactRepository) FetchArficatByName(name string) (*Artifact, error) {
	query := "SELECT name, type, size, hash FROM artifact WHERE name = ?;"
	r.logger.Trace(query)
	row := r.db.QueryRow(query, name)
	artifact := Artifact{}
	err := row.Scan(&artifact.Name, &artifact.Type, &artifact.Size, &artifact.Hash)

	if errors.Is(err, sql.ErrNoRows) {
		r.logger.Warn("No artifact with name %s was found", name)
		return nil, nil
	}
	if err != nil {
		r.logger.Error(err.Error())
		return nil, err
	}

	return &artifact, nil
}

func (r SQLiteArtifactRepository) DeleteArtifact(name string) (bool, error) {
	query := "DELETE FROM artifact WHERE name = ?;"
	r.logger.Trace(query)
	res, err := r.db.Exec(query, name)
	if err != nil {
		r.logger.Error(err.Error())
		return false, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		r.logger.Error(err.Error())
		return false, err
	}
	return count != 0, nil
}

func (r SQLiteArtifactRepository) UpdateArtifact(name, hash string, size int64) (Artifact, error) {
	query := "UPDATE artifact SET hash = ?, size = ? WHERE name = ?;"
	r.logger.Trace(query)
	_, err := r.db.Exec(query, hash, size, name)
	if err != nil {
		return Artifact{}, err
	}
	artifact, err := r.FetchArficatByName(name)
	if err != nil {
		r.logger.Error(err.Error())
		return Artifact{}, err
	}
	if artifact == nil {
		return Artifact{}, sql.ErrNoRows
	}
	return *artifact, nil
}

func NewSQLiteArtifactRepository(db *sql.DB) ArtifactRepository {
	return &SQLiteArtifactRepository{
		db:     db,
		logger: utils.GetLogger(),
	}
}
