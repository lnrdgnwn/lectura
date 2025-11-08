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

func userOK(c *fiber.Ctx, status int, msg string, data any) error {
	resp := fiber.Map{
		"success": true,
		"message": msg,
	}
	if data != nil {
		resp["data"] = data
	}
	return c.Status(status).JSON(resp)
}

func userOKList(c *fiber.Ctx, msg string, data any, page, limit int, total int64) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": msg,
		"data":    data,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func userFail(c *fiber.Ctx, status int, msg string, err error) error {
	resp := fiber.Map{
		"success": false,
		"message": msg,
	}
	if err != nil && os.Getenv("APP_ENV") != "production" {
		resp["error"] = err.Error()
	}
	return c.Status(status).JSON(resp)
}

func getAuthFromAccessCookieUser(c *fiber.Ctx) (uint, string, error) {
	enc := c.Cookies("access_token", "")
	if enc == "" {
		return 0, "", errors.New("missing access_token cookie")
	}

	tokenStr, err := utils.Decrypt(enc)
	if err != nil || strings.TrimSpace(tokenStr) == "" {
		return 0, "", errors.New("invalid encrypted cookie")
	}

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

	if expVal, ok := claims["exp"]; ok {
		if exp, ok := expVal.(float64); ok {
			if time.Unix(int64(exp), 0).Before(time.Now()) {
				return 0, "", errors.New("token expired")
			}
		}
	}

	var uid uint
	switch v := claims["id"].(type) {
	case float64:
		uid = uint(v)
	case string:
		if n, err := strconv.Atoi(v); err == nil {
			uid = uint(n)
		} else {
			return 0, "", errors.New("token id not numeric")
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

// GetMe godoc
// @Summary      Get profil user login
// @Description  Mengambil detail profil user yang sedang login berdasarkan access_token (cookie).
// @Tags         Users
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Detail profil berhasil diambil"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized / token tidak valid"
// @Failure      404  {object}  map[string]interface{}  "User tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil user"
// @Router       /users/me [get]
func GetMe(c *fiber.Ctx) error {
	uid, _, err := getAuthFromAccessCookieUser(c)
	if err != nil {
		return userFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	var u models.User
	if err := database.DB.First(&u, uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return userFail(c, http.StatusNotFound, "User tidak ditemukan", nil)
		}
		return userFail(c, http.StatusInternalServerError, "Gagal mengambil user", err)
	}
	u.PasswordHash = ""
	return userOK(c, http.StatusOK, "Detail profil berhasil diambil", u)
}

// UpdateProfile godoc
// @Summary      Update profil user login
// @Description  Mengubah username, email, dan/atau foto profil user yang sedang login. Bisa via JSON atau form-data.
// @Tags         Users
// @Security     CookieAuth
// @Accept       multipart/form-data
// @Accept       application/json
// @Produce      json
// @Param        username          formData  string  false  "Username baru"
// @Param        email             formData  string  false  "Email baru"
// @Param        password          formData  string  false  "Password baru (opsional)"
// @Param        profile_picture   formData  file    false  "Foto profil baru"
// @Success      200               {object}  map[string]interface{}  "Profil berhasil diperbarui"
// @Failure      400               {object}  map[string]interface{}  "Payload tidak valid"
// @Failure      401               {object}  map[string]interface{}  "Unauthorized"
// @Failure      409               {object}  map[string]interface{}  "Username atau email sudah digunakan"
// @Failure      500               {object}  map[string]interface{}  "Gagal menyimpan perubahan profil"
// @Router       /users/me [put]
func UpdateProfile(c *fiber.Ctx) error {
	uid, _, err := getAuthFromAccessCookieUser(c)
	if err != nil {
		return userFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	var u models.User
	if err := database.DB.First(&u, uid).Error; err != nil {
		return userFail(c, http.StatusInternalServerError, "Gagal mengambil user", err)
	}

	var payload struct {
		Username string `json:"username" form:"username"`
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}
	_ = c.BodyParser(&payload)

	if payload.Username == "" {
		payload.Username = strings.TrimSpace(c.FormValue("username"))
	}
	if payload.Email == "" {
		payload.Email = strings.TrimSpace(c.FormValue("email"))
	}
	if payload.Password == "" {
		payload.Password = c.FormValue("password")
	}

	if s := strings.TrimSpace(payload.Username); s != "" && s != u.Username {
		var cnt int64
		if err := database.DB.Model(&models.User{}).Where("username = ? AND id <> ?", s, u.ID).Count(&cnt).Error; err != nil {
			return userFail(c, http.StatusInternalServerError, "Gagal memeriksa duplikasi username", err)
		}
		if cnt > 0 {
			return userFail(c, http.StatusConflict, "Username sudah digunakan", nil)
		}
		u.Username = s
	}
	if s := strings.TrimSpace(payload.Email); s != "" && s != u.Email {
		var cnt int64
		if err := database.DB.Model(&models.User{}).Where("email = ? AND id <> ?", s, u.ID).Count(&cnt).Error; err != nil {
			return userFail(c, http.StatusInternalServerError, "Gagal memeriksa duplikasi email", err)
		}
		if cnt > 0 {
			return userFail(c, http.StatusConflict, "Email sudah digunakan", nil)
		}
		u.Email = s
	}

	if payload.Password != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			return userFail(c, http.StatusInternalServerError, "Gagal memproses password", err)
		}
		u.PasswordHash = string(h)
	}

	if _, ferr := c.FormFile("profile_picture"); ferr == nil {
		if path, err := utils.SaveFile(c, "profile_picture", "profile_picture"); err == nil {
			u.ProfilePicture = &path
		} else {
			return userFail(c, http.StatusBadRequest, "Gagal menyimpan foto profil", err)
		}
	}

	u.UpdatedAt = time.Now()
	if err := database.DB.Save(&u).Error; err != nil {
		return userFail(c, http.StatusInternalServerError, "Gagal menyimpan perubahan profil", err)
	}

	u.PasswordHash = ""
	return userOK(c, http.StatusOK, "Profil berhasil diperbarui", u)
}

// ListUsers godoc
// @Summary      List semua user (admin)
// @Description  Mengambil daftar user dengan pagination. Hanya bisa diakses oleh admin.
// @Tags         Users
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        page   query     int  false  "Halaman (default 1)"
// @Param        limit  query     int  false  "Jumlah per halaman (default 20)"
// @Success      200    {object}  map[string]interface{}  "Daftar user + meta pagination"
// @Failure      401    {object}  map[string]interface{}  "Unauthorized"
// @Failure      403    {object}  map[string]interface{}  "Forbidden - admin only"
// @Failure      500    {object}  map[string]interface{}  "Gagal mengambil daftar user"
// @Router       /admin/users [get]
func ListUsers(c *fiber.Ctx) error {
	_, role, err := getAuthFromAccessCookieUser(c)
	if err != nil {
		return userFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}
	if strings.ToLower(role) != "admin" {
		return userFail(c, http.StatusForbidden, "Forbidden - admin only", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	if err := database.DB.Model(&models.User{}).Count(&total).Error; err != nil {
		return userFail(c, http.StatusInternalServerError, "Gagal menghitung total user", err)
	}

	var users []models.User
	if err := database.DB.Model(&models.User{}).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&users).Error; err != nil {
		return userFail(c, http.StatusInternalServerError, "Gagal mengambil daftar user", err)
	}
	for i := range users {
		users[i].PasswordHash = ""
	}

	return userOKList(c, "Daftar user berhasil diambil", users, page, limit, total)
}

// ChangePassword godoc
// @Summary      Ganti password user login
// @Description  Mengganti password user yang sedang login. Wajib kirim password lama & baru.
// @Tags         Users
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "Body: {\"old_password\": \"lama\", \"new_password\": \"baruMinimal8\"}"
// @Success      200   {object}  map[string]interface{}  "Password berhasil diubah"
// @Failure      400   {object}  map[string]interface{}  "Payload tidak valid / rule gagal"
// @Failure      401   {object}  map[string]interface{}  "Unauthorized / password lama salah"
// @Failure      404   {object}  map[string]interface{}  "User tidak ditemukan"
// @Failure      500   {object}  map[string]interface{}  "Gagal menyimpan password baru"
// @Router       /users/me/changepassword [put]
func ChangePassword(c *fiber.Ctx) error {
	uid, _, err := getAuthFromAccessCookieUser(c)
	if err != nil {
		return userFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	var payload struct {
		OldPassword string `json:"old_password" form:"old_password"`
		NewPassword string `json:"new_password" form:"new_password"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return userFail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}

	payload.OldPassword = strings.TrimSpace(payload.OldPassword)
	payload.NewPassword = strings.TrimSpace(payload.NewPassword)

	if payload.OldPassword == "" || payload.NewPassword == "" {
		return userFail(c, http.StatusBadRequest, "old_password dan new_password wajib diisi", nil)
	}
	if payload.OldPassword == payload.NewPassword {
		return userFail(c, http.StatusBadRequest, "Password baru harus berbeda dari password lama", nil)
	}
	if len(payload.NewPassword) < 8 {
		return userFail(c, http.StatusBadRequest, "Password baru minimal 8 karakter", nil)
	}

	var u models.User
	if err := database.DB.First(&u, uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return userFail(c, http.StatusNotFound, "User tidak ditemukan", nil)
		}
		return userFail(c, http.StatusInternalServerError, "Gagal mengambil user", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(payload.OldPassword)); err != nil {
		return userFail(c, http.StatusUnauthorized, "Password lama salah", nil)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return userFail(c, http.StatusInternalServerError, "Gagal memproses password baru", err)
	}

	u.PasswordHash = string(hashed)
	u.UpdatedAt = time.Now()
	if err := database.DB.Save(&u).Error; err != nil {
		return userFail(c, http.StatusInternalServerError, "Gagal menyimpan password baru", err)
	}

	return userOK(c, http.StatusOK, "Password berhasil diubah", nil)
}

// DeleteUser godoc
// @Summary      Hapus user (admin)
// @Description  Menghapus user berdasarkan ID. Hanya bisa diakses oleh admin.
// @Tags         Users
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  map[string]interface{}  "User berhasil dihapus"
// @Failure      400  {object}  map[string]interface{}  "Parameter id wajib / tidak valid"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden - admin only"
// @Failure      500  {object}  map[string]interface{}  "Gagal menghapus user"
// @Router       /admin/users/{id} [delete]
func DeleteUser(c *fiber.Ctx) error {
	_, role, err := getAuthFromAccessCookieUser(c)
	if err != nil {
		return userFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}
	if strings.ToLower(role) != "admin" {
		return userFail(c, http.StatusForbidden, "Forbidden - admin only", nil)
	}

	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return userFail(c, http.StatusBadRequest, "Parameter id wajib", nil)
	}
	if err := database.DB.Delete(&models.User{}, id).Error; err != nil {
		return userFail(c, http.StatusInternalServerError, "Gagal menghapus user", err)
	}
	return userOK(c, http.StatusOK, "User berhasil dihapus", nil)
}
