package searchengine

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type TypesenseSearchEngine struct {
	TypesenseSearchEngineConfig
}

type TypesenseSearchEngineConfig struct {
	BaseApiUrl     string `validate:"nonzero"`
	ApiKey         string `validate:"nonzero"`
	CollectionName string `validate:"nonzero"`
}

type SearchRequest struct {
	Q             string `json:"q"`
	QueryBy       string `json:"query_by"`
	Collection    string `json:"collection"`
	Prefix        string `json:"prefix"`
	VectorQuery   string `json:"vector_query,omitempty"`
	ExcludeFields string `json:"exclude_fields"`
	PerPage       int    `json:"per_page"`
	Page          int    `json:"page,omitempty"`
}

type MultiSearch struct {
	Searches []SearchRequest `json:"searches"`
}

func NewSearchEngine(cfg TypesenseSearchEngineConfig) (*TypesenseSearchEngine, error) {
	return &TypesenseSearchEngine{
		TypesenseSearchEngineConfig: cfg,
	}, nil
}

func (se *TypesenseSearchEngine) Search(
	q,
	queryBy,
	excludeFields string,
	perPage,
	page int,
) (interface{}, error) {
	url := se.BaseApiUrl + "/multi_search"

	payload := MultiSearch{
		Searches: []SearchRequest{
			{
				Q:             q,
				QueryBy:       queryBy,
				Collection:    se.CollectionName,
				Prefix:        "false",
				VectorQuery:   "embedding:([], k: 200)",
				ExcludeFields: excludeFields,
				PerPage:       perPage,
				Page:          page,
			},
		},
	}

	resp, err := se.doRequest(http.MethodPost, url, payload)
	if err != nil {
		return "", err
	}

	return resp, nil
}

type FieldEmbedConfig struct {
	From        []string `json:"from"`
	ModelConfig struct {
		ModelName string `json:"model_name"`
	} `json:"model_config"`
}

type Field struct {
	Name  string            `json:"name"`
	Type  string            `json:"type"`
	Embed *FieldEmbedConfig `json:"embed,omitempty"`
	Facet bool              `json:"facet,omitempty"`
	Index bool              `json:"index,omitempty"`
	Stem  bool              `json:"stem,omitempty"`
	Sort  bool              `json:"sort,omitempty"`
}

type CreateCollectionPayload struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

type Advertisement struct {
	ID        int64    `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	ThumbURL  string   `json:"thumb_url"`
	Tags      []string `json:"tags"`
	UpdatedAt int64    `json:"updated_at"`
	ImageURLs []string `json:"image_urls"`
}

type Advertisements []Advertisement

func (se *TypesenseSearchEngine) CreateCollection() (string, error) {
	payload := CreateCollectionPayload{
		Name: se.CollectionName,
		Fields: []Field{
			{Name: "title", Type: "string"},
			{Name: "content", Type: "string", Stem: true},
			{Name: "thumb_url", Type: "string", Index: false},
			{Name: "tags", Type: "string[]", Facet: true},
			{Name: "updated_at", Type: "int64", Facet: true, Sort: true},
			{Name: "image_urls", Type: "string[]", Index: false},
			{
				Name: "embedding",
				Type: "float[]",
				Embed: &FieldEmbedConfig{
					From: []string{"content"},
					ModelConfig: struct {
						ModelName string `json:"model_name"`
					}{
						ModelName: "ts/all-MiniLM-L12-v2",
					},
				},
			},
		},
	}

	fmt.Println("Creating collection...")

	url := se.BaseApiUrl + "/collections"
	resp, err := se.doRequest(http.MethodPost, url, payload)
	if err != nil {
		return "", fmt.Errorf("collection creation failed: %v", err)
	}
	return resp, nil
}

func (se *TypesenseSearchEngine) GetNumDocuments() (int, error) {
	url := se.BaseApiUrl + "/collections/" + se.CollectionName

	respStr, err := se.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	var collectionInfo struct {
		NumDocuments int `json:"num_documents"`
	}
	if err := json.Unmarshal([]byte(respStr), &collectionInfo); err != nil {
		return 0, fmt.Errorf("failed to parse collection info: %v", err)
	}

	return collectionInfo.NumDocuments, nil
}

func (se *TypesenseSearchEngine) DropCollection() (bool, error) {
	fmt.Println("Dropping collection...")

	url := se.BaseApiUrl + "/collections/" + se.CollectionName

	respStr, err := se.doRequest(http.MethodDelete, url, nil)
	if err != nil {
		fmt.Println("Error dropping collection:", err)
		return false, err
	}

	var resp struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return false, fmt.Errorf("failed to parse response: %v", err)
	}

	if resp.Name == se.CollectionName {
		return true, nil
	}

	return false, fmt.Errorf("unexpected response: %s", respStr)
}

func (se *TypesenseSearchEngine) ImportData(filepath string) (string, error) {
	se.DropCollection()
	_, err := se.CreateCollection()
	if err != nil {
		fmt.Println("Error creating collection:", err)
		return "", err
	}

	dataBuffer, err := se.LoadData(filepath)
	if err != nil {
		return "", err
	}

	url := se.BaseApiUrl + "/collections/" + se.CollectionName + "/documents/import?action=create"

	fmt.Println("Importing documents...")
	importResp, err := se.doStreamRequest(http.MethodPost, url, dataBuffer, "text/plain")
	if err != nil {
		fmt.Println("Error importing documents:", err)
		return "", fmt.Errorf("failed to import documents: %v", err)
	}

	fmt.Println("Import response:", importResp)

	count, err := se.GetNumDocuments()
	if err != nil {
		return importResp, fmt.Errorf("failed to get document count: %v", err)
	}

	fmt.Printf("Number of documents in collection '%s': %d\n", se.CollectionName, count)

	return importResp, nil
}

func (se *TypesenseSearchEngine) LoadData(filepath string) (*bytes.Buffer, error) {
	fmt.Println("Loading data...")

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("unable to open source file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var buf bytes.Buffer

	for scanner.Scan() {
		line := scanner.Bytes()

		var doc map[string]interface{}
		if err := json.Unmarshal(line, &doc); err != nil {
			continue
		}

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
			default:
				doc["id"] = fmt.Sprintf("%v", v)
			}
		} else {
			return nil, fmt.Errorf("missing 'id' field in document")
		}

		fixedLine, err := json.Marshal(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal fixed document: %v", err)
		}

		buf.Write(fixedLine)
		buf.WriteByte('\n')
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %v", err)
	}

	return &buf, nil
}

func (se *TypesenseSearchEngine) doRequest(method, url string, body interface{}) (string, error) {
	var buf *bytes.Buffer

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return "", err
		}
		buf = bytes.NewBuffer(data)
	} else {
		buf = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TYPESENSE-API-KEY", se.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

func (se *TypesenseSearchEngine) doStreamRequest(method, url string, documents *bytes.Buffer, contentType string) (string, error) {
	req, err := http.NewRequest(method, url, documents)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-TYPESENSE-API-KEY", se.ApiKey)
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}
