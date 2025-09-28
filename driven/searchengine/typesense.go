package searchengine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/typesense/typesense-go/v3/typesense"
	"github.com/typesense/typesense-go/v3/typesense/api"
	"github.com/typesense/typesense-go/v3/typesense/api/pointer"
)

type TypesenseSearchEngine struct {
	TypesenseSearchEngineConfig
	client *typesense.Client
}

type TypesenseSearchEngineConfig struct {
	BaseApiUrl     string `validate:"nonzero"`
	ApiKey         string `validate:"nonzero"`
	CollectionName string `validate:"nonzero"`
}

func NewSearchEngine(cfg TypesenseSearchEngineConfig) (*TypesenseSearchEngine, error) {
	client := typesense.NewClient(
		typesense.WithServer(cfg.BaseApiUrl),
		typesense.WithAPIKey(cfg.ApiKey),
		typesense.WithConnectionTimeout(100*time.Second),
		typesense.WithCircuitBreakerMaxRequests(50),
		typesense.WithCircuitBreakerInterval(2*time.Minute),
		typesense.WithCircuitBreakerTimeout(30*time.Minute),
	)

	return &TypesenseSearchEngine{
		TypesenseSearchEngineConfig: cfg,
		client:                      client,
	}, nil
}

func (tse *TypesenseSearchEngine) CreateCollection(ctx context.Context, schema *api.CollectionSchema) error {
	_, err := tse.client.Collections().Create(ctx, schema)
	return err
}

func (tse *TypesenseSearchEngine) GetCollection(ctx context.Context) (*api.CollectionResponse, error) {
	return tse.client.Collection(tse.CollectionName).Retrieve(ctx)
}

func (tse *TypesenseSearchEngine) DropCollection(ctx context.Context) error {
	_, err := tse.client.Collection(tse.CollectionName).Delete(ctx)
	return err
}

func (tse *TypesenseSearchEngine) GetNumDocuments(ctx context.Context) (int64, error) {
	collection, err := tse.GetCollection(ctx)
	if err != nil {
		return 0, err
	}
	return int64(*collection.NumDocuments), nil
}

func (tse *TypesenseSearchEngine) SearchCollection(ctx context.Context, searchParams *api.SearchCollectionParams) (*api.SearchResult, error) {
	return tse.client.Collection(tse.CollectionName).Documents().Search(ctx, searchParams)
}

func (tse *TypesenseSearchEngine) Search(
	ctx context.Context,
	query string,
	perPage int,
	page int,
) (*api.SearchResult, error) {
	searchParams := &api.SearchCollectionParams{
		Q: pointer.String(query),
		// QueryBy:              pointer.String("title,tags"),
		QueryBy:              pointer.String("title, embedding"),
		VectorQuery:          pointer.String("embedding:([], alpha: 0.8, distance_threshold: 1.0)"),
		PrioritizeExactMatch: pointer.True(),
		DropTokensThreshold:  pointer.Int(0),
		ExcludeFields:        pointer.String("embedding"),
		PerPage:              pointer.Int(perPage),
		Page:                 pointer.Int(page),
	}

	// if filterBy != "" {
	// 	searchParams.FilterBy = pointer.String(filterBy)
	// }

	// if sortBy != "" {
	// 	searchParams.SortBy = &sortBy
	// }

	// Log the search parameters in a readable format
	paramsJSON, err := json.MarshalIndent(searchParams, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling search params: %v\n", err)
	} else {
		fmt.Printf("Searching with params:\n%s\n", string(paramsJSON))
	}

	return tse.SearchCollection(ctx, searchParams)
}

