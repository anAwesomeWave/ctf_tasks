package storage

import (
	"accessCtf/internal/config"
	"accessCtf/internal/util"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("not Unique")
)

type Storage interface {
	CreateUser(login, password string) (*uuid.UUID, error)
	//GetUserByUUID(id uuid.UUID) (*models.Users, error)
	//GetUserByLoginPassword(login, password string) (*models.Users, error)
	//CreateImage(creator *models.Users, path string) (*models.Images, error)
	//CreateAvatar(creator *models.Users, path string) (*models.Avatars, error)
}

type PgStorage struct {
	Conn    *pgxpool.Pool
	timeout time.Duration
}

func NewPgStorage(storage config.Storage, timeout time.Duration) (*PgStorage, error) {
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := pgxpool.NewWithConfig(ctx, connConf)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err = conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s: error pinging: %w", fn, err)
	}

	return &PgStorage{conn, timeout}, nil
}

func (p PgStorage) CreateUser(login, password string) (*uuid.UUID, error) {
	/*
		1. проверить, что пользователя с таким логином еще нет в базе
		2. сделать хеш пароля
		3. сохранить в бд и отдать uuid
	*/
	const fn = "storage.CreateUser"

	pHash, err := util.GetHashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	var id uuid.UUID
	stmt := `INSERT INTO users(login, password_hash) VALUES($1, $2) RETURNING id`
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if err := p.Conn.QueryRow(ctx, stmt, login, pHash).Scan(&id); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.UniqueViolation {
				return nil, fmt.Errorf("%s: Login {%s} not unique: %w", fn, login, ErrExists)
			}
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &id, nil
}
