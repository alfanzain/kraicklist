# KraickList

Fun startup advertisement listings called KraickList. This app solves the users to search an ads published in the startup.

This is a public challenge previously for Haraj take home challenge forked from: https://github.com/riandyrn/kraicklist

If you face any issues, please share the details on the issue or reach me at alfanzainkp@gmail.com


## Table of Contents

- [KraickList](#kraicklist)
  - [Table of Contents](#table-of-contents)
  - [Demo](#demo)
  - [Features](#features)
  - [Prerequisites](#prerequisites)
  - [Quick Start](#quick-start)
  - [Indexes (Typesense)](#indexes-typesense)
    - [Get the collection](#get-the-collection)
  - [About Sample Data](#about-sample-data)
  - [Further improvement](#further-improvement)

## Demo
https://kraicklist.my.id/

## Features
- Full-text search with semantic search using Typesense
- Enhanced UX

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

## Indexes (Typesense)

### Get the collection

To check the existing collection, we can use this outside the application

```
curl -H "X-TYPESENSE-API-KEY: ${SEARCH_ENGINE_API_KEY}" \
     -X GET \
    "http://localhost:8108/collections/ads"
```

## About Sample Data

The data is translated from Arabic to English using Google Translate, so sometimes you will find funny translation on it. 🤣

## Further improvement

- More enhance on UX
- Handle infinite scroll with threshold. Then "load more" after the threshold.
- Experiment with full text search + models