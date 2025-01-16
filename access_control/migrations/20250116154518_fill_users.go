package migrations

import (
	"accessCtf/migrations/migutil"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upFillUsers, downFillUsers)
}

func upFillUsers(ctx context.Context, tx *sql.Tx) error {
	users := []string{"user1", "user2"}
	assetsPath := "./static/assets/"
	for _, user := range users {
		login := user
		password := uuid.New().String()
		userId, err := migutil.CreateUser(tx, login, password, false)
		if err != nil {
			return err
		}
		userAssetsPath := assetsPath + user + "/"
		if err := migutil.UploadUserAvatar(tx, *userId, userAssetsPath+"avatar/1.jpeg"); err != nil {
			return err
		}
		for i := 1; i <= 2; i++ {
			if err := migutil.UploadImage(tx, *userId, userAssetsPath+fmt.Sprintf("%d.jpeg", i), i); err != nil {
				return err
			}
		}
	}
	// This code is executed when the migration is applied.
	return nil
}

func downFillUsers(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
