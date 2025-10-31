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

	// 1) Coba parse JSON body dulu (support application/json)
	var payload struct {
		Username string `json:"username" form:"username"`
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}
	_ = c.BodyParser(&payload) // BodyParser bekerja untuk JSON dan juga x-www-form-urlencoded

	// 2) Jika BodyParser gagal atau kosong, fallback ke FormValue (berguna saat multipart/form-data)
	// (BodyParser sudah menangani urlencoded; untuk multipart kita perlu FormValue dan FormFile)
	if payload.Username == "" {
		payload.Username = c.FormValue("username")
	}
	if payload.Email == "" {
		payload.Email = c.FormValue("email")
	}
	if payload.Password == "" {
		payload.Password = c.FormValue("password")
	}

	// apply changes
	if strings.TrimSpace(payload.Username) != "" {
		u.Username = strings.TrimSpace(payload.Username)
	}
	if strings.TrimSpace(payload.Email) != "" {
		u.Email = strings.TrimSpace(payload.Email)
	}
	if payload.Password != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menghash password"})
		}
		u.PasswordHash = string(h)
	}

	// 3) Handle upload avatar (multipart/form-data)
	// Note: FormFile hanya bekerja jika Content-Type adalah multipart/form-data
	if _, ferr := c.FormFile("profile_picture"); ferr == nil {
		if path, err := utils.SaveFile(c, "profile_picture", "cover"); err == nil {
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
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "forbidden - admin only"})
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

func ChangePassword(c *fiber.Ctx) error {
	uid, err := getUserIDFromLocals2(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	// support JSON atau form-data
	var payload struct {
		OldPassword string `json:"old_password" form:"old_password"`
		NewPassword string `json:"new_password" form:"new_password"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "payload tidak valid"})
	}

	payload.OldPassword = strings.TrimSpace(payload.OldPassword)
	payload.NewPassword = strings.TrimSpace(payload.NewPassword)

	if payload.OldPassword == "" || payload.NewPassword == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "old_password dan new_password wajib diisi"})
	}
	if payload.OldPassword == payload.NewPassword {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "password baru harus berbeda dari password lama"})
	}
	// opsional: validasi panjang minimal password
	if len(payload.NewPassword) < 8 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "password baru minimal 8 karakter"})
	}

	// ambil user
	var u models.User
	if err := database.DB.First(&u, uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "user tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil user"})
	}

	// verifikasi old password
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(payload.OldPassword)); err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "password lama salah"})
	}

	// hash password baru dan simpan
	hashed, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal memproses password baru"})
	}

	u.PasswordHash = string(hashed)
	u.UpdatedAt = time.Now()

	if err := database.DB.Save(&u).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menyimpan password baru"})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "password berhasil diubah"})
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
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "forbidden - admin only"})
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
