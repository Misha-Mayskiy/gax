package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"api_gateway/pkg/logger"

	"go.uber.org/zap"
)

type SearchService struct {
	addr   string
	client *http.Client
}

func NewSearchService(addr string) *SearchService {
	return &SearchService{
		addr: addr,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// NewSearchServiceForTest создает SearchService для тестов
func NewSearchServiceForTest(addr string, client *http.Client) *SearchService {
	if client == nil {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	return &SearchService{
		addr:   addr,
		client: client,
	}
}

// Экспортируйте структуру для тестов
type ExportSearchServiceForTest struct {
	BaseURL    string
	HTTPClient *http.Client
}

func (s *SearchService) Search(ctx context.Context, q, t string, limit, offset int, highlight bool) map[string]any {
	url := fmt.Sprintf("%s/search?q=%s&type=%s&limit=%d&offset=%d&highlight=%t",
		s.addr, q, t, limit, offset, highlight)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "SearchService request failed", zap.Error(err))
		return nil
	}
	defer resp.Body.Close()

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "SearchService decode failed", zap.Error(err))
		return nil
	}
	return data
}
