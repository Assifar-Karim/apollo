package db

import (
	"database/sql"
	"errors"

	"github.com/Assifar-Karim/apollo/internal/utils"
)

type JobRepository interface {
	CreateJob(nReducers int, startTime int64, id, inputPath, inputType, outputPath string, useSSL bool) (Job, error)
	FetchJobs() ([]Job, error)
	FetchJobByID(id string) (*Job, error)
}

type SQLiteJobRepository struct {
	db     *sql.DB
	logger *utils.Logger
}

func (r *SQLiteJobRepository) CreateJob(
	nReducers int, startTime int64,
	id, inputPath, inputType, outputPath string,
	useSSL bool) (Job, error) {
	inputDataID := 0
	transactionLogic := func(tx *sql.Tx) error {
		query := "SELECT location FROM output_location WHERE location=?;"
		r.logger.Trace(query)
		if err := tx.QueryRow(query, outputPath).Scan(); errors.Is(err, sql.ErrNoRows) {
			query = "INSERT INTO output_location VALUES (?, ?);"
			r.logger.Trace(query)
			_, err := tx.Exec(query, outputPath, useSSL)
			if err != nil {
				return err
			}
		}

		query = "INSERT INTO input_data VALUES (NULL, ?, ?, NULL, NULL);"
		r.logger.Trace(query)
		res, err := tx.Exec(query, inputPath, inputType)
		if err != nil {
			return err
		}

		inputId, err := res.LastInsertId()
		inputDataID = int(inputId)
		if err != nil {
			return err
		}
		query = "INSERT INTO job VALUES (?, ?, ?, ?, ?, NULL);"
		r.logger.Trace(query)
		_, err = tx.Exec(query, id, nReducers, outputPath, inputId, startTime)
		return err
	}

	if err := runInTx(r.db, transactionLogic); err != nil {
		r.logger.Error(err.Error())
		return Job{}, err
	}

	return Job{
		Id:        id,
		NReducers: nReducers,
		OutputLocation: OutputLocation{
			Location: outputPath,
			UseSSL:   useSSL,
		},
		InputData: InputData{
			Id:   inputDataID,
			Path: inputPath,
			Type: inputType,
		},
		StartTime: startTime,
	}, nil
}

func (r *SQLiteJobRepository) FetchJobs() ([]Job, error) {
	query := `SELECT j.id, j.n_reducers, o.location, o.use_ssl, i.id,
	i.path, i.type, i.split_start, i.split_end, j.start_time, j.end_time FROM job j
	JOIN input_data i ON i.id = j.input_id
	JOIN output_location o ON o.location = j.output_path;`

	r.logger.Trace(query)
	rows, err := r.db.Query(query)
	if err != nil {
		return []Job{}, err
	}
	defer rows.Close()
	jobs := []Job{}
	for rows.Next() {
		job := Job{}
		inputData := InputData{}
		outputLocation := OutputLocation{}

		err := rows.Scan(
			&job.Id,
			&job.NReducers,
			&outputLocation.Location,
			&outputLocation.UseSSL,
			&inputData.Id,
			&inputData.Path,
			&inputData.Type,
			&inputData.SplitStart,
			&inputData.SplitEnd,
			&job.StartTime,
			&job.EndTime)

		if err != nil {
			r.logger.Error(err.Error())
			return []Job{}, err
		}

		job.InputData = inputData
		job.OutputLocation = outputLocation
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (r *SQLiteJobRepository) FetchJobByID(id string) (*Job, error) {
	query := `SELECT j.id, j.n_reducers, o.location, o.use_ssl, i.id,
	i.path, i.type, i.split_start, i.split_end, j.start_time, j.end_time FROM job j
	JOIN input_data i ON i.id = j.input_id
	JOIN output_location o ON o.location = j.output_path
	WHERE j.id = ?;`

	r.logger.Trace(query)
	row := r.db.QueryRow(query, id)

	job := Job{}
	inputData := InputData{}
	outputLocation := OutputLocation{}
	err := row.Scan(
		&job.Id,
		&job.NReducers,
		&outputLocation.Location,
		&outputLocation.UseSSL,
		&inputData.Id,
		&inputData.Path,
		&inputData.Type,
		&inputData.SplitStart,
		&inputData.SplitEnd,
		&job.StartTime,
		&job.EndTime)

	if errors.Is(err, sql.ErrNoRows) {
		r.logger.Warn("No job with id %s was found", id)
		return nil, nil
	}
	if err != nil {
		r.logger.Error(err.Error())
		return nil, err
	}
	job.InputData = inputData
	job.OutputLocation = outputLocation
	return &job, nil

}

func NewSQLiteJobsRepository(db *sql.DB) JobRepository {
	return &SQLiteJobRepository{
		db:     db,
		logger: utils.GetLogger(),
	}
}
