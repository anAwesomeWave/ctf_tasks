package models

import "github.com/google/uuid"

type Users struct {
	Id           uuid.UUID
	Login        string
	PasswordHash string
	IsAdmin      bool // can be some int or struct for more roles
	AvatarPath   string
}

type Images struct {
	Id        int64
	Path      string
	CreatorId uuid.UUID
}

type Avatars struct {
	Id      int64
	Path    string
	OwnerId uuid.UUID
}
