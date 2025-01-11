package storage

import (
	"accessCtf/internal/config"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

type Storage interface {
	CreateUser(login, password string) (uuid.UUID, error)
	//GetUserByUUID(id uuid.UUID) (*models.Users, error)
	//GetUserByLoginPassword(login, password string) (*models.Users, error)
	//CreateImage(creator *models.Users, path string) (*models.Images, error)
	//CreateAvatar(creator *models.Users, path string) (*models.Avatars, error)
}

type PgStorage struct {
	Conn *pgxpool.Pool
}

func NewPgStorage(storage config.Storage) (Storage, error) {
	const fn = "storage.New"
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		storage.User,
		storage.Password,
		storage.Path,
		storage.DbName,
	)
	connConf, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	connConf.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}
	conn, err := pgxpool.NewWithConfig(context.Background(), connConf)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	if err = conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("%s: error pinging: %w", fn, err)
	}

	return PgStorage{conn}, nil
}

func (p PgStorage) CreateUser(login, password string) (uuid.UUID, error) { return uuid.New(), nil }
