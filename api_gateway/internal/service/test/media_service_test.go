// api_gateway/internal/service/test/media_service_test.go
package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"api_gateway/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockHTTPRoundTripper - мок для http.RoundTripper
type MockHTTPRoundTripper struct {
	mock.Mock
}

func (m *MockHTTPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// MockReadCloser - мок для io.ReadCloser
type MockReadCloser struct {
	io.Reader
}

func (m *MockReadCloser) Close() error {
	return nil
}

func TestMediaService_DownloadFile(t *testing.T) {
	tests := []struct {
		name           string
		fileID         string
		setupMock      func(*MockHTTPRoundTripper)
		expectedBody   string
		expectedHeader string
		expectedError  bool
	}{
		{
			name:   "Успешная загрузка файла",
			fileID: "file-123",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       &MockReadCloser{Reader: strings.NewReader("file content here")},
					Header:     make(http.Header),
				}
				resp.Header.Set("Content-Disposition", "attachment; filename=\"test.txt\"")
				resp.Header.Set("Content-Type", "text/plain")

				mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodGet &&
						strings.Contains(req.URL.String(), "/download/file-123")
				})).Return(resp, nil)
			},
			expectedBody:   "file content here",
			expectedHeader: "attachment; filename=\"test.txt\"",
			expectedError:  false,
		},
		{
			name:   "Файл не найден",
			fileID: "not-found",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				resp := &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       &MockReadCloser{Reader: strings.NewReader("File not found")},
					Header:     make(http.Header),
				}

				mockRT.On("RoundTrip", mock.Anything).Return(resp, nil)
			},
			expectedBody:   "",
			expectedHeader: "",
			expectedError:  true,
		},
		{
			name:   "Ошибка сети",
			fileID: "file-123",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				mockRT.On("RoundTrip", mock.Anything).Return(nil, fmt.Errorf("network error"))
			},
			expectedBody:   "",
			expectedHeader: "",
			expectedError:  true,
		},
		{
			name:   "Успешная загрузка без заголовка Content-Disposition",
			fileID: "file-no-header",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       &MockReadCloser{Reader: strings.NewReader("content without header")},
					Header:     make(http.Header),
				}
				resp.Header.Set("Content-Type", "text/plain")

				mockRT.On("RoundTrip", mock.Anything).Return(resp, nil)
			},
			expectedBody:   "content without header",
			expectedHeader: "",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRT := new(MockHTTPRoundTripper)
			tt.setupMock(mockRT)

			httpClient := &http.Client{
				Transport: mockRT,
			}

			service := service.NewMediaServiceForTest("http://media-service:8080", httpClient)

			reader, filename, err := service.DownloadFile(context.Background(), tt.fileID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, reader)
				assert.Empty(t, filename)
			} else {
				require.NoError(t, err)
				require.NotNil(t, reader)
				defer reader.Close()

				content, err := io.ReadAll(reader)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, string(content))
				assert.Equal(t, tt.expectedHeader, filename)
			}

			mockRT.AssertExpectations(t)
		})
	}
}

