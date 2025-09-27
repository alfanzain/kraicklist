<div>
    <h1>
		Kraicklist
    </h1>
    <p>Find your needs</p>
</div>

- [Demo](#demo)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Recommendation](#recommendation)

## Demo
- 

## Features
- Full-text search with semantic search using Typesense

## Prerequisites
- Golang 1.24.1
- Docker
- Docker Compose

## Quick Start
Manual

- Setup environment variables on `.env`
    ```
    PORT=8080
    SEARCH_ENGINE_API_URL=http://localhost:8108
    SEARCH_ENGINE_API_KEY=xyz
    SEARCH_ENGINE_COLLECTION_NAME=ads
    DATA_FILE=data.txt
    ```
- Run `go run main.go`
- Open `http://localhost:8080`

Docker

- Setup environment variables
    ```
    export PORT=8080
    export SEARCH_ENGINE_API_URL=http://typesense:8108
    export SEARCH_ENGINE_API_KEY=xyz
    export SEARCH_ENGINE_COLLECTION_NAME=ads
    export DATA_FILE=data.txt
    ```
- Run `make run`
- Open `http://localhost:8080`

## Recommendation
- 