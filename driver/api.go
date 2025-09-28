package driver

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"challenge.haraj.com.sa/kraicklist/core"
)

type Api struct {
	ApiConfig
}

type ApiConfig struct {
	Ctx     context.Context
	Service core.Service `validate:"nonnil"`
}

func NewApi(cfg ApiConfig) (*Api, error) {
	return &Api{ApiConfig: cfg}, nil
}

func (a *Api) GetHandler() http.Handler {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	mux.HandleFunc("/api/v1/search", a.handleGetSearch)

	return mux
}

func (a *Api) handleGetSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if len(q) == 0 {
		q = "*"
	}

	perPage := 9
	perPageQuery := r.URL.Query().Get("perPage")
	if perPageQuery != "" {
		perPage, _ = strconv.Atoi(perPageQuery)
	}

	page := 1
	pageQuery := r.URL.Query().Get("page")
	if pageQuery != "" {
		page, _ = strconv.Atoi(pageQuery)
	}

	resp, err := a.Service.GetSearch(a.Ctx, q, perPage, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respObj := SearchResponse{
		Ok:   true,
		Data: resp,
		Ts:   time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(respObj); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

}
