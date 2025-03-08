---
slug: /api
sidebar_position: 4
---
# REST API Reference

## Artifacts API
### `/api/v1/artifacts`
The `/api/v1/artifacts` endpoint is used to create artifacts and get their metadata.
#### List all available artifacts
| Method            | Path |
|-------------------|------|
| `/api/v1/artifacts` | `GET`  |
#### Sample request
```bash
curl --location 'http://localhost:4750/api/v1/artifacts' --header 'Content-Type: application/json' | jq .
```
#### Sample response
```json
[
{
    "name": "name-1",
    "type": "type",
    "size": 0,
    "hash": "hash"
},
{
    "name": "name-2",
    "type": "type",
    "size": 0,
    "hash": "hash"
}
]
```
#### Create new artifact
| Method            | Path |
|-------------------|------|
| `/api/v1/artifacts` | `PUT`  |
#### Sample request
```bash
curl --request PUT --location 'http://localhost:4750/api/v1/artifacts' \
--form 'program=@"/path/to/binary/executable"'
```
#### Sample response
```json
{
    "name": "name",
    "type": "type",
    "size": 0,
    "hash": "hash"
}
```
### `/api/v1/artifacts/{filename}` 
The `/api/v1/artifacts/{filename}` endpoint is used to get a specific artifact's metadata or delete it.
#### Get artifact metadata
| Method            | Path |
|-------------------|------|
| `/api/v1/artifacts/{filename}` | `GET`  |
#### Sample request
```bash
curl --location 'http://localhost:4750/api/v1/artifacts/{filename}' --header 'Content-Type: application/json' | jq .
```
#### Sample response
```json
{
    "name": "name",
    "type": "type",
    "size": 0,
    "hash": "hash"
}
```
#### Delete artifact
| Method            | Path |
|-------------------|------|
| `/api/v1/artifacts/{filename}` | `DELETE`  |
#### Sample request
```bash
curl --request DELETE --location 'http://localhost:4750/api/v1/artifacts/{filename}' --header 'Content-Type: application/json' | jq .
```
#### Sample response
:::info

If the delete operation happened with no issues then the coordinator server should return a No Content HTTP code. 

:::
## Jobs API
### `/api/v1/jobs`
The `/api/v1/jobs` endpoint is used to schedule jobs and fetch all the jobs that were scheduled.
#### List all scheduled jobs
| Method            |  Path  |
|-------------------|--------|
| `/api/v1/jobs`    | `GET`  |
#### Sample request
```bash
curl --location 'http://localhost:4750/api/v1/jobs' --header 'Content-Type: application/json' | jq .
```
#### Sample response
```json
[
{
    "id": "UUID",
    "nReducers": 1,
    "outputLocation": {
        "location":"location",
        "useSSL": false
    },
    "inputData": {
        "id": "id",
        "path": "path",
        "type": "type",
        "splitStart": 0,
        "splitEnd": 12000
    },
    "startTime": 1741436779000
},
{
    "id": "UUID",
    "nReducers": 1,
    "outputLocation": {
        "location":"location",
        "useSSL": false
    },
    "inputData": {
        "id": "id",
        "path": "path",
        "type": "type",
        "splitStart": 0,
        "splitEnd": 12000
    },
    "startTime": 1741436779000
}
]
```
#### List all scheduled jobs
| Method            |  Path  |
|-------------------|--------|
| `/api/v1/jobs`    | `POST`  |
#### Sample request
```bash
curl --location 'http://localhost:4750/api/v1/jobs' --header 'Content-Type: application/json' --data '{
    "nReducers": 1,
    "inputPath": "http://minio:9000/test/test.txt",
    "inputType": "file/txt",
    "outputPath": "http://minio:9000",
    "useSSL": false,
    "mapperName": "mapper",
    "reducerName": "reducer",
    "inputStorageCredentials": {
        "username": "username",
        "password": "password"
    },
    "outputStorageCredentials": {
        "username": "username",
        "password": "password"
    }
}' | jq .
```
#### Sample response
```json
{
    "id": "UUID",
    "nReducers": 1,
    "outputLocation": {
        "location":"location",
        "useSSL": false
    },
    "inputData": {
        "id": "id",
        "path": "path",
        "type": "type",
        "splitStart": 0,
        "splitEnd": 12000
    },
    "startTime": 1741436779000
}
```
### `/api/v1/jobs/{id}`
The `/api/v1/jobs/{id}` endpoint is used to get a specific job's metadata or stop it.
#### Get job metadata
| Method            |  Path  |
|-------------------|--------|
| `/api/v1/jobs/{id}`    | `GET`  |
#### Sample request
```bash
curl --location 'http://localhost:4750/api/v1/jobs/{id}' --header 'Content-Type: application/json' | jq .
```
#### Sample response
```json
{
    "id": "UUID",
    "nReducers": 1,
    "outputLocation": {
        "location":"location",
        "useSSL": false
    },
    "inputData": {
        "id": "id",
        "path": "path",
        "type": "type",
        "splitStart": 0,
        "splitEnd": 12000
    },
    "startTime": 1741436779000
},
```
#### Stop job
| Method            |  Path  |
|-------------------|--------|
| `/api/v1/jobs/{id}`    | `DELETE`  |
#### Sample request
```bash
curl --request DELETE --location 'http://localhost:4750/api/v1/jobs' --header 'Content-Type: application/json' | jq .
```
#### Sample response
```
Job id was successfully stopped
```
### `/api/v1/jobs/{id}/tasks`
The `/api/v1/jobs/{id}/tasks` endpoint is used to get a specific job's task .
#### Get job tasks
| Method            |  Path  |
|-------------------|--------|
| `/api/v1/jobs/{id}/tasks`    | `GET`  |
#### Sample request
```bash
curl --location 'http://localhost:4750/api/v1/jobs/{id}/tasks' --header 'Content-Type: application/json' | jq .
```
#### Sample response
```json
[
{
    "id": "id-1",
    "type": "type",
    "status": "status",
    "program": {
        "name": "name",
        "type": "type",
        "size": 0,
        "hash": "hash"
    },
    "inputData": {
        "id": "id",
        "path": "path",
        "type": "type",
        "splitStart": 0,
        "splitEnd": 12000
    },
    "podName": "podName",
    "startTime": 1741436779000
},
{
    "id": "id-2",
    "type": "type",
    "status": "status",
    "program": {
        "name": "name",
        "type": "type",
        "size": 0,
        "hash": "hash"
    },
    "inputData": {
        "id": "id",
        "path": "path",
        "type": "type",
        "splitStart": 0,
        "splitEnd": 12000
    },
    "podName": "podName",
    "startTime": 1741436779000
}
]
```
