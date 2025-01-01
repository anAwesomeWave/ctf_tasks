package storage

import (
	"accessCtf/internal/storage/models"
	"github.com/google/uuid"
)

type Storage interface {
	CreateUser(login, password string) (uuid.UUID, error)
	GetUserByUUID(id uuid.UUID) (*models.Users, error)
	GetUserByLoginPassword(login, password string) (*models.Users, error)
	CreateImage(creator *models.Users, path string) (*models.Images, error)
	CreateAvatar(creator *models.Users, path string) (*models.Avatars, error)
}
