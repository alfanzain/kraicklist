package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"challenge.haraj.com.sa/kraicklist/core"
	"challenge.haraj.com.sa/kraicklist/driven/searchengine"
	"challenge.haraj.com.sa/kraicklist/driver"
	"github.com/caarlos0/env/v11"
)

type config struct {
	Port                       string `env:"PORT,required" envDefault:"8080"`
	SearchEngineApiUrl         string `env:"SEARCH_ENGINE_API_URL,required" envDefault:"http://localhost:8108"`
	SearchEngineApiKey         string `env:"SEARCH_ENGINE_API_KEY,required" envDefault:"xyz"`
	SearchEngineCollectionName string `env:"SEARCH_ENGINE_COLLECTION_NAME,required" envDefault:"ads"`
	DataFile                   string `env:"DATA_FILE,required" envDefault:"data.txt"`
}

func main() {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// initialize Typesense search engine
	searchEngine, err := searchengine.NewSearchEngine(searchengine.TypesenseSearchEngineConfig{
		BaseApiUrl:     cfg.SearchEngineApiUrl,
		ApiKey:         cfg.SearchEngineApiKey,
		CollectionName: cfg.SearchEngineCollectionName,
	})
	if err != nil {
		log.Fatalf("failed initialize search engine: %v", err)
	}

	_, err = searchEngine.ImportData(ctx, "data.txt")
	if err != nil {
		log.Fatalf("failed import data: %v", err)
	}

	// initialize service
	svc, err := core.NewService(core.ServiceConfig{
		SearchEngine: searchEngine,
	})
	if err != nil {
		log.Fatalf("failed initialize service: %v", err)
	}

	// initialize handler
	api, err := driver.NewApi(driver.ApiConfig{
		Ctx:     ctx,
		Service: svc,
	})
	if err != nil {
		log.Fatalf("failed to initialize API: %v", err)
	}

	// start server
	fmt.Println("Server is listening on", cfg.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), api.GetHandler())
	if err != nil {
		fmt.Println("Server is listeninsasdsadg on", cfg.Port)
		log.Fatalf("unable to start server due: %v", err)
	}
}
