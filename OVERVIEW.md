<div>
    <h1>
		Kraicklist
    </h1>
    <p>Find your needs</p>
</div>

- [Demo](#demo)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
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

## Installation
- Setup environment variables on `.env`
    ```
    PORT=8080
    SEARCH_ENGINE_API_URL=http://localhost:8108
    SEARCH_ENGINE_API_KEY=xyz
    SEARCH_ENGINE_COLLECTION_NAME=ads
    DATA_FILE=data.txt
    ```
- Setup the `Typesense` by run `docker compose up -d` 

## Quick Start
- Use `go run main.go` command
- Open `http://localhost:8080`

## Recommendation
- 