func TestMediaService_GetFileMeta(t *testing.T) {
	tests := []struct {
		name           string
		fileID         string
		setupMock      func(*MockHTTPRoundTripper)
		expectedResult map[string]any
		expectedError  bool
	}{
		{
			name:   "Успешное получение метаданных",
			fileID: "file-123",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				expectedResult := map[string]any{
					"id":          "file-123",
					"filename":    "test.txt",
					"size":        1024,
					"mime_type":   "text/plain",
					"uploaded_at": "2024-01-01T12:00:00Z",
					"uploaded_by": "user-123",
				}
				resultJSON, _ := json.Marshal(expectedResult)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       &MockReadCloser{Reader: bytes.NewReader(resultJSON)},
					Header:     make(http.Header),
				}
				resp.Header.Set("Content-Type", "application/json")

				mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodGet &&
						strings.Contains(req.URL.String(), "/media/file-123")
				})).Return(resp, nil)
			},
			expectedResult: map[string]any{
				"id":          "file-123",
				"filename":    "test.txt",
				"size":        float64(1024),
				"mime_type":   "text/plain",
				"uploaded_at": "2024-01-01T12:00:00Z",
				"uploaded_by": "user-123",
			},
			expectedError: false,
		},
		{
			name:   "Файл не найден",
			fileID: "not-found",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				resp := &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       &MockReadCloser{Reader: strings.NewReader(`{"error":"file not found"}`)},
					Header:     make(http.Header),
				}
				resp.Header.Set("Content-Type", "application/json")

				mockRT.On("RoundTrip", mock.Anything).Return(resp, nil)
			},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:   "Некорректный JSON в ответе",
			fileID: "file-123",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       &MockReadCloser{Reader: strings.NewReader("invalid json")},
					Header:     make(http.Header),
				}
				resp.Header.Set("Content-Type", "application/json")

				mockRT.On("RoundTrip", mock.Anything).Return(resp, nil)
			},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:   "Ошибка сети",
			fileID: "file-123",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				mockRT.On("RoundTrip", mock.Anything).Return(nil, fmt.Errorf("network error"))
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRT := new(MockHTTPRoundTripper)
			tt.setupMock(mockRT)

			httpClient := &http.Client{
				Transport: mockRT,
			}

			service := service.NewMediaServiceForTest("http://media-service:8080", httpClient)

			result, err := service.GetFileMeta(context.Background(), tt.fileID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRT.AssertExpectations(t)
		})
	}
}

func TestMediaService_DeleteFile(t *testing.T) {
	tests := []struct {
		name          string
		fileID        string
		userID        string
		setupMock     func(*MockHTTPRoundTripper)
		expectedError bool
	}{
		{
			name:   "Успешное удаление файла с userID",
			fileID: "file-123",
			userID: "user-123",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       &MockReadCloser{Reader: strings.NewReader(`{"success":true}`)},
					Header:     make(http.Header),
				}

				mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodDelete &&
						strings.Contains(req.URL.String(), "/media/delete/file-123") &&
						req.URL.Query().Get("user_id") == "user-123"
				})).Return(resp, nil)
			},
			expectedError: false,
		},
		{
			name:   "Успешное удаление файла без userID",
			fileID: "file-123",
			userID: "",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       &MockReadCloser{Reader: strings.NewReader(`{"success":true}`)},
					Header:     make(http.Header),
				}

				mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodDelete &&
						strings.Contains(req.URL.String(), "/media/delete/file-123") &&
						req.URL.Query().Get("user_id") == ""
				})).Return(resp, nil)
			},
			expectedError: false,
		},
		{
			name:   "Ошибка доступа",
			fileID: "file-123",
			userID: "user-456",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				resp := &http.Response{
					StatusCode: http.StatusForbidden,
					Body:       &MockReadCloser{Reader: strings.NewReader(`{"error":"access denied"}`)},
					Header:     make(http.Header),
				}

				mockRT.On("RoundTrip", mock.Anything).Return(resp, nil)
			},
			expectedError: true,
		},
		{
			name:   "Файл не найден при удалении",
			fileID: "not-found",
			userID: "user-123",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				resp := &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       &MockReadCloser{Reader: strings.NewReader(`{"error":"file not found"}`)},
					Header:     make(http.Header),
				}

				mockRT.On("RoundTrip", mock.Anything).Return(resp, nil)
			},
			expectedError: true,
		},
		{
			name:   "Ошибка сети",
			fileID: "file-123",
			userID: "user-123",
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				mockRT.On("RoundTrip", mock.Anything).Return(nil, fmt.Errorf("network error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRT := new(MockHTTPRoundTripper)
			tt.setupMock(mockRT)

			httpClient := &http.Client{
				Transport: mockRT,
			}

			service := service.NewMediaServiceForTest("http://media-service:8080", httpClient)

			err := service.DeleteFile(context.Background(), tt.fileID, tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRT.AssertExpectations(t)
		})
	}
}

