package handler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

type ChangePassForm struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (h *Handler) settings(c *gin.Context) {
	profile, err := database.GetInfoForProfilePage(c.GetString("username"))

	if err != nil {
		fmt.Println(err)
	}

	c.HTML(http.StatusOK, "settings.tmpl", gin.H{"Name3": profile.Name3})
}

func (h *Handler) passwd(c *gin.Context) {
	var form ChangePassForm
	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Неуспешно"})
		return
	}
	CurrentPassword := form.CurrentPassword
	NewPassword := form.NewPassword
	ConfirmPassword := form.ConfirmPassword
	username := c.GetString("username")

	conn, err := pgx.Connect(context.Background(), database.DbURL)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(context.Background())

	var ActualPasswordHash string

	query := `
  SELECT password_hash FROM users WHERE username = $1;
 `

	err = conn.QueryRow(context.Background(), query, username).Scan(&ActualPasswordHash)
	if err != nil {
		fmt.Println(err)
	}

	Hash := md5.Sum([]byte(CurrentPassword)) // Хэширование в MD5

	PasswordHash := hex.EncodeToString(Hash[:])
	if PasswordHash != ActualPasswordHash {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный пароль"})
	}

	if NewPassword != ConfirmPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Новый пароль не совпадает с паролем в блоке подтвержденный пароль"})
	}

	query = `
		UPDATE users
  		SET password_hash = $1
  		WHERE username = $2;
  	`

	NewHash := md5.Sum([]byte(NewPassword)) // Хэширование в MD5

	NewPasswordHash := hex.EncodeToString(NewHash[:])
	_, err = conn.Exec(context.Background(), query, NewPasswordHash, username)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
	}

}
