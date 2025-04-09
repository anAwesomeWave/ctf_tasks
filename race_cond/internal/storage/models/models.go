package models

import (
	"github.com/google/uuid"
	"time"
)

type Users struct {
	Id       uuid.UUID
	Login    string
	Password string
	IsAdmin  bool // can be some int or struct for more roles
}

type Images struct {
	Id           int64
	Path         string
	PathId       int64
	CreatorId    uuid.UUID
	CreationTime time.Time
}

type Avatars struct {
	Id           int64
	Path         string
	PathId       int64
	OwnerId      uuid.UUID
	CreationTime time.Time
}

type ImageWithUser struct {
	ImagePath  string
	AvatarPath string
	Login      string
	IsAdmin    bool
}
