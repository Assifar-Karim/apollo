package db

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/Assifar-Karim/apollo/internal/db"
	_ "modernc.org/sqlite"
)

func TestSQLiteDBEagerLoad(t *testing.T) {
	// Given
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	driver := "sqlite"
	dbName := fmt.Sprintf("%s/test.db", currentDir)

	queries := make([]string, 5)

	queries[0] = `CREATE TABLE output_location (
    location VARCHAR PRIMARY KEY NOT NULL,
    use_SSL BOOLEAN NOT NULL)`

	queries[1] = `CREATE TABLE input_data (
    id INTEGER PRIMARY KEY NOT NULL,
    path VARCHAR NOT NULL,
    type VARCHAR NOT NULL,
    split_start INTEGER,
    split_end INTEGER)`

	queries[2] = `CREATE TABLE job (
    id VARCHAR PRIMARY KEY NOT NULL,
    n_reducers INTEGER NOT NULL,
    output_path VARCHAR NOT NULL,
    input_id INTEGER NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
	FOREIGN KEY(input_id) REFERENCES input_data(id),
	FOREIGN KEY(output_path) REFERENCES output_location(location))`

	queries[3] = `CREATE TABLE artifact (
    name VARCHAR PRIMARY KEY NOT NULL,
    type VARCHAR NOT NULL DEFAULT executable,
    size INTEGER NOT NULL DEFAULT 0,
	hash VARCHAR NOT NULL)`

	queries[4] = `CREATE TABLE task (
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
	FOREIGN KEY(program_name) REFERENCES artifact(name))`
	slices.Sort(queries)

	// When
	database, err := db.New(driver, dbName, true)
	if err != nil {
		t.Fatalf("Can't connect to database: %s", err)
	}
	defer os.Remove(dbName)

	// Then
	rows, err := database.Query("SELECT DISTINCT sql FROM SQLITE_MASTER;")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	fetchedQueries := []string{}
	for rows.Next() {
		fetchedQuery := ""
		rows.Scan(&fetchedQuery)
		if strings.TrimSpace(fetchedQuery) != "" {
			fetchedQueries = append(fetchedQueries, fetchedQuery)
		}
	}
	slices.Sort(fetchedQueries)
	for i := 0; i < 4; i++ {
		if queries[i] != fetchedQueries[i] {
			t.Errorf("Expected %s but found %s", queries[i], fetchedQueries[i])
		}
	}

}
