package storage

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"race_cond/internal/storage/models"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("not Unique")
)

type AvatarPath = string

type Storage interface {
	CreateUser(login, password string) (*models.User, error)
	GetUser(login, password string) (*models.User, error)
	GetUserById(id int64) (*models.User, error)
	UpdateBalance(id int64) (*models.User, error)
	//IsAdmin(userId uuid.UUID) (bool, error)
	//GetUserByLoginPassword(login, password string) (*models.Users, error)
	//CreateImage(creator *models.Users, path string) (*models.Images, error)
	//CreateAvatar(creator *models.Users, path string) (*models.Avatars, error)
}

type LiteStrg struct {
	Db *sql.DB
}

func InitDB() LiteStrg {
	db, err := sql.Open("sqlite3", "./ctf.db")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
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
	return LiteStrg{Db: db}
}

func (p LiteStrg) CreateUser(login, password string) (*models.User, error) {
	const fn = "storage.CreateUser"

	res, err := p.Db.Exec("INSERT INTO users(login, password) VALUES (?, ?)", login, password)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return p.GetUserById(id)
}

func (p LiteStrg) GetUser(login, password string) (*models.User, error) {
	const fn = "storage.GetUser"

	var user models.User

	if err := p.Db.QueryRow("SELECT * FROM users WHERE login= ? AND password= ?", login, password).
		Scan(&user.Id, &user.Login, &user.Password, &user.Balance, &user.GotBonus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println(err)
			return nil, fmt.Errorf("%s: User {%s} with password {%s} not found : %w", fn, login, password, ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return &user, nil
}

func (p LiteStrg) GetUserById(id int64) (*models.User, error) {
	var user models.User
	err := p.Db.QueryRow("SELECT id, login, password, balance, got_bonus FROM users WHERE id = ?", id).
		Scan(&user.Id, &user.Login, &user.Password, &user.Balance, &user.GotBonus)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}

	return &user, nil
}

func (p LiteStrg) UpdateBalance(id int64) (*models.User, error) {
	user, err := p.GetUserById(id)
	if err != nil {
		return nil, err
	}
	result, err := p.Db.Exec("UPDATE users SET balance = ?, got_bonus = ? WHERE id = ?", user.Balance+100, 1, user.Id)
	if err != nil {
		return nil, err
	}
	time.Sleep(2 * time.Second)
	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return p.GetUserById(id)
}
