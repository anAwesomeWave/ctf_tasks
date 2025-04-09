package storage

import (
	"context"
	"errors"
	"fmt"
	"sqli/internal/config"
	"sqli/internal/storage/models"
	"time"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
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
	//IsAdmin(userId uuid.UUID) (bool, error)
	//GetUserByLoginPassword(login, password string) (*models.Users, error)
	//CreateImage(creator *models.Users, path string) (*models.Images, error)
	//CreateAvatar(creator *models.Users, path string) (*models.Avatars, error)
}

type liteStrg struct {
	Db *sql.DB
}

func InitDB() liteStrg {
	db, err := sql.Open("sqlite3", "./ctf.db")
	if err != nil {
		log.Fatal(err)
	}
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			login TEXT UNIQUE,
			password TEXT,
			balance INTEGER DEFAULT 0,
			got_bonus INTEGER DEFAULT 0
		)`)
	if err != nil {
		log.Fatal(err)
	}
	return liteStrg{Db: db}
}

func (p liteStrg) CreateUser(login, password string) (int, error) {
	const fn = "storage.CreateUser"

	res, err := p.db.Exec("INSERT INTO users(login, password) VALUES (?, ?)", login, password)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	id, err := res.LastInsertId()
    if err != nil {
		log.Println(err)
		return 0, err
    }
	return id, nil
}

func (p liteStrg) GetUser(login, password string) (*models.Users, error) {
	const fn = "storage.GetUser"

	var user models.Users
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if err := p.Conn.QueryRow(ctx, fmt.Sprintf("SELECT * FROM users WHERE login='%s' AND password='%s'", login, password)).Scan(&user.Id, &user.Login, &user.Password, &user.IsAdmin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println(err)
			return nil, fmt.Errorf("%s: User {%s} with password {%s} not found : %w", fn, login, password, ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return &user, nil
}

func (p liteStrg) GetUserById(id uuid.UUID) (*models.Users, error) {
	const fn = "storage.GetUser"

	var user models.Users
	stmt := `SELECT * FROM users WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	binary, err := id.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	if err := p.Conn.QueryRow(ctx, stmt, binary).Scan(&user.Id, &user.Login, &user.Password, &user.IsAdmin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: User with id {%s} not found : %w", fn, id.String(), ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return &user, nil
}