func (tse *TypesenseSearchEngine) LoadData(filepath string) ([]interface{}, error) {
	fmt.Println("Loading data...")

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("unable to open source file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var documents []interface{}

	for scanner.Scan() {
		line := scanner.Bytes()

		var doc map[string]interface{}
		if err := json.Unmarshal(line, &doc); err != nil {
			continue
		}

		// Ensure ID is string type for Typesense
		if idVal, ok := doc["id"]; ok {
			switch v := idVal.(type) {
			case float64:
				doc["id"] = fmt.Sprintf("%.0f", v)
			case int:
				doc["id"] = fmt.Sprintf("%d", v)
			case int64:
				doc["id"] = fmt.Sprintf("%d", v)
			case json.Number:
				doc["id"] = v.String()
			case string:
				// Already a string, keep as is
			default:
				doc["id"] = fmt.Sprintf("%v", v)
			}
		} else {
			return nil, fmt.Errorf("missing 'id' field in document")
		}

		documents = append(documents, doc)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %v", err)
	}

	fmt.Printf("Loaded %d documents\n", len(documents))
	return documents, nil
}

func (tse *TypesenseSearchEngine) ImportData(ctx context.Context, filepath string) ([]*api.ImportDocumentResponse, error) {
	// Drop existing collection
	fmt.Println("Dropping existing collection...")
	tse.DropCollection(ctx)

	// Keyword search schema
	// schema := &api.CollectionSchema{
	// 	Name: tse.CollectionName,
	// 	Fields: []api.Field{
	// 		{
	// 			Name: "title",
	// 			Type: "string",
	// 		},
	// 		{
	// 			Name:  "content",
	// 			Type:  "string",
	// 			Index: pointer.False(),
	// 		},
	// 		{
	// 			Name:  "tags",
	// 			Type:  "string[]",
	// 			Facet: pointer.True(),
	// 		},
	// 		{
	// 			Name: "updated_at",
	// 			Type: "int64",
	// 		},
	// 		{
	// 			Name:     "thumb_url",
	// 			Type:     "string",
	// 			Optional: pointer.True(),
	// 		},
	// 		{
	// 			Name:     "image_urls",
	// 			Type:     "string[]",
	// 			Optional: pointer.True(),
	// 		},
	// 	},
	// 	DefaultSortingField: pointer.String("updated_at"),
	// }

	// Semantic search
	schema := &api.CollectionSchema{
		Name: tse.CollectionName,
		Fields: []api.Field{
			{
				Name: "title",
				Type: "string",
			},
			{
				Name: "content",
				Type: "string",
			},
			{
				Name:  "tags",
				Type:  "string[]",
				Facet: pointer.True(),
			},
			{
				Name: "updated_at",
				Type: "int64",
			},
			{
				Name:     "thumb_url",
				Type:     "string",
				Optional: pointer.True(),
			},
			{
				Name:     "image_urls",
				Type:     "string[]",
				Optional: pointer.True(),
			},
			{
				Name: "embedding",
				Type: "float[]",
				Embed: &struct {
					From        []string `json:"from"`
					ModelConfig struct {
						AccessToken    *string `json:"access_token,omitempty"`
						ApiKey         *string `json:"api_key,omitempty"`
						ClientId       *string `json:"client_id,omitempty"`
						ClientSecret   *string `json:"client_secret,omitempty"`
						IndexingPrefix *string `json:"indexing_prefix,omitempty"`
						ModelName      string  `json:"model_name"`
						ProjectId      *string `json:"project_id,omitempty"`
						QueryPrefix    *string `json:"query_prefix,omitempty"`
						RefreshToken   *string `json:"refresh_token,omitempty"`
						Url            *string `json:"url,omitempty"`
					} `json:"model_config"`
				}{
					From: []string{"title", "tags", "content"},
					ModelConfig: struct {
						AccessToken    *string `json:"access_token,omitempty"`
						ApiKey         *string `json:"api_key,omitempty"`
						ClientId       *string `json:"client_id,omitempty"`
						ClientSecret   *string `json:"client_secret,omitempty"`
						IndexingPrefix *string `json:"indexing_prefix,omitempty"`
						ModelName      string  `json:"model_name"`
						ProjectId      *string `json:"project_id,omitempty"`
						QueryPrefix    *string `json:"query_prefix,omitempty"`
						RefreshToken   *string `json:"refresh_token,omitempty"`
						Url            *string `json:"url,omitempty"`
					}{
						// ModelName: "ts/e5-large-v2",
						ModelName: "ts/all-MiniLM-L12-v2",
					},
				},
			},
		},
		DefaultSortingField: pointer.String("updated_at"),
	}

	// Create collection
	fmt.Println("Creating collection...")
	err := tse.CreateCollection(ctx, schema)
	if err != nil {
		return nil, fmt.Errorf("error creating collection: %v", err)
	}

	// Load data
	documents, err := tse.LoadData(filepath)
	if err != nil {
		return nil, err
	}

	// Import parameters
	params := &api.ImportDocumentsParams{
		Action:    (*api.IndexAction)(pointer.String(string(api.Create))),
		BatchSize: pointer.Int(10),
	}

	// Import documents with retry logic
	fmt.Println("Importing documents...")
	maxRetries := 3
	var importResp []*api.ImportDocumentResponse

	for i := 0; i < maxRetries; i++ {
		importResp, err = tse.client.Collection(tse.CollectionName).Documents().Import(ctx, documents, params)
		if err != nil {
			if i < maxRetries-1 {
				fmt.Printf("Import attempt %d failed: %v, retrying...\n", i+1, err)
				time.Sleep(time.Duration(i+2) * time.Second) // Exponential backoff
				continue
			}
			return nil, fmt.Errorf("failed to import documents after %d attempts: %v", maxRetries, err)
		}
		break
	}

	// fmt.Printf("Import completed. Response: %+v\n", importResp)
	fmt.Printf("Import completed.\n")

	time.Sleep(2 * time.Second)

	// Get document count
	count, err := tse.GetNumDocuments(ctx)
	if err != nil {
		return importResp, fmt.Errorf("failed to get document count: %v", err)
	}

	// Create synonyms
	err = tse.SeedSynonyms(ctx)
	if err != nil {
		return importResp, fmt.Errorf("failed to seed synonyms: %v", err)
	}

	fmt.Printf("Number of documents in collection '%s': %d\n", tse.CollectionName, count)
	return importResp, nil
}

func (tse *TypesenseSearchEngine) CreateSynonyms(ctx context.Context, id string, schema *api.SearchSynonymSchema) error {
	_, err := tse.client.Collection(tse.CollectionName).Synonyms().Upsert(ctx, id, schema)
	if err != nil {
		return fmt.Errorf("failed to create android phone synonyms: %w", err)
	}

	return nil
}

func (tse *TypesenseSearchEngine) ListSynonyms(ctx context.Context) ([]*api.SearchSynonym, error) {
	synonyms, err := tse.client.Collection(tse.CollectionName).Synonyms().Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve synonyms: %w", err)
	}

	fmt.Printf("Found %d synonym groups:\n", len(synonyms))
	for _, synonym := range synonyms {
		fmt.Printf("Synonym ID: %s\n", *synonym.Id)
		fmt.Printf("  Synonyms: %v\n", synonym.Synonyms)
		if synonym.Root != nil {
			fmt.Printf("  Root: %s\n", *synonym.Root)
		}
	}

	return synonyms, nil
}

