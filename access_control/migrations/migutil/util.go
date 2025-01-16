package migutil

import (
	"accessCtf/internal/util"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"io"
	"io/fs"
	"os"
)

func isExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func copyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Flush the file contents to disk
	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func CreateUser(tx *sql.Tx, login, password string, isAdmin bool) (*uuid.UUID, error) {
	pHash, err := util.GetHashPassword(password)
	if err != nil {
		return nil, err
	}

	// return id and create avatar with image
	// create assets folder and fill users images with it
	var userId uuid.UUID
	if err := tx.QueryRow(
		`INSERT INTO users(login, password_hash, is_admin) VALUES($1, $2, $3) RETURNING id`,
		login,
		pHash,
		isAdmin,
	).Scan(&userId); err != nil {
		return nil, err
	}
	return &userId, nil
}

func CreateUserAvatar(tx *sql.Tx, userId uuid.UUID, avatarAssetPath string) error {
	avatarPath := "./static/users/upload/avatars/" + userId.String()

	ex, err := isExists(avatarPath)
	if err != nil {
		return err
	}
	if !ex {
		if err := os.Mkdir(avatarPath, 0744); err != nil {
			return err
		}
	}
	avatarPath += "/1.jpeg"
	if err := copyFile(avatarAssetPath, avatarPath); err != nil {
		return err
	}

	stmt := `INSERT INTO avatars(path, path_id, owner_id) VALUES($1, $2, $3)`

	binary, err := userId.MarshalBinary()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(stmt, avatarPath, 1, binary); err != nil {
		return err
	}
	return nil
}
