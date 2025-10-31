package controllers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"final_project/database"
	"final_project/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ----------------- Helpers -----------------

// generateRandomHex menghasilkan hex string acak (n bytes)
func generateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken = SHA256(raw) -> hex
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func signAccessToken(userID uint, role string, ttlMinutes int) (string, time.Time, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", time.Time{}, errors.New("JWT_SECRET not configured")
	}
	now := time.Now()
	exp := now.Add(time.Duration(ttlMinutes) * time.Minute)
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"iat":  now.Unix(),
		"exp":  exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	return signed, exp, err
}

func getIPPtr(c *fiber.Ctx) *string {
	ip := c.IP()
	if ip == "" {
		return nil
	}
	return &ip
}

func getUserAgentPtr(c *fiber.Ctx) *string {
	ua := c.Get("User-Agent")
	if ua == "" {
		return nil
	}
	return &ua
}

// ----------------- Auth handlers -----------------

// Register: POST /api/v1/auth/register
func Register(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Masukkan Field dengan benar"})
	}
	body.Username = strings.TrimSpace(body.Username)
	body.Email = strings.TrimSpace(body.Email)
	if body.Username == "" || body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "username, email, password wajib diisi"})
	}

	// cek duplicate
	var cnt int64
	database.DB.Model(&models.User{}).Where("username = ? OR email = ?", body.Username, body.Email).Count(&cnt)
	if cnt > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "username atau email sudah terpakai"})
	}

	pwHash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal memproses password"})
	}

	u := models.User{
		Username:     body.Username,
		Email:        body.Email,
		PasswordHash: string(pwHash),
		Role:         "READER",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := database.DB.Create(&u).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal membuat user", "error": err.Error()})
	}

	u.PasswordHash = ""
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "user terdaftar", "user": u})
}

// Login: POST /api/v1/auth/login
// returns access_token (JWT) + refresh_token (raw); server menyimpan hash
func Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Masukkan Field dengan benar"})
	}
	body.Email = strings.TrimSpace(body.Email)
	if body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "email dan password wajib diisi"})
	}

	var u models.User
	if err := database.DB.Where("email = ?", body.Email).First(&u).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "kredensial salah"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "kredensial salah"})
	}

	// buat access token
	accessTTL := 15
	if v := os.Getenv("ACCESS_TTL_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			accessTTL = n
		}
	}
	accessToken, exp, err := signAccessToken(u.ID, u.Role, accessTTL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal membuat access token"})
	}

	// SINGLE-ACTIVE STRATEGY:
	// Revoke semua refresh token lama untuk user ini terlebih dahulu,
	// sehingga hanya akan ada satu active token (yang akan kita buat sekarang).
	if err := database.DB.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", u.ID, false).
		Updates(map[string]interface{}{"revoked": true}).Error; err != nil {
		// tidak fatal untuk user, tapi log dan return error 500
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal revoke token lama"})
	}

	// buat refresh token raw, simpan hash di DB (lebih aman)
	raw, err := generateRandomHex(32)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal membuat refresh token"})
	}
	hashed := hashToken(raw)

	rt := models.RefreshToken{
		UserID:    u.ID,
		TokenHash: hashed,
		Revoked:   false,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // default 7 hari
		CreatedAt: time.Now(),
		IPAddress: getIPPtr(c),
		UserAgent: getUserAgentPtr(c),
	}
	if err := database.DB.Create(&rt).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menyimpan refresh token"})
	}

	return c.JSON(fiber.Map{
		"access_token":       accessToken,
		"access_expires_at":  exp.Format(time.RFC3339),
		"refresh_token":      raw, // kirim token mentah ke client — server hanya simpan hash
		"refresh_expires_at": rt.ExpiresAt.Format(time.RFC3339),
	})
}

// Refresh: POST /api/v1/auth/refresh
// body: { "refresh_token" }
func Refresh(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload tidak valid"})
	}
	rtValue := strings.TrimSpace(body.RefreshToken)
	if rtValue == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "refresh_token wajib"})
	}

	hash := hashToken(rtValue)

	var rtRecord models.RefreshToken
	if err := database.DB.Where("token_hash = ?", hash).First(&rtRecord).Error; err != nil {
		// token tidak ditemukan -> invalid
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token tidak valid"})
	}

	// cek kondisi token
	if rtRecord.Revoked {
		// sudah direvoke -> tolak
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token revoked"})
	}
	if time.Now().After(rtRecord.ExpiresAt) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token kadaluarsa"})
	}

	// ambil user
	var u models.User
	if err := database.DB.First(&u, rtRecord.UserID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "user tidak ditemukan"})
	}

	// buat access token baru saja (tidak mengganti refresh token)
	accessTTL := 15
	if v := os.Getenv("ACCESS_TTL_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			accessTTL = n
		}
	}
	accessToken, exp, err := signAccessToken(u.ID, u.Role, accessTTL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal membuat access token"})
	}

	// kembalikan access token baru dan juga refresh token yang sama (client tetap
	// memakai rtValue sampai ia di-revoke atau expired)
	return c.JSON(fiber.Map{
		"access_token":       accessToken,
		"access_expires_at":  exp.Format(time.RFC3339),
		"refresh_token":      rtValue,
		"refresh_expires_at": rtRecord.ExpiresAt.Format(time.RFC3339),
	})
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

// Logout: POST /api/v1/auth/logout
// body: { "refresh_token" } -> set revoked = true
func Logout(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload tidak valid"})
	}
	rtValue := strings.TrimSpace(body.RefreshToken)
	if rtValue == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "refresh_token wajib"})
	}
	hash := hashToken(rtValue)
	if err := database.DB.Model(&models.RefreshToken{}).Where("token_hash = ?", hash).
		Update("revoked", true).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal revoke token"})
	}
	return c.JSON(fiber.Map{"message": "logout sukses"})
}
