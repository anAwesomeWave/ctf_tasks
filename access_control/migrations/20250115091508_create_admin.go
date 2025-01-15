package migrations

import (
	"accessCtf/internal/util"
	"context"
	"database/sql"
	"fmt"
	"github.com/pressly/goose/v3"
	"os"
)

func init() {
	goose.AddMigrationContext(upCreateAdmin, downCreateAdmin)
}

func upCreateAdmin(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	adminLogin := os.Getenv("ADMIN_LOGIN")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminLogin == "" || adminPassword == "" {
		return fmt.Errorf("environment variables ADMIN_LOGIN or ADMIN_PASSWORD are not set")
	}
	pHash, err := util.GetHashPassword(adminPassword)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO users(login, password_hash, is_admin) VALUES($1, $2, $3)`, adminLogin, pHash, true)
	if err != nil {
		return err
	}

	return nil
}

func downCreateAdmin(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	adminLogin := os.Getenv("ADMIN_LOGIN")
	if adminLogin == "" {
		return fmt.Errorf("environment variable ADMIN_LOGIN is not set")
	}
	_, err := tx.Exec(`DELETE FROM users WHERE login = $1`, adminLogin)
	if err != nil {
		return err
	}

	return nil
}
