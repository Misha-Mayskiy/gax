// api_gateway/internal/service/room_service_test.go
package test

import (
	"context"
	"testing"

	"api_gateway/internal/service"
	pb "api_gateway/pkg/api/room"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// MockRoomServiceClient - мок для pb.RoomServiceClient
type MockRoomServiceClient struct {
	mock.Mock
}

func (m *MockRoomServiceClient) CreateRoom(ctx context.Context, in *pb.CreateRoomRequest, opts ...grpc.CallOption) (*pb.RoomResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.RoomResponse), args.Error(1)
}

func (m *MockRoomServiceClient) JoinRoom(ctx context.Context, in *pb.JoinRoomRequest, opts ...grpc.CallOption) (*pb.RoomResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.RoomResponse), args.Error(1)
}

func (m *MockRoomServiceClient) SetPlayback(ctx context.Context, in *pb.SetPlaybackRequest, opts ...grpc.CallOption) (*pb.RoomResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.RoomResponse), args.Error(1)
}

func (m *MockRoomServiceClient) GetState(ctx context.Context, in *pb.GetStateRequest, opts ...grpc.CallOption) (*pb.RoomResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.RoomResponse), args.Error(1)
}

func TestRoomService_CreateRoom(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockRoomServiceClient)
		req            *pb.CreateRoomRequest
		expectedResult *pb.RoomResponse
		expectedError  bool
	}{
		{
			name: "Успешное создание комнаты",
			req: &pb.CreateRoomRequest{
				HostId:   "user-123",
				TrackUrl: "https://example.com/track.mp3",
			},
			setupMock: func(mockClient *MockRoomServiceClient) {
				expectedResponse := &pb.RoomResponse{
					RoomId:    "room-abc123",
					HostId:    "user-123",
					TrackUrl:  "https://example.com/track.mp3",
					Timestamp: 1234567890,
					Status:    "play",
					Users:     []string{"user-123"},
				}
				mockClient.On("CreateRoom", mock.Anything, mock.Anything, mock.Anything).
					Return(expectedResponse, nil)
			},
			expectedResult: &pb.RoomResponse{
				RoomId:    "room-abc123",
				HostId:    "user-123",
				TrackUrl:  "https://example.com/track.mp3",
				Timestamp: 1234567890,
				Status:    "play",
				Users:     []string{"user-123"},
			},
			expectedError: false,
		},
		// ... остальные тестовые случаи
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockClient *MockRoomServiceClient
			if tt.setupMock != nil {
				mockClient = new(MockRoomServiceClient)
				tt.setupMock(mockClient)
			}

			// Используем экспортированный конструктор
			service := service.NewTestRoomService(mockClient)

			result, err := service.CreateRoom(context.Background(), tt.req)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			if mockClient != nil {
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestRoomService_JoinRoom(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockRoomServiceClient)
		req            *pb.JoinRoomRequest
		expectedResult *pb.RoomResponse
		expectedError  bool
	}{
		{
			name: "Успешное присоединение к комнате",
			req: &pb.JoinRoomRequest{
				RoomId: "room-abc123",
				UserId: "user-456",
			},
			setupMock: func(mockClient *MockRoomServiceClient) {
				expectedResponse := &pb.RoomResponse{
					RoomId:    "room-abc123",
					HostId:    "user-123",
					TrackUrl:  "https://example.com/track.mp3",
					Timestamp: 1234567890,
					Status:    "play",
					Users:     []string{"user-123", "user-456"},
				}
				mockClient.On("JoinRoom", mock.Anything, mock.Anything, mock.Anything).
					Return(expectedResponse, nil)
			},
			expectedResult: &pb.RoomResponse{
				RoomId:    "room-abc123",
				HostId:    "user-123",
				TrackUrl:  "https://example.com/track.mp3",
				Timestamp: 1234567890,
				Status:    "play",
				Users:     []string{"user-123", "user-456"},
			},
			expectedError: false,
		},
		{
			name: "Комната не найдена",
			req: &pb.JoinRoomRequest{
				RoomId: "non-existent-room",
				UserId: "user-456",
			},
			setupMock: func(mockClient *MockRoomServiceClient) {
				mockClient.On("JoinRoom", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRoomServiceClient)
			tt.setupMock(mockClient)

			service := service.NewTestRoomService(mockClient)

			result, err := service.JoinRoom(context.Background(), tt.req)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRoomService_SetPlayback(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockRoomServiceClient)
		req            *pb.SetPlaybackRequest
		expectedResult *pb.RoomResponse
		expectedError  bool
	}{
		{
			name: "Успешная установка воспроизведения - play",
			req: &pb.SetPlaybackRequest{
				RoomId:    "room-abc123",
				Action:    "play",
				Timestamp: 1234567890,
			},
			setupMock: func(mockClient *MockRoomServiceClient) {
				expectedResponse := &pb.RoomResponse{
					RoomId:    "room-abc123",
					HostId:    "user-123",
					TrackUrl:  "https://example.com/track.mp3",
					Timestamp: 1234567890,
					Status:    "play",
					Users:     []string{"user-123", "user-456"},
				}
				mockClient.On("SetPlayback", mock.Anything, mock.Anything, mock.Anything).
					Return(expectedResponse, nil)
			},
			expectedResult: &pb.RoomResponse{
				RoomId:    "room-abc123",
				HostId:    "user-123",
				TrackUrl:  "https://example.com/track.mp3",
				Timestamp: 1234567890,
				Status:    "play",
				Users:     []string{"user-123", "user-456"},
			},
			expectedError: false,
		},
		{
			name: "Успешная установка воспроизведения - pause",
			req: &pb.SetPlaybackRequest{
				RoomId:    "room-abc123",
				Action:    "pause",
				Timestamp: 1234567890,
			},
			setupMock: func(mockClient *MockRoomServiceClient) {
				expectedResponse := &pb.RoomResponse{
					RoomId:    "room-abc123",
					HostId:    "user-123",
					TrackUrl:  "https://example.com/track.mp3",
					Timestamp: 1234567890,
					Status:    "pause",
					Users:     []string{"user-123", "user-456"},
				}
				mockClient.On("SetPlayback", mock.Anything, mock.Anything, mock.Anything).
					Return(expectedResponse, nil)
			},
			expectedResult: &pb.RoomResponse{
				RoomId:    "room-abc123",
				HostId:    "user-123",
				TrackUrl:  "https://example.com/track.mp3",
				Timestamp: 1234567890,
				Status:    "pause",
				Users:     []string{"user-123", "user-456"},
			},
			expectedError: false,
		},
		{
			name: "Ошибка при установке воспроизведения",
			req: &pb.SetPlaybackRequest{
				RoomId:    "room-abc123",
				Action:    "invalid-action",
				Timestamp: 1234567890,
			},
			setupMock: func(mockClient *MockRoomServiceClient) {
				mockClient.On("SetPlayback", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRoomServiceClient)
			tt.setupMock(mockClient)

			service := service.NewTestRoomService(mockClient)

			result, err := service.SetPlayback(context.Background(), tt.req)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRoomService_GetState(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockRoomServiceClient)
		req            *pb.GetStateRequest
		expectedResult *pb.RoomResponse
		expectedError  bool
	}{
		{
			name: "Успешное получение состояния комнаты",
			req: &pb.GetStateRequest{
				RoomId: "room-abc123",
			},
			setupMock: func(mockClient *MockRoomServiceClient) {
				expectedResponse := &pb.RoomResponse{
					RoomId:    "room-abc123",
					HostId:    "user-123",
					TrackUrl:  "https://example.com/track.mp3",
					Timestamp: 1234567890,
					Status:    "play",
					Users:     []string{"user-123", "user-456"},
				}
				mockClient.On("GetState", mock.Anything, mock.Anything, mock.Anything).
					Return(expectedResponse, nil)
			},
			expectedResult: &pb.RoomResponse{
				RoomId:    "room-abc123",
				HostId:    "user-123",
				TrackUrl:  "https://example.com/track.mp3",
				Timestamp: 1234567890,
				Status:    "play",
				Users:     []string{"user-123", "user-456"},
			},
			expectedError: false,
		},
		{
			name: "Комната не найдена",
			req: &pb.GetStateRequest{
				RoomId: "non-existent-room",
			},
			setupMock: func(mockClient *MockRoomServiceClient) {
				mockClient.On("GetState", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRoomServiceClient)
			tt.setupMock(mockClient)

			service := service.NewTestRoomService(mockClient)

			result, err := service.GetState(context.Background(), tt.req)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRoomService_UserHasAccess(t *testing.T) {
	tests := []struct {
		name           string
		roomID         string
		userID         string
		expectedResult bool
	}{
		{
			name:           "Пользователь имеет доступ",
			roomID:         "room-123",
			userID:         "user-456",
			expectedResult: true,
		},
		{
			name:           "Другой пользователь имеет доступ",
			roomID:         "room-abc",
			userID:         "user-xyz",
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := service.NewTestRoomService(nil)

			result, err := service.UserHasAccess(tt.roomID, tt.userID)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestRoomService_GetRoomState(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockRoomServiceClient)
		roomID         string
		expectedResult *pb.RoomResponse
		expectedError  bool
	}{
		{
			name:   "Успешное получение состояния комнаты по ID",
			roomID: "room-abc123",
			setupMock: func(mockClient *MockRoomServiceClient) {
				expectedResponse := &pb.RoomResponse{
					RoomId:    "room-abc123",
					HostId:    "user-123",
					TrackUrl:  "https://example.com/track.mp3",
					Timestamp: 1234567890,
					Status:    "play",
					Users:     []string{"user-123", "user-456"},
				}
				mockClient.On("GetState", mock.Anything, mock.MatchedBy(func(req *pb.GetStateRequest) bool {
					return req.RoomId == "room-abc123"
				}), mock.Anything).
					Return(expectedResponse, nil)
			},
			expectedResult: &pb.RoomResponse{
				RoomId:    "room-abc123",
				HostId:    "user-123",
				TrackUrl:  "https://example.com/track.mp3",
				Timestamp: 1234567890,
				Status:    "play",
				Users:     []string{"user-123", "user-456"},
			},
			expectedError: false,
		},
		{
			name:   "Ошибка при получении состояния",
			roomID: "non-existent-room",
			setupMock: func(mockClient *MockRoomServiceClient) {
				mockClient.On("GetState", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRoomServiceClient)
			tt.setupMock(mockClient)

			service := service.NewTestRoomService(mockClient)

			result, err := service.GetRoomState(tt.roomID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRoomService_UpdatePlaybackState(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockRoomServiceClient)
		roomID        string
		action        string
		position      float64
		volume        float64
		expectedError bool
	}{
		{
			name:     "Успешное обновление состояния воспроизведения - play",
			roomID:   "room-abc123",
			action:   "play",
			position: 123.45,
			volume:   0.8,
			setupMock: func(mockClient *MockRoomServiceClient) {
				expectedResponse := &pb.RoomResponse{
					RoomId:    "room-abc123",
					HostId:    "user-123",
					TrackUrl:  "https://example.com/track.mp3",
					Timestamp: 123, // position конвертируется в int64
					Status:    "play",
					Users:     []string{"user-123", "user-456"},
				}
				mockClient.On("SetPlayback", mock.Anything, mock.MatchedBy(func(req *pb.SetPlaybackRequest) bool {
					return req.RoomId == "room-abc123" &&
						req.Action == "play" &&
						req.Timestamp == 123 // Проверяем конвертацию
				}), mock.Anything).
					Return(expectedResponse, nil)
			},
			expectedError: false,
		},
		{
			name:     "Успешное обновление состояния воспроизведения - pause",
			roomID:   "room-abc123",
			action:   "pause",
			position: 456.78,
			volume:   0.5,
			setupMock: func(mockClient *MockRoomServiceClient) {
				expectedResponse := &pb.RoomResponse{
					RoomId:    "room-abc123",
					HostId:    "user-123",
					TrackUrl:  "https://example.com/track.mp3",
					Timestamp: 456, // position конвертируется в int64
					Status:    "pause",
					Users:     []string{"user-123", "user-456"},
				}
				mockClient.On("SetPlayback", mock.Anything, mock.MatchedBy(func(req *pb.SetPlaybackRequest) bool {
					return req.RoomId == "room-abc123" &&
						req.Action == "pause" &&
						req.Timestamp == 456
				}), mock.Anything).
					Return(expectedResponse, nil)
			},
			expectedError: false,
		},
		{
			name:     "Ошибка при обновлении состояния",
			roomID:   "room-abc123",
			action:   "invalid-action",
			position: 0,
			volume:   0,
			setupMock: func(mockClient *MockRoomServiceClient) {
				mockClient.On("SetPlayback", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRoomServiceClient)
			tt.setupMock(mockClient)

			service := service.NewTestRoomService(mockClient)

			err := service.UpdatePlaybackState(tt.roomID, tt.action, tt.position, tt.volume)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRoomService_EdgeCases(t *testing.T) {
	t.Run("Нулевой клиент в методе CreateRoom", func(t *testing.T) {
		service := service.NewTestRoomService(nil)

		req := &pb.CreateRoomRequest{
			HostId:   "user-123",
			TrackUrl: "https://example.com/track.mp3",
		}

		result, err := service.CreateRoom(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "room service not available")
	})

	t.Run("Нулевой клиент в методе GetState", func(t *testing.T) {
		service := service.NewTestRoomService(nil)

		req := &pb.GetStateRequest{
			RoomId: "room-123",
		}

		result, err := service.GetState(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "room service not available")
	})

	t.Run("Нулевой клиент в методе JoinRoom", func(t *testing.T) {
		service := service.NewTestRoomService(nil)

		req := &pb.JoinRoomRequest{
			RoomId: "room-123",
			UserId: "user-456",
		}

		result, err := service.JoinRoom(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "room service not available")
	})

	t.Run("Нулевой клиент в методе SetPlayback", func(t *testing.T) {
		service := service.NewTestRoomService(nil)

		req := &pb.SetPlaybackRequest{
			RoomId:    "room-123",
			Action:    "play",
			Timestamp: 1234567890,
		}

		result, err := service.SetPlayback(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "room service not available")
	})

	t.Run("Пустые параметры в UserHasAccess", func(t *testing.T) {
		service := service.NewTestRoomService(nil)

		// Тестируем с пустыми строками
		result, err := service.UserHasAccess("", "")

		assert.NoError(t, err)
		assert.True(t, result) // Метод всегда возвращает true
	})
}

func TestRoomService_ErrorHandling(t *testing.T) {
	t.Run("Контекст с таймаутом", func(t *testing.T) {
		mockClient := new(MockRoomServiceClient)

		// Настраиваем мок для возвращения ошибки таймаута
		mockClient.On("CreateRoom", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, context.DeadlineExceeded)

		service := service.NewTestRoomService(mockClient)

		req := &pb.CreateRoomRequest{
			HostId:   "user-123",
			TrackUrl: "https://example.com/track.mp3",
		}

		// Создаем контекст с очень коротким таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 1)
		defer cancel()

		result, err := service.CreateRoom(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, context.DeadlineExceeded, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("Неизвестное действие в UpdatePlaybackState", func(t *testing.T) {
		mockClient := new(MockRoomServiceClient)

		// Мок будет возвращать ошибку для неизвестного действия
		mockClient.On("SetPlayback", mock.Anything, mock.MatchedBy(func(req *pb.SetPlaybackRequest) bool {
			return req.Action == "unknown-action"
		}), mock.Anything).
			Return(nil, assert.AnError)

		service := service.NewTestRoomService(mockClient)

		err := service.UpdatePlaybackState("room-123", "unknown-action", 0, 0)

		assert.Error(t, err)

		mockClient.AssertExpectations(t)
	})
}
