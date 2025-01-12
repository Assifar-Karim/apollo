package coordinator

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/Assifar-Karim/apollo/internal/db"
	"github.com/Assifar-Karim/apollo/internal/utils"
)

type ArtifactManager interface {
	CreateArtifact(filename, artifactType string, size int64, file multipart.File) (db.Artifact, error)
	GetAllArtifactDetails() ([]db.Artifact, error)
	GetArtifactDetailsByName(filename string) (*db.Artifact, error)
	DeleteArtifact(filename string) (bool, error)
}

type ArtifactMngmtSvc struct {
	artifactRepository db.ArtifactRepository
	config             *Config
	logger             *utils.Logger
}

func getFileContent(file multipart.File) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func hash(buf []byte) (string, error) {
	h := sha256.New()
	if _, err := h.Write(buf); err != nil {
		return "", nil
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func writeFile(path string, fileContent []byte) error {
	if err := os.WriteFile(path, fileContent, 0666); err != nil {
		return err
	}
	return nil
}

func (s ArtifactMngmtSvc) CreateArtifact(filename, artifactType string, size int64, file multipart.File) (db.Artifact, error) {
	path := fmt.Sprintf("%s/%s", s.config.GetArtifactsPath(), filename)
	fileContent, err := getFileContent(file)
	if err != nil {
		s.logger.Error(err.Error())
		return db.Artifact{}, err
	}
	fileHash, err := hash(fileContent)
	if err != nil {
		s.logger.Error(err.Error())
		return db.Artifact{}, err
	}
	artifact, err := s.artifactRepository.FetchArficatByName(filename)
	if err != nil {
		return db.Artifact{}, err
	}
	if artifact == nil {
		if err = writeFile(path, fileContent); err != nil {
			s.logger.Error(err.Error())
			return db.Artifact{}, err
		}
		return s.artifactRepository.CreateArtifact(filename, artifactType, fileHash, size)
	}

	if fileHash == artifact.Hash {
		return *artifact, nil
	}

	if err = writeFile(path, fileContent); err != nil {
		s.logger.Error(err.Error())
		return db.Artifact{}, err
	}

	return s.artifactRepository.UpdateArtifact(filename, fileHash, size)
}

func (s ArtifactMngmtSvc) GetAllArtifactDetails() ([]db.Artifact, error) {
	return s.artifactRepository.FetchArtifacts()
}

func (s ArtifactMngmtSvc) GetArtifactDetailsByName(filename string) (*db.Artifact, error) {
	return s.artifactRepository.FetchArficatByName(filename)
}

func (s ArtifactMngmtSvc) DeleteArtifact(filename string) (bool, error) {
	path := fmt.Sprintf("%s/%s", s.config.artifactsPath, filename)
	if _, err := os.Stat(path); err != nil {
		s.logger.Error(err.Error())
		return false, err
	}
	if err := os.Remove(path); err != nil {
		s.logger.Error(err.Error())
		return false, err
	}
	return s.artifactRepository.DeleteArtifact(filename)
}

func NewArtifactManager(artifactRepository db.ArtifactRepository) ArtifactManager {
	return &ArtifactMngmtSvc{
		artifactRepository: artifactRepository,
		config:             GetConfig(),
		logger:             utils.GetLogger(),
	}
}
