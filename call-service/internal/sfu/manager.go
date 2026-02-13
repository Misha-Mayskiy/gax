package sfu

import (
	"call-service/internal/repository"
	"sync"
)

type RoomManager struct {
	Lock  sync.RWMutex
	Rooms map[string]*Room
	Repo  *repository.RedisRepo
}

func NewRoomManager(repo *repository.RedisRepo) *RoomManager {
	return &RoomManager{
		Rooms: make(map[string]*Room),
		Repo:  repo,
	}
}

func (m *RoomManager) GetOrCreateRoom(roomID string) *Room {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	if room, exists := m.Rooms[roomID]; exists {
		return room
	}

	newRoom := NewRoom(roomID)
	m.Rooms[roomID] = newRoom
	return newRoom
}
