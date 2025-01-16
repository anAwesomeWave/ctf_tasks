package migrations

import (
	"accessCtf/migrations/migutil"
	"context"
	"database/sql"
	"fmt"
	"github.com/anAwesomeWave/text2img"
	"github.com/pressly/goose/v3"
	"image/color"
	"image/jpeg"
	"os"
)

func init() {
	goose.AddMigrationContext(upCreateAdmin, downCreateAdmin)
}

func createImage(dir, text string) (string, error) {
	path := "./static/assets/font.ttf"
	d, err := text2img.NewDrawer(text2img.Params{
		BackgroundColor: color.RGBA{R: 244, G: 235, B: 218, A: 255},
		TextColor:       color.RGBA{R: 51, G: 51, B: 51, A: 255},
		Width:           1400,
		FontPath:        path,
	})
	if err != nil {
		return "", err
	}

	img, err := d.Draw(text)

	if err != nil {
		return "", err
	}

	file, err := os.Create(dir + "/1.jpeg")

	if err != nil {
		return "", err
	}
	defer file.Close()

	err = jpeg.Encode(file, img, &jpeg.Options{Quality: 100})

	if err != nil {
		return "", err
	}
	return dir + "/1.jpeg", nil
}

func upCreateAdmin(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	adminLogin := os.Getenv("ADMIN_LOGIN")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminLogin == "" || adminPassword == "" {
		return fmt.Errorf("environment variables ADMIN_LOGIN or ADMIN_PASSWORD are not set")
	}
	adminId, err := migutil.CreateUser(tx, adminLogin, adminPassword, true)
	if err != nil {
		return err
	}
	// 2. сгенерировать картинку и положить ее в нужную папку
	adminImagePath := "./static/users/upload/" + adminId.String()
	if err := os.Mkdir(adminImagePath, 0744); err != nil {
		return err
	}
	flag := os.Getenv("CTF_FLAG")
	createdImagePath, err := createImage(adminImagePath, flag)

	if err != nil {
		return err
	}
	// 3. добавить в бд

	binary, err := adminId.MarshalBinary()
	if err != nil {
		return err
	}

	var imageId int64
	stmt := `INSERT INTO images(path, path_id, creator_id) VALUES($1, $2, $3) RETURNING id`
	if err := tx.QueryRow(stmt, createdImagePath, 1, binary).Scan(&imageId); err != nil {
		return err
	}

	if err := migutil.UploadUserAvatar(tx, *adminId, "./static/assets/admin/avatar.jpeg"); err != nil {
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
