package es

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/domain"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/rs/zerolog"
)

type Repo struct {
	client    *elasticsearch.Client
	userIndex string
	chatIndex string
	fileIndex string
	msgIndex  string
}

func NewRepo(addr string, log zerolog.Logger) (*Repo, error) {
	cfg := elasticsearch.Config{Addresses: []string{addr}}
	cl, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Err(err).Msg("couldn't create es client")
		return nil, err
	}
	return &Repo{
		client:    cl,
		userIndex: "users",
		chatIndex: "chats",
		fileIndex: "files",
		msgIndex:  "messages",
	}, nil
}

// --- Index operations ---

func (r *Repo) IndexMessage(ctx context.Context, m domain.Message) error {
	return r.index(ctx, r.msgIndex, m.ID, m)
}

func (r *Repo) IndexChat(ctx context.Context, c domain.Chat) error {
	return r.index(ctx, r.chatIndex, c.ID, c)
}

func (r *Repo) IndexUser(ctx context.Context, u domain.UserIndex) error {
	return r.index(ctx, r.userIndex, u.UUID, u)
}

func (r *Repo) IndexChatIndex(ctx context.Context, c domain.ChatIndex) error {
	return r.index(ctx, r.chatIndex, c.ID, c)
}

func (r *Repo) IndexFile(ctx context.Context, f domain.FileIndex) error {
	return r.index(ctx, r.fileIndex, f.ID, f)
}

func (r *Repo) index(ctx context.Context, index, id string, doc any) error {
	data, _ := json.Marshal(doc)
	req := esapi.IndexRequest{Index: index, DocumentID: id, Body: strings.NewReader(string(data))}
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("es index error: %s", res.String())
	}
	return nil
}

// --- Ensure indices with mappings ---
func (r *Repo) EnsureIndices(ctx context.Context) error {
	body := func(mapping string) *strings.Reader { return strings.NewReader(mapping) }

	usersMapping := `{
      "settings": { "index": { "number_of_shards": 1, "number_of_replicas": 0 } },
      "mappings": {
        "properties": {
          "id":        { "type": "keyword" },
          "username":  { "type": "text", "analyzer": "standard" },
          "email":     { "type": "text", "analyzer": "standard" },
          "about_me":  { "type": "text", "analyzer": "standard" },
          "friends":   { "type": "keyword" },
          "updated_at":{ "type": "date", "format": "epoch_second" }
        }
      }
    }`

	chatsMapping := `{
      "settings": { "index": { "number_of_shards": 1, "number_of_replicas": 0 } },
      "mappings": {
        "properties": {
          "id":         { "type": "keyword" },
          "title":      { "type": "text", "analyzer": "standard" },
          "member_ids": { "type": "keyword" },
          "kind":       { "type": "keyword" },
          "updated_at": { "type": "date", "format": "epoch_second" }
        }
      }
    }`

	filesMapping := `{
  "settings": { 
    "index": { 
      "number_of_shards": 1, 
      "number_of_replicas": 0 
    } 
  },
  "mappings": {
    "properties": {
      "id":         { "type": "keyword" },
      "filename":   { "type": "text", "analyzer": "standard", "fields": { "keyword": { "type": "keyword" } } },
      "url":        { "type": "text", "analyzer": "standard" },
      "mime":       { "type": "keyword" },
      "type":       { "type": "keyword" },
      "size_bytes": { "type": "long" },
      "author_id":  { "type": "keyword" },
      "chat_id":    { "type": "keyword" },
      "bucket":     { "type": "keyword" },
      "object_name": { "type": "keyword" },
      "content_type": { "type": "keyword" },
      "tags":       { "type": "keyword" },
      "created_at": { "type": "date", "format": "epoch_millis" },
      "updated_at": { "type": "date", "format": "epoch_second" }
    }
  }
}`

	messagesMapping := `{
      "settings": { "index": { "number_of_shards": 1, "number_of_replicas": 0 } },
      "mappings": {
        "properties": {
          "id":        { "type": "keyword" },
          "chat_id":   { "type": "keyword" },
          "author_id": { "type": "keyword" },
          "text":      { "type": "text", "analyzer": "standard" },
          "created_at":{ "type": "date", "format": "epoch_second" },
          "updated_at":{ "type": "date", "format": "epoch_second" },
          "deleted":   { "type": "boolean" }
        }
      }
    }`

	for idx, mapping := range map[string]string{
		r.userIndex: usersMapping,
		r.chatIndex: chatsMapping,
		r.fileIndex: filesMapping,
		r.msgIndex:  messagesMapping,
	} {
		req := esapi.IndicesCreateRequest{Index: idx, Body: body(mapping)}
		res, err := req.Do(ctx, r.client)
		if err != nil {
			return err
		}
		res.Body.Close()
	}
	return nil
}