func TestMediaService_ListUserFiles(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		limit          int
		offset         int
		setupMock      func(*MockHTTPRoundTripper)
		expectedResult map[string]any
		expectedError  bool
	}{
		{
			name:   "Успешное получение списка файлов",
			userID: "user-123",
			limit:  10,
			offset: 0,
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				expectedResult := map[string]any{
					"files": []any{
						map[string]any{"id": "file-1", "filename": "test1.txt", "size": 1024},
						map[string]any{"id": "file-2", "filename": "test2.jpg", "size": 2048},
					},
					"total": 2,
				}
				resultJSON, _ := json.Marshal(expectedResult)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       &MockReadCloser{Reader: bytes.NewReader(resultJSON)},
					Header:     make(http.Header),
				}
				resp.Header.Set("Content-Type", "application/json")

				mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodGet &&
						strings.Contains(req.URL.String(), "/media/list") &&
						strings.Contains(req.URL.String(), "user_id=user-123") &&
						strings.Contains(req.URL.String(), "limit=10") &&
						strings.Contains(req.URL.String(), "offset=0")
				})).Return(resp, nil)
			},
			expectedResult: map[string]any{
				"files": []any{
					map[string]any{"id": "file-1", "filename": "test1.txt", "size": float64(1024)},
					map[string]any{"id": "file-2", "filename": "test2.jpg", "size": float64(2048)},
				},
				"total": float64(2),
			},
			expectedError: false,
		},
		{
			name:   "Пустой список файлов",
			userID: "user-456",
			limit:  10,
			offset: 0,
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				expectedResult := map[string]any{
					"files": []any{},
					"total": 0,
				}
				resultJSON, _ := json.Marshal(expectedResult)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       &MockReadCloser{Reader: bytes.NewReader(resultJSON)},
					Header:     make(http.Header),
				}
				resp.Header.Set("Content-Type", "application/json")

				mockRT.On("RoundTrip", mock.Anything).Return(resp, nil)
			},
			expectedResult: map[string]any{
				"files": []any{},
				"total": float64(0),
			},
			expectedError: false,
		},
		{
			name:   "Ошибка сервера",
			userID: "user-123",
			limit:  10,
			offset: 0,
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       &MockReadCloser{Reader: strings.NewReader("Server Error")},
					Header:     make(http.Header),
				}

				mockRT.On("RoundTrip", mock.Anything).Return(resp, nil)
			},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:   "Ошибка сети",
			userID: "user-123",
			limit:  10,
			offset: 0,
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				mockRT.On("RoundTrip", mock.Anything).Return(nil, fmt.Errorf("network error"))
			},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:   "Некорректный JSON в ответе",
			userID: "user-123",
			limit:  10,
			offset: 0,
			setupMock: func(mockRT *MockHTTPRoundTripper) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       &MockReadCloser{Reader: strings.NewReader("invalid json")},
					Header:     make(http.Header),
				}
				resp.Header.Set("Content-Type", "application/json")

				mockRT.On("RoundTrip", mock.Anything).Return(resp, nil)
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRT := new(MockHTTPRoundTripper)
			tt.setupMock(mockRT)

			httpClient := &http.Client{
				Transport: mockRT,
			}

			service := service.NewMediaServiceForTest("http://media-service:8080", httpClient)

			result, err := service.ListUserFiles(context.Background(), tt.userID, tt.limit, tt.offset)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRT.AssertExpectations(t)
		})
	}
}

