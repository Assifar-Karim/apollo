package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Assifar-Karim/apollo/internal/utils"
)

type Job struct {
	Id             string         `json:"id"`
	NReducers      int            `json:"nReducers"`
	OutputLocation OutputLocation `json:"outputLocation"`
	InputData      InputData      `json:"inputData"`
	StartTime      int64          `json:"startTime"`
	EndTime        *int64         `json:"endTime,omitempty"`
}

type Task struct {
	Id        string     `json:"id"`
	Job       *Job       `json:"job,omitempty"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	Program   Artifact   `json:"program"`
	InputData *InputData `json:"inputData,omitempty"`
	PodName   *string    `json:"podName,omitempty"`
	StartTime int64      `json:"startTime"`
	EndTime   *int64     `json:"endTime,omitempty"`
}

type InputData struct {
	Id         int    `json:"id"`
	Path       string `json:"path"`
	Type       string `json:"type"`
	SplitStart *int64 `json:"splitStart,omitempty"`
	SplitEnd   *int64 `json:"splitEnd,omitempty"`
}

type OutputLocation struct {
	Location string `json:"location"`
	UseSSL   bool   `json:"useSSL"`
}

type Artifact struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int64  `json:"size"`
	Hash string `json:"hash"`
}

func runInTx(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	err = fn(tx)
	if err == nil {
		return tx.Commit()
	}
	rollbackErr := tx.Rollback()
	if rollbackErr != nil {
		// In case even the rollback fails
		return errors.Join(err, rollbackErr)
	}
	return err
}

func New(driver, dbName string) (*sql.DB, error) {
	logger := utils.GetLogger()
	// Open DB connection
	logger.Info("Connecting to %s:%s database", driver, dbName)
	db, err := sql.Open(driver, dbName)
	if err != nil {
		return nil, err
	}
	// Setup DB tables
	queries := make([]string, 5)

	queries[0] = `CREATE TABLE IF NOT EXISTS output_location (
    location VARCHAR PRIMARY KEY NOT NULL,
    use_SSL BOOLEAN NOT NULL);`

	queries[1] = `CREATE TABLE IF NOT EXISTS input_data (
    id INTEGER PRIMARY KEY NOT NULL,
    path VARCHAR NOT NULL,
    type VARCHAR NOT NULL,
    split_start INTEGER,
    split_end INTEGER);`

	queries[2] = `CREATE TABLE IF NOT EXISTS job (
    id VARCHAR PRIMARY KEY NOT NULL,
    n_reducers INTEGER NOT NULL,
    output_path VARCHAR NOT NULL,
    input_id INTEGER NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
	FOREIGN KEY(input_id) REFERENCES input_data(id),
	FOREIGN KEY(output_path) REFERENCES output_location(location));`

	queries[3] = `CREATE TABLE IF NOT EXISTS artifact (
    name VARCHAR PRIMARY KEY NOT NULL,
    type VARCHAR NOT NULL DEFAULT executable,
    size INTEGER NOT NULL DEFAULT 0,
	hash VARCHAR NOT NULL);`

	queries[4] = `CREATE TABLE IF NOT EXISTS task (
    id VARCHAR PRIMARY KEY NOT NULL,
    job_id VARCHAR NOT NULL,
    type VARCHAR NOT NULL,
    status VARCHAR NOT NULL DEFAULT scheduled,
    program_name VARCHAR NOT NULL,
    input_data_id INTEGER,
    pod_name VARCHAR,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    FOREIGN KEY(job_id) REFERENCES job(id),
    FOREIGN KEY(input_data_id) REFERENCES input_data(id),
	FOREIGN KEY(program_name) REFERENCES artifact(name));`

	for _, query := range queries {
		logger.Trace(query)
		_, err := db.Exec(query)
		if err != nil {
			return nil, err
		}
	}
	return db, err
}

func (t Task) GetType() (int64, error) {
	taskType := strings.ToLower(t.Type)
	if taskType == "mapper" {
		return 0, nil
	} else if taskType == "reducer" {
		return 1, nil
	}
	return -1, fmt.Errorf("%s isn't supported by apollo", taskType)
}

func (t Task) GetProgramContent(origin string) ([]byte, error) {
	path := fmt.Sprintf("%s/%s", origin, t.Program.Name)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, fInfo.Size())
	_, err = file.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}
