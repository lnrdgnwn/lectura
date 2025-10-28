package controllers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"final_project/database"
	"final_project/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
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

// convert c.Locals("user_id") ke uint
func getUserIDFromLocals(c *fiber.Ctx) (uint, error) {
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
		// parse string
		if n, err := strconv.Atoi(t); err == nil {
			return uint(n), nil
		}
	}
	return 0, errors.New("cannot convert user_id")
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

// revokeAllUserTokens -> tandai semua token user sebagai revoked
func revokeAllUserTokens(userID uint) error {
	return database.DB.Model(&models.RefreshToken{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{"revoked": true}).Error
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload tidak valid"})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload tidak valid"})
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

	var old models.RefreshToken
	if err := database.DB.Where("token_hash = ?", hash).First(&old).Error; err != nil {
		// token tidak ditemukan -> kemungkinan reuse atau invalid
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token tidak valid"})
	}

	// cek kondisi token
	if old.Revoked {
		// reuse suspected — revoke semua token user
		_ = revokeAllUserTokens(old.UserID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token revoked"})
	}
	if time.Now().After(old.ExpiresAt) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token kadaluarsa"})
	}
	// kalau token sudah diganti (replaced_by != nil) => reuse
	if old.ReplacedBy != nil {
		_ = revokeAllUserTokens(old.UserID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token reuse detected"})
	}

	// load user
	var u models.User
	if err := database.DB.First(&u, old.UserID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "user tidak ditemukan"})
	}

	// buat access token baru
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

	// rotate refresh token: buat new token (hash) & set parent -> revoke old + set replaced_by
	newRaw, err := generateRandomHex(32)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal membuat refresh token baru"})
	}
	newHash := hashToken(newRaw)

	tx := database.DB.Begin()
	newRT := models.RefreshToken{
		UserID:    u.ID,
		TokenHash: newHash,
		ParentID:  &old.ID,
		Revoked:   false,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
		IPAddress: getIPPtr(c),
		UserAgent: getUserAgentPtr(c),
	}
	if err := tx.Create(&newRT).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menyimpan refresh token baru"})
	}
	// update old: set replaced_by = newRT.ID, revoked = true
	if err := tx.Model(&models.RefreshToken{}).Where("id = ?", old.ID).
		Updates(map[string]interface{}{"replaced_by": newRT.ID, "revoked": true}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal revoke token lama"})
	}
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal commit transaksi token"})
	}

	return c.JSON(fiber.Map{
		"access_token":       accessToken,
		"access_expires_at":  exp.Format(time.RFC3339),
		"refresh_token":      newRaw,
		"refresh_expires_at": newRT.ExpiresAt.Format(time.RFC3339),
	})
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
