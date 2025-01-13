package models

import "github.com/google/uuid"

type Users struct {
	Id           uuid.UUID
	Login        string
	PasswordHash string
	IsAdmin      bool // can be some int or struct for more roles
}

type Images struct {
	Id        int64
	Path      string
	PathId    int64
	CreatorId uuid.UUID
}

type Avatars struct {
	Id      int64
	Path    string
	PathId  int64
	OwnerId uuid.UUID
}

type ImageWithUser struct {
	ImagePath string
	Login     string
	IsAdmin   bool
}
