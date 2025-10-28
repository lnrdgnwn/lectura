package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"final_project/database"
	"final_project/models"
	"final_project/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// helper konversi user_id dari locals
func getUserIDFromLocals2(c *fiber.Ctx) (uint, error) {
	v := c.Locals("user_id")
	if v == nil {
		return 0, errors.New("user_id missing")
	}
	switch t := v.(type) {
	case float64:
		return uint(t), nil
	case int:
		return uint(t), nil
	case int64:
		return uint(t), nil
	case uint:
		return t, nil
	case string:
		if n, err := strconv.Atoi(t); err == nil {
			return uint(n), nil
		}
	}
	return 0, errors.New("cannot convert user_id")
}

// GetMe: GET /api/v1/users/me
func GetMe(c *fiber.Ctx) error {
	uid, err := getUserIDFromLocals2(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}
	var u models.User
	if err := database.DB.First(&u, uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "user tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil user"})
	}
	u.PasswordHash = ""
	return c.JSON(fiber.Map{"user": u})
}

// UpdateProfile: PUT /api/v1/users/me (form-data atau json)
func UpdateProfile(c *fiber.Ctx) error {
	uid, err := getUserIDFromLocals2(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}
	var u models.User
	if err := database.DB.First(&u, uid).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil user"})
	}

	// ambil fields
	username := c.FormValue("username")
	email := c.FormValue("email")
	newPass := c.FormValue("password")

	if username != "" {
		u.Username = strings.TrimSpace(username)
	}
	if email != "" {
		u.Email = strings.TrimSpace(email)
	}
	if newPass != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menghash password"})
		}
		u.PasswordHash = string(h)
	}

	// upload avatar optional (form key "avatar")
	if _, ferr := c.FormFile("avatar"); ferr == nil {
		if path, err := utils.SaveFile(c, "avatar", "avatars"); err == nil {
			u.ProfilePicture = &path
		}
	}

	u.UpdatedAt = time.Now()
	if err := database.DB.Save(&u).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menyimpan perubahan"})
	}

	u.PasswordHash = ""
	return c.JSON(fiber.Map{"message": "profil diperbarui", "user": u})
}

// ListUsers: GET /api/v1/admin/users?page=1&limit=20  (admin only)
func ListUsers(c *fiber.Ctx) error {
	// defensive role check
	role := c.Locals("role")
	if role == nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "forbidden"})
	}
	if rs, ok := role.(string); ok {
		if strings.ToLower(rs) != "admin" {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "forbidden"})
		}
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var users []models.User
	if err := database.DB.Model(&models.User{}).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil user"})
	}
	// sembunyikan password
	for i := range users {
		users[i].PasswordHash = ""
	}
	return c.JSON(fiber.Map{"page": page, "limit": limit, "data": users})
}

// DeleteUser: DELETE /api/v1/admin/users/:id (admin only)
func DeleteUser(c *fiber.Ctx) error {
	// defensive role check
	role := c.Locals("role")
	if role == nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "forbidden"})
	}
	if rs, ok := role.(string); ok {
		if strings.ToLower(rs) != "admin" {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "forbidden"})
		}
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "id wajib"})
	}
	if err := database.DB.Delete(&models.User{}, id).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menghapus user"})
	}
	return c.JSON(fiber.Map{"message": "user dihapus"})
}
