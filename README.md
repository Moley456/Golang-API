# Golang API Server

## Note

The `.env` file is committed just for submission purposes.

## To Build Locally

### Prerequisites

- Installed Docker >=20.10.17 and Docker Compose V2

### Steps

1. In the root folder, run `docker compose --profile prod build`
2. Then run `docker compose --profile prod up`
3. The server should be accessible on port 8080

## To Run Locally

### Prerequisites

- Installed Docker >=20.10.17 and Docker Compose V2
- Installed Golang v1.22.1

### Steps

1. In the root folder, run `docker compose --profile dev build`
2. Then run `docker compose --profile dev up`
3. Then run `go run .` to start the server on port 8080

## To Run Tests

### Prerequisites

- Installed Docker >=20.10.17 and Docker Compose V2

### Steps

1. In the root folder, run `docker compose --profile test build`
2. Then run `docker compose --profile test up`
3. Wait for the tests to finish running.
