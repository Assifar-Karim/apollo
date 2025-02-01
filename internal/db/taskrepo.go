package db

import (
	"database/sql"
	"fmt"

	"github.com/Assifar-Karim/apollo/internal/utils"
)

type TaskRepository interface {
	CreateTasksBatch(jobId, taskType string, pods []string, inputs []InputData, program Artifact, startTime int64, count int) ([]Task, error)
	FetchTasksByJobID(jobId string) ([]Task, error)
	UpdateTaskStatusByID(id, status string) error
	UpdateTaskEndTimeByID(id string, endTs int64) error
	UpdateUnfinishedTasksStatusByJobID(status, jobId string) error
}

type SQLiteTaskRepository struct {
	db     *sql.DB
	logger *utils.Logger
}

func (r *SQLiteTaskRepository) CreateTasksBatch(jobId, taskType string,
	pods []string, inputs []InputData, program Artifact, startTime int64, count int) ([]Task, error) {

	tasks := make([]Task, count)
	transactionLogic := func(tx *sql.Tx) error {
		if len(inputs) > 0 {
			query := `INSERT INTO input_data (id, path, type, split_start, split_end) VALUES `
			queryParams := []any{}
			for _, inputData := range inputs {
				queryParams = append(queryParams, inputData.Path, inputData.Type, inputData.SplitStart, inputData.SplitEnd)
				query += `(NULL, ?, ?, ?, ?),`
			}
			query = query[:len(query)-1] + ";"
			r.logger.Trace(query)
			res, err := tx.Exec(query, queryParams...)
			if err != nil {
				return err
			}
			lastInputId, err := res.LastInsertId()
			if err != nil {
				return err
			}
			offset := int(lastInputId) - len(inputs) + 1
			for i := range inputs {
				inputs[i].Id = offset + i
			}
		}

		query := `INSERT INTO task (
		id, job_id, type, program_name, input_data_id,
		pod_name, start_time, end_time) VALUES `
		queryParams := []any{}
		for i := 0; i < count; i++ {
			id := fmt.Sprintf("%s-%c-%v", jobId, taskType[0], i)
			task := Task{
				Id:        id,
				Type:      taskType,
				Status:    "scheduled",
				Program:   program,
				PodName:   &pods[i],
				StartTime: startTime,
			}
			if len(inputs) > 0 {
				task.InputData = &inputs[i]
				queryParams = append(queryParams,
					task.Id,
					jobId,
					task.Type,
					program.Name,
					task.InputData.Id,
					*task.PodName,
					task.StartTime)
				query += `(?, ?, ?, ?, ?, ?, ?, NULL),`
			} else {
				queryParams = append(queryParams,
					task.Id,
					jobId,
					task.Type,
					program.Name,
					*task.PodName,
					task.StartTime)
				query += `(?, ?, ?, ?, NULL, ?, ?, NULL),`
			}
			tasks[i] = task
		}
		query = query[:len(query)-1] + ";"
		r.logger.Trace(query)
		_, err := tx.Exec(query, queryParams...)
		return err
	}
	if err := runInTx(r.db, transactionLogic); err != nil {
		r.logger.Error(err.Error())
		return []Task{}, err
	}
	return tasks, nil
}

func (r *SQLiteTaskRepository) FetchTasksByJobID(jobId string) ([]Task, error) {
	query := `SELECT t.id, t.type, t.status, t.pod_name, t.start_time, t.end_time,
	a.name, a.type, a.size, a.hash,
	i.id, i.path, i.type, i.split_start, i.split_end
	FROM task t 
	JOIN artifact a ON a.name = t.program_name
	LEFT OUTER JOIN input_data i ON i.id = t.input_data_id
	WHERE t.job_id = ?;`

	r.logger.Trace(query)
	rows, err := r.db.Query(query, jobId)
	if err != nil {
		r.logger.Error(err.Error())
		return []Task{}, err
	}
	defer rows.Close()
	tasks := []Task{}
	for rows.Next() {
		task := Task{}
		inputData := InputData{}
		artifact := Artifact{}
		// input data scan verification vars
		var iId sql.NullInt32
		var iPath, iType sql.NullString
		err := rows.Scan(
			&task.Id,
			&task.Type,
			&task.Status,
			&task.PodName,
			&task.StartTime,
			&task.EndTime,
			&artifact.Name,
			&artifact.Type,
			&artifact.Size,
			&artifact.Hash,
			&iId,
			&iPath,
			&iType,
			&inputData.SplitStart,
			&inputData.SplitEnd)

		if err != nil {
			r.logger.Error(err.Error())
			return []Task{}, err
		}

		if iId.Valid && iPath.Valid && iType.Valid {
			inputData.Id = int(iId.Int32)
			inputData.Path = iPath.String
			inputData.Type = iType.String
			task.InputData = &inputData
		}

		task.Program = artifact
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *SQLiteTaskRepository) UpdateTaskStatusByID(id, status string) error {
	query := "UPDATE task SET status = ? WHERE id = ?;"
	r.logger.Trace(query)
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *SQLiteTaskRepository) UpdateTaskEndTimeByID(id string, endTs int64) error {
	query := "UPDATE task SET end_time = ? WHERE id = ?;"
	r.logger.Trace(query)
	_, err := r.db.Exec(query, endTs, id)
	return err
}

func (r *SQLiteTaskRepository) UpdateUnfinishedTasksStatusByJobID(status, jobId string) error {
	query := "UPDATE task SET status = ? WHERE job_id = ? AND status != completed"
	r.logger.Trace(query)
	_, err := r.db.Exec(query, status, jobId)
	return err
}

func NewSQLiteTaskRepository(db *sql.DB) TaskRepository {
	return &SQLiteTaskRepository{
		db:     db,
		logger: utils.GetLogger(),
	}
}
