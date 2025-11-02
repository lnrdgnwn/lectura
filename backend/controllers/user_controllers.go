package controllers

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"final_project/database"
	"final_project/models"
	"final_project/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

/* ===================== AUTH HELPERS (COOKIE-BASED) ===================== */

// Ambil user_id & role dari cookie access_token (terenkripsi AES-GCM).
func getAuthFromAccessCookieUser(c *fiber.Ctx) (uint, string, error) {
	enc := c.Cookies("access_token", "")
	if enc == "" {
		return 0, "", errors.New("missing access_token cookie")
	}

	// Cookie kita simpan terenkripsi → decrypt dulu ke plaintext JWT
	tokenStr, err := utils.Decrypt(enc)
	if err != nil || strings.TrimSpace(tokenStr) == "" {
		return 0, "", errors.New("invalid encrypted cookie")
	}

	// Parse & verify JWT (HS256) dengan secret env
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return 0, "", errors.New("server misconfigured: JWT_SECRET missing")
	}
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || tok == nil || !tok.Valid {
		return 0, "", errors.New("invalid or expired token")
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", errors.New("invalid token claims")
	}

	// Validasi exp secara eksplisit (opsional)
	if expVal, ok := claims["exp"]; ok {
		switch exp := expVal.(type) {
		case float64:
			if time.Unix(int64(exp), 0).Before(time.Now()) {
				return 0, "", errors.New("token expired")
			}
		}
	}

	// Ambil id (claim "id") & role
	var uid uint
	switch v := claims["id"].(type) {
	case float64:
		uid = uint(v)
	case string:
		if n, err := strconv.Atoi(v); err == nil {
			uid = uint(n)
		}
	default:
		return 0, "", errors.New("token missing id")
	}

	role := ""
	if r, ok := claims["role"].(string); ok {
		role = r
	}

	return uid, role, nil
}

/* ===================== HANDLERS ===================== */

// GetMe: GET /api/v1/users/me
func GetMe(c *fiber.Ctx) error {
	uid, _, err := getAuthFromAccessCookieUser(c)
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
	uid, _, err := getAuthFromAccessCookieUser(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	var u models.User
	if err := database.DB.First(&u, uid).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil user"})
	}

	// 1) BodyParser akan handle JSON & x-www-form-urlencoded
	var payload struct {
		Username string `json:"username" form:"username"`
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}
	_ = c.BodyParser(&payload)

	// 2) Fallback untuk multipart/form-data (kalau ada)
	if payload.Username == "" {
		payload.Username = c.FormValue("username")
	}
	if payload.Email == "" {
		payload.Email = c.FormValue("email")
	}
	if payload.Password == "" {
		payload.Password = c.FormValue("password")
	}

	// apply perubahan
	if s := strings.TrimSpace(payload.Username); s != "" {
		u.Username = s
	}
	if s := strings.TrimSpace(payload.Email); s != "" {
		u.Email = s
	}
	if payload.Password != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menghash password"})
		}
		u.PasswordHash = string(h)
	}

	// 3) handle upload avatar (multipart)
	if _, ferr := c.FormFile("profile_picture"); ferr == nil {
		if path, err := utils.SaveFile(c, "profile_picture", "profile_picture"); err == nil {
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
	_, role, err := getAuthFromAccessCookieUser(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}
	if strings.ToLower(role) != "admin" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "forbidden - admin only"})
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
	for i := range users {
		users[i].PasswordHash = ""
	}
	return c.JSON(fiber.Map{"page": page, "limit": limit, "data": users})
}

// ChangePassword: PUT /api/v1/users/change-password
func ChangePassword(c *fiber.Ctx) error {
	uid, _, err := getAuthFromAccessCookieUser(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

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
	if len(payload.NewPassword) < 8 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "password baru minimal 8 karakter"})
	}

	var u models.User
	if err := database.DB.First(&u, uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "user tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil user"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(payload.OldPassword)); err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "password lama salah"})
	}

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
	_, role, err := getAuthFromAccessCookieUser(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}
	if strings.ToLower(role) != "admin" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "forbidden - admin only"})
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
