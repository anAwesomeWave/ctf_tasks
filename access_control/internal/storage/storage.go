package storage

import (
	"accessCtf/internal/config"
	"accessCtf/internal/storage/models"
	"accessCtf/internal/util"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
	"log"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("not Unique")
)

type AvatarPath = string

type Storage interface {
	CreateUser(login, password string) (*uuid.UUID, error)
	GetUser(login, password string) (*models.Users, error)
	GetUserById(id uuid.UUID) (*models.Users, error)
	GetMaxUserImageId(userId uuid.UUID) (int64, error)
	UpdateUserLogin(userId uuid.UUID, newLogin string) error
	InsertImage(userId uuid.UUID, imageId int64, imagePath string) (int64, error)
	GetAllImagesWithUserInfo() ([]*models.ImageWithUser, error)
	GetAllUserImages(userId uuid.UUID) ([]*models.Images, error)
	GetMaxUserAvatarId(userId uuid.UUID) (int64, error)
	InsertAvatar(userId uuid.UUID, avatarId int64, avatarPath string) (int64, error)
	GetLastUploadAvatar(userId uuid.UUID) (AvatarPath, error)
	//IsAdmin(userId uuid.UUID) (bool, error)
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

func (p PgStorage) GetUser(login, password string) (*models.Users, error) {
	const fn = "storage.GetUser"

	var user models.Users
	stmt := `SELECT * FROM users WHERE login = $1`
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if err := p.Conn.QueryRow(ctx, stmt, login).Scan(&user.Id, &user.Login, &user.PasswordHash, &user.IsAdmin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: User {%s} with password {%s} not found : %w", fn, login, password, ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	if !util.IsHashEqualPassword(user.PasswordHash, password) {
		return nil, fmt.Errorf("%s: User {%s} with password {%s} not found: passwords don't match : %w", fn, login, password, ErrNotFound)
	}
	return &user, nil
}

func (p PgStorage) GetUserById(id uuid.UUID) (*models.Users, error) {
	const fn = "storage.GetUser"

	var user models.Users
	stmt := `SELECT * FROM users WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	binary, err := id.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	if err := p.Conn.QueryRow(ctx, stmt, binary).Scan(&user.Id, &user.Login, &user.PasswordHash, &user.IsAdmin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: User with id {%s} not found : %w", fn, id.String(), ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return &user, nil
}
func (p PgStorage) GetMaxUserImageId(userId uuid.UUID) (int64, error) {
	const fn = "storage.GetMaxUserImageId"
	var imageId int64
	stmt := `SELECT coalesce(MAX(path_id), 0) FROM images WHERE images.creator_id  = $1`
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	binary, err := userId.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}
	if err := p.Conn.QueryRow(ctx, stmt, binary).Scan(&imageId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s: No images found for user with id {%s}: %w", fn, userId.String(), ErrNotFound)
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}
	return imageId, nil
}

func (p PgStorage) InsertImage(userId uuid.UUID, imageId int64, imagePath string) (int64, error) {
	const fn = "storage.InsertImage"

	var id int64

	binary, err := userId.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	stmt := `INSERT INTO images(path, path_id, creator_id) VALUES($1, $2, $3) RETURNING id`
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if err := p.Conn.QueryRow(ctx, stmt, imagePath, imageId, binary).Scan(&id); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.UniqueViolation {
				log.Println(fn, pgError)
				return 0, fmt.Errorf("%s: image is not unique %d, %s: %w", fn, imageId, imagePath, ErrExists)
			}
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (p PgStorage) GetAllImagesWithUserInfo() ([]*models.ImageWithUser, error) {
	const fn = "storage.GetAllImagesWithUserInfo"

	stmt := `SELECT images.path, login, is_admin, avatars.path FROM users JOIN images ON images.creator_id = users.id LEFT OUTER JOIN avatars ON users.id = avatars.owner_id ORDER BY images.creation_time DESC`
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	rows, err := p.Conn.Query(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("%s: error quering data from db: %v", fn, err)
	}
	defer rows.Close()

	var images []*models.ImageWithUser

	for rows.Next() {
		var image models.ImageWithUser
		var nullStr sql.NullString
		if err := rows.Scan(&image.ImagePath, &image.Login, &image.IsAdmin, &nullStr); err != nil {
			return nil, fmt.Errorf("%s: error getting next row %v", fn, err)
		}
		if nullStr.Valid {
			image.AvatarPath = nullStr.String
		}
		images = append(images, &image)
	}
	return images, nil
}

func (p PgStorage) GetAllUserImages(userId uuid.UUID) ([]*models.Images, error) {
	const fn = "storage.GetAllUserImages"

	stmt := `SELECT id, path, path_id FROM images WHERE creator_id = $1 ORDER BY creation_time DESC`

	binary, err := userId.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	rows, err := p.Conn.Query(ctx, stmt, binary)
	if err != nil {
		return nil, fmt.Errorf("%s: error quering data from db: %v", fn, err)
	}
	defer rows.Close()

	var images []*models.Images

	for rows.Next() {
		var image models.Images
		if err := rows.Scan(&image.Id, &image.Path, &image.PathId); err != nil {
			return nil, fmt.Errorf("%s: error getting next row %v", fn, err)
		}
		images = append(images, &image)
	}
	return images, nil
}

func (p PgStorage) GetMaxUserAvatarId(userId uuid.UUID) (int64, error) {
	const fn = "storage.GetMaxUserAvatarId"
	var avatarId int64
	stmt := `SELECT coalesce(MAX(path_id), 0) FROM avatars WHERE avatars.owner_id  = $1`
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	binary, err := userId.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}
	if err := p.Conn.QueryRow(ctx, stmt, binary).Scan(&avatarId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s: No avatars found for user with id {%s}: %w", fn, userId.String(), ErrNotFound)
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}
	return avatarId, nil
}

func (p PgStorage) GetLastUploadAvatar(userId uuid.UUID) (AvatarPath, error) {
	const fn = "storage.GetLastUploadAvatar"
	stmt := `SELECT path FROM users JOIN avatars ON avatars.owner_id = $1 ORDER BY creation_time DESC LIMIT 1`

	binary, err := userId.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}
	var path AvatarPath
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	if err := p.Conn.QueryRow(ctx, stmt, binary).Scan(&path); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: No avatars found for user with id {%s}: %w", fn, userId.String(), ErrNotFound)
		}
		return "", fmt.Errorf("%s: %w", fn, err)
	}
	return path, nil
}

func (p PgStorage) InsertAvatar(userId uuid.UUID, avatarId int64, avatarPath string) (int64, error) {
	const fn = "storage.InsertAvatar"

	var id int64

	binary, err := userId.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	stmt := `INSERT INTO avatars(path, path_id, owner_id) VALUES($1, $2, $3) RETURNING id`
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if err := p.Conn.QueryRow(ctx, stmt, avatarPath, avatarId, binary).Scan(&id); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.UniqueViolation {
				log.Println(fn, pgError)
				return 0, fmt.Errorf("%s: avatar is not unique %d, %s: %w", fn, avatarId, avatarPath, ErrExists)
			}
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (p PgStorage) UpdateUserLogin(userId uuid.UUID, newLogin string) error {
	const fn = "storage.UpdateUserLogin"

	binary, err := userId.MarshalBinary()
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	stmt := `UPDATE users SET login = $1 WHERE id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	if _, err := p.Conn.Exec(ctx, stmt, newLogin, binary); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: User not found : %w", fn, ErrNotFound)
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