func (tse *TypesenseSearchEngine) GetSynonym(ctx context.Context, synonymID string) (*api.SearchSynonym, error) {
	synonym, err := tse.client.Collection(tse.CollectionName).Synonym(synonymID).Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve synonym %s: %w", synonymID, err)
	}

	fmt.Printf("Synonym ID: %s\n", *synonym.Id)
	fmt.Printf("Synonyms: %v\n", synonym.Synonyms)
	if synonym.Root != nil {
		fmt.Printf("Root: %s\n", *synonym.Root)
	}

	return synonym, nil
}

func (tse *TypesenseSearchEngine) DeleteSynonym(ctx context.Context, synonymID string) error {
	_, err := tse.client.Collection(tse.CollectionName).Synonym(synonymID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete synonym %s: %w", synonymID, err)
	}

	fmt.Printf("Successfully deleted synonym: %s\n", synonymID)
	return nil
}

// Seedings

func (tse *TypesenseSearchEngine) SeedSynonyms(ctx context.Context) error {
	var err error

	// Android
	err = tse.CreateSynonyms(ctx, "android-phone-synonyms", &api.SearchSynonymSchema{
		Synonyms: []string{
			"android phone",
			"samsung galaxy",
			"galaxy phone",
			"android smartphone",
			"samsung smartphone",
			"android mobile",
			"galaxy mobile",
			"samsung mobile",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create android phone synonyms: %w", err)
	}

	synonymsResp, err := tse.ListSynonyms(ctx)
	if err != nil {
		return fmt.Errorf("failed to list synonyms: %w", err)
	}

	fmt.Printf("Successfully seeding %d synonym groups\n", len(synonymsResp))
	return nil
}
