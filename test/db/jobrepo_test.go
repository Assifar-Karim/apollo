package db

import (
	"os"
	"testing"
	"time"

	"github.com/Assifar-Karim/apollo/internal/db"
)

func TestCreateJob(t *testing.T) {
	// Given
	nReducers := 1
	startTime := time.Now().UTC().UnixMilli()
	id := "id"
	inputPath := "input-path"
	inputType := "input-type"
	outputPath := "output-path"
	useSSL := false

	expectedJob := db.Job{
		Id:        id,
		NReducers: nReducers,
		OutputLocation: db.OutputLocation{
			Location: outputPath,
			UseSSL:   useSSL,
		},
		InputData: db.InputData{
			Id:   1,
			Path: inputPath,
			Type: inputType,
		},
		StartTime: startTime,
	}
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	jobRepo := db.NewSQLiteJobsRepository(database)

	// When
	job, err := jobRepo.CreateJob(nReducers, startTime, id, inputPath, inputType, outputPath, useSSL)
	if err != nil {
		t.Fatalf("The job creation operation failed! %v", err)
	}

	// Then
	if job.Id != expectedJob.Id ||
		job.NReducers != expectedJob.NReducers ||
		job.StartTime != expectedJob.StartTime ||
		job.OutputLocation.Location != expectedJob.OutputLocation.Location ||
		job.OutputLocation.UseSSL != expectedJob.OutputLocation.UseSSL ||
		job.InputData.Id != expectedJob.InputData.Id ||
		job.InputData.Path != expectedJob.InputData.Path ||
		job.InputData.Type != expectedJob.InputData.Type {
		t.Errorf("Expected %v but found %v!", expectedJob, job)
	}
}

func TestFetchJobsWhenNoJobExists(t *testing.T) {
	// Given
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	jobRepo := db.NewSQLiteJobsRepository(database)

	// When
	jobs, err := jobRepo.FetchJobs()
	if err != nil {
		t.Fatalf("The job fetch operation failed! %v", err)
	}

	// Then
	if len(jobs) != 0 {
		t.Errorf("Expected to find no jobs but found %v", jobs)
	}
}

func TestFetchJobsWhenJobsExist(t *testing.T) {
	// Given
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	jobRepo := db.NewSQLiteJobsRepository(database)
	job, err := jobRepo.CreateJob(1, time.Now().UnixMilli(), "id", "input-path", "input-type", "output-path", false)
	if err != nil {
		t.Fatal("Couldn't populate db with job for test logic!")
	}

	// When
	jobs, err := jobRepo.FetchJobs()
	if err != nil {
		t.Fatalf("The job fetch operation failed! %v", err)
	}

	// Then
	if len(jobs) != 1 && (job.Id != jobs[0].Id ||
		job.NReducers != jobs[0].NReducers ||
		job.StartTime != jobs[0].StartTime ||
		job.OutputLocation.Location != jobs[0].OutputLocation.Location ||
		job.OutputLocation.UseSSL != jobs[0].OutputLocation.UseSSL ||
		job.InputData.Id != jobs[0].InputData.Id ||
		job.InputData.Path != jobs[0].InputData.Path ||
		job.InputData.Type != jobs[0].InputData.Type) {
		t.Errorf("Expected to find 1 job but found %v", jobs)
	}
}

func TestFetchJobByIDWhenJobDoesNotExist(t *testing.T) {
	// Given
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	jobRepo := db.NewSQLiteJobsRepository(database)

	// When
	job, err := jobRepo.FetchJobByID("id")

	// Then
	if job != nil || err != nil {
		t.Errorf("Expected to find no job but found %v, %v", job, err)
	}
}

func TestFetchJobByIDWhenJobExists(t *testing.T) {
	// Given
	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	jobRepo := db.NewSQLiteJobsRepository(database)
	job, err := jobRepo.CreateJob(1, time.Now().UnixMilli(), "id", "input-path", "input-type", "output-path", false)
	if err != nil {
		t.Fatal("Couldn't populate db with job for test logic!")
	}

	// When
	fetchedJob, err := jobRepo.FetchJobByID(job.Id)
	if err != nil {
		t.Fatalf("The job fetch operation failed! %v", err)
	}

	// Then
	if job.Id != fetchedJob.Id ||
		job.NReducers != fetchedJob.NReducers ||
		job.StartTime != fetchedJob.StartTime ||
		job.OutputLocation.Location != fetchedJob.OutputLocation.Location ||
		job.OutputLocation.UseSSL != fetchedJob.OutputLocation.UseSSL ||
		job.InputData.Id != fetchedJob.InputData.Id ||
		job.InputData.Path != fetchedJob.InputData.Path ||
		job.InputData.Type != fetchedJob.InputData.Type {
		t.Errorf("Expected to find %v but found %v", job, fetchedJob)
	}
}

func TestUpdateJobEndTimeByIDWhenJobExists(t *testing.T) {
	// Given
	id := "id"
	endTs := time.Now().UnixMilli()

	database, dbName, err := setupDB()
	t.Cleanup(func() { os.Remove(dbName) })
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	jobRepo := db.NewSQLiteJobsRepository(database)
	_, err = jobRepo.CreateJob(1, time.Now().UnixMilli(), id, "input-path", "input-type", "output-path", false)
	if err != nil {
		t.Fatal("Couldn't populate db with job for test logic!")
	}

	// When
	if err := jobRepo.UpdateJobEndTimeByID(id, endTs); err != nil {
		t.Fatalf("Update operation failed! %v", err)
	}

	// Then
	fetchedEndTs := int64(0)
	row := database.QueryRow("SELECT end_time FROM job WHERE id = id;")
	if err := row.Scan(&fetchedEndTs); err != nil {
		t.Fatalf("Data fetching operation failed for verification failed! %v", err)
	}
	if fetchedEndTs != endTs {
		t.Errorf("Expected %v but found %v", endTs, fetchedEndTs)
	}
}