// Search
func (r *Repo) Search(ctx context.Context, q domain.SearchQuery) ([]domain.SearchResult, error) {
	indexes := []string{r.userIndex, r.chatIndex, r.fileIndex, r.msgIndex}

	// Фильтрация по типу
	switch q.Type {
	case "user":
		indexes = []string{r.userIndex}
	case "chat":
		indexes = []string{r.chatIndex}
	case "file":
		indexes = []string{r.fileIndex}
	case "message":
		indexes = []string{r.msgIndex}
	case "media":
		indexes = []string{r.fileIndex}
	}

	// Настройки подсветки
	highlight := map[string]any{}
	if q.Highlight {
		highlight = map[string]any{
			"fields": map[string]any{
				"filename": map[string]any{"pre_tags": []string{"<b>"}, "post_tags": []string{"</b>"}},
				"text":     map[string]any{"pre_tags": []string{"<b>"}, "post_tags": []string{"</b>"}},
				"title":    map[string]any{"pre_tags": []string{"<b>"}, "post_tags": []string{"</b>"}},
				"username": map[string]any{"pre_tags": []string{"<b>"}, "post_tags": []string{"</b>"}},
				"about_me": map[string]any{"pre_tags": []string{"<b>"}, "post_tags": []string{"</b>"}},
			},
			"require_field_match": false,
		}
	}

	// Определяем поля для поиска в зависимости от типа
	searchFields := []string{"username^2", "email", "about_me", "title^2", "text^3", "filename^2", "url"}

	if q.Type == "file" || q.Type == "media" {
		searchFields = []string{"filename^3", "url^2", "tags"}
	} else if q.Type == "user" {
		searchFields = []string{"username^3", "email^2", "about_me"}
	} else if q.Type == "chat" {
		searchFields = []string{"title^3"}
	} else if q.Type == "message" {
		searchFields = []string{"text^3"}
	}

	body := map[string]any{
		"from": q.Offset,
		"size": q.Limit,
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":     q.Q,
				"fields":    searchFields,
				"fuzziness": "AUTO",
				"operator":  "or",
			},
		},
		"highlight": highlight,
		"sort": []map[string]any{
			{"_score": map[string]any{"order": "desc"}},
			{"updated_at": map[string]any{"order": "desc"}},
		},
	}

	data, _ := json.Marshal(body)
	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(indexes...),
		r.client.Search.WithBody(strings.NewReader(string(data))),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var raw map[string]any
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, err
	}

	hits := (((raw["hits"]).(map[string]any))["hits"]).([]any)
	out := make([]domain.SearchResult, 0, len(hits))

	for _, h := range hits {
		hi := h.(map[string]any)
		_src := hi["_source"].(map[string]any)
		index := hi["_index"].(string)

		// Определяем тип по индексу
		var t string
		switch index {
		case r.userIndex:
			t = "user"
		case r.chatIndex:
			t = "chat"
		case r.fileIndex:
			t = "file"
		case r.msgIndex:
			t = "message"
		}

		// Извлекаем данные
		id := fmt.Sprintf("%v", _src["id"])
		snippet := ""

		// Обработка подсветки
		if q.Highlight {
			if hl, ok := hi["highlight"].(map[string]any); ok {
				for _, v := range hl {
					arr := v.([]any)
					if len(arr) > 0 {
						snippet = arr[0].(string)
						break
					}
				}
			}
		}

		// Формируем результат в зависимости от типа
		result := domain.SearchResult{
			ID:      id,
			Type:    t,
			Snippet: snippet,
		}

		switch t {
		case "user":
			result.Title = fmt.Sprintf("%v", _src["username"])
			result.ExtraIDs = []string{fmt.Sprintf("%v", _src["email"])}

		case "chat":
			result.Title = fmt.Sprintf("%v", _src["title"])
			if mem, ok := _src["member_ids"].([]any); ok {
				extra := make([]string, 0, len(mem))
				for _, m := range mem {
					extra = append(extra, fmt.Sprintf("%v", m))
				}
				result.ExtraIDs = extra
			}

		case "file":
			result.Title = fmt.Sprintf("%v", _src["filename"])
			extra := []string{}
			if chatID, ok := _src["chat_id"]; ok && chatID != "" {
				extra = append(extra, fmt.Sprintf("%v", chatID))
			}
			if authorID, ok := _src["author_id"]; ok && authorID != "" {
				extra = append(extra, fmt.Sprintf("%v", authorID))
			}
			if mime, ok := _src["mime"]; ok && mime != "" {
				extra = append(extra, fmt.Sprintf("%v", mime))
			}
			result.ExtraIDs = extra

		case "message":
			result.Title = fmt.Sprintf("%v", _src["text"])
			extra := []string{}
			if chatID, ok := _src["chat_id"]; ok && chatID != "" {
				extra = append(extra, fmt.Sprintf("%v", chatID))
			}
			if authorID, ok := _src["author_id"]; ok && authorID != "" {
				extra = append(extra, fmt.Sprintf("%v", authorID))
			}
			result.ExtraIDs = extra
		}

		out = append(out, result)
	}

	return out, nil
}
func (r *Repo) DeleteFile(ctx context.Context, id string) error {
	req := esapi.DeleteRequest{
		Index:      r.fileIndex,
		DocumentID: id,
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("es delete error: %s", res.String())
	}

	return nil
}