func TestMediaService_EdgeCases(t *testing.T) {
	t.Run("Загрузка пустого файла", func(t *testing.T) {
		mockRT := new(MockHTTPRoundTripper)

		expectedResult := map[string]any{
			"id":       "empty-file",
			"filename": "empty.txt",
			"size":     0,
		}
		resultJSON, _ := json.Marshal(expectedResult)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       &MockReadCloser{Reader: bytes.NewReader(resultJSON)},
			Header:     make(http.Header),
		}
		resp.Header.Set("Content-Type", "application/json")

		mockRT.On("RoundTrip", mock.Anything).Return(resp, nil)

		httpClient := &http.Client{
			Transport: mockRT,
		}

		service := service.NewMediaServiceForTest("http://media-service:8080", httpClient)

		// Пустой файл
		file := strings.NewReader("")
		result, err := service.UploadFile(context.Background(), file, "empty.txt", "text/plain", "user-123", "chat-456")

		assert.NoError(t, err)
		assert.Equal(t, expectedResult["id"], result["id"])
		assert.Equal(t, expectedResult["filename"], result["filename"])
		assert.Equal(t, float64(0), result["size"])

		mockRT.AssertExpectations(t)
	})

	t.Run("Загрузка с очень длинным именем файла", func(t *testing.T) {
		mockRT := new(MockHTTPRoundTripper)

		longFilename := strings.Repeat("a", 255) + ".txt"
		expectedResult := map[string]any{
			"id":       "long-file",
			"filename": longFilename,
			"size":     100,
		}
		resultJSON, _ := json.Marshal(expectedResult)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       &MockReadCloser{Reader: bytes.NewReader(resultJSON)},
			Header:     make(http.Header),
		}
		resp.Header.Set("Content-Type", "application/json")

		mockRT.On("RoundTrip", mock.Anything).Return(resp, nil)

		httpClient := &http.Client{
			Transport: mockRT,
		}

		service := service.NewMediaServiceForTest("http://media-service:8080", httpClient)

		file := strings.NewReader("test content")
		result, err := service.UploadFile(context.Background(), file, longFilename, "text/plain", "user-123", "chat-456")

		assert.NoError(t, err)
		assert.Equal(t, expectedResult["filename"], result["filename"])

		mockRT.AssertExpectations(t)
	})

	t.Run("Специальные символы в имени файла", func(t *testing.T) {
		mockRT := new(MockHTTPRoundTripper)

		filename := "тест-файл с пробелами & спец.символами.txt"
		expectedResult := map[string]any{
			"id":       "special-file",
			"filename": filename,
			"size":     50,
		}
		resultJSON, _ := json.Marshal(expectedResult)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       &MockReadCloser{Reader: bytes.NewReader(resultJSON)},
			Header:     make(http.Header),
		}
		resp.Header.Set("Content-Type", "application/json")

		mockRT.On("RoundTrip", mock.Anything).Return(resp, nil)

		httpClient := &http.Client{
			Transport: mockRT,
		}

		service := service.NewMediaServiceForTest("http://media-service:8080", httpClient)

		file := strings.NewReader("test content with special chars")
		result, err := service.UploadFile(context.Background(), file, filename, "text/plain", "user-123", "chat-456")

		assert.NoError(t, err)
		assert.Equal(t, expectedResult["filename"], result["filename"])

		mockRT.AssertExpectations(t)
	})
}

// func TestMediaService_ContextCancellation(t *testing.T) {
// 	t.Run("Контекст отменен перед запросом", func(t *testing.T) {
// 		mockRT := new(MockHTTPRoundTripper)

// 		httpClient := &http.Client{
// 			Transport: mockRT,
// 		}

// 		service := service.NewMediaServiceForTest("http://media-service:8080", httpClient)

// 		ctx, cancel := context.WithCancel(context.Background())
// 		cancel() // Отменяем контекст сразу

// 		file := strings.NewReader("test")
// 		result, err := service.UploadFile(ctx, file, "test.txt", "text/plain", "user-123", "chat-456")

// 		// Запрос не должен был быть сделан из-за отмененного контекста
// 		assert.Error(t, err)
// 		assert.Nil(t, result)
// 		assert.Contains(t, err.Error(), "context")
// 	})
// }
