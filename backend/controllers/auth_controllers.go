// controllers/auth_controllers.go
package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
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
)

// ----------------- helpers -----------------

// generateRandomHex menghasilkan hex string acak (n bytes)
func generateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// signAccessToken menandatangani JWT access token (HMAC)
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

// ----------------- handlers -----------------

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
// Returns access_token + refresh_token (and expiry). Uses Redis if available, otherwise DB.
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

	// Buat refresh token menggunakan utils.GenerateRefreshToken (mirip contoh)
	signedRefresh, err := utils.GenerateRefreshToken(c, int(u.ID), u.Username, u.Role)
	if err != nil {
		// fallback: jika utils tidak ada atau error, generate random hex (basic)
		raw, rerr := generateRandomHex(32)
		if rerr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal membuat refresh token"})
		}
		signedRefresh = raw
	}

	// collect ip & ua pointers
	ipPtr := getIPPtr(c)
	uaPtr := getUserAgentPtr(c)

	// SINGLE-ACTIVE: simpan refresh token, pastikan token lama tidak aktif
	if database.Rdb != nil {
		rt := models.RefreshTokens{
			RefreshToken: signedRefresh,
			ParentToken:  "",
			UserID:       int(u.ID),
			Exp:          time.Now().Add(7 * 24 * time.Hour).Unix(),
			IPAddress:    ipPtr,
			UserAgent:    uaPtr,
		}
		if err := utils.SaveRefreshToken(c, rt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menyimpan refresh token (redis)", "error": err.Error()})
		}
	} else {
		// fallback DB: hapus token lama (single active) lalu simpan baru
		_ = database.DB.Where("user_id = ?", int(u.ID)).Delete(&models.RefreshTokens{}).Error
		rt := models.RefreshTokens{
			RefreshToken: signedRefresh,
			ParentToken:  "",
			UserID:       int(u.ID),
			Exp:          time.Now().Add(7 * 24 * time.Hour).Unix(),
			IPAddress:    ipPtr,
			UserAgent:    uaPtr,
		}
		if err := database.DB.Create(&rt).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menyimpan refresh token (db)", "error": err.Error()})
		}
	}

	return c.JSON(fiber.Map{
		"access_token":       accessToken,
		"access_expires_at":  exp.Format(time.RFC3339),
		"refresh_token":      signedRefresh,
		"refresh_expires_at": time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
	})
}

// RefreshToken: POST /api/v1/auth/refresh
// Flow mirip contoh: decrypt cookie/body, validasi via Redis (preferred) or DB fallback, rotate token.
func RefreshToken(c *fiber.Ctx) error {
	rdbCtx := context.Background()

	// ambil encrypted token dari cookie "refresh_token" atau body JSON { "refresh_token": "..." }
	enc := c.Cookies("refresh_token", "")
	if enc == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.BodyParser(&body)
		enc = strings.TrimSpace(body.RefreshToken)
	}
	if enc == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token missing"})
	}

	// Decrypt untuk mendapat refreshStr (JWT/genereted string)
	refreshStr, err := utils.Decrypt(enc)
	if err != nil {
		// Jika utils.Decrypt gagal, mungkin enc adalah token plain (fallback) => gunakan enc as-is
		refreshStr = enc
	}

	// Parse JWT (jika format JWT); jika bukan JWT, parse akan error and we treat as plain token for DB lookup
	token, err := jwt.Parse(refreshStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		secret := os.Getenv("JWT_SECRET")
		return []byte(secret), nil
	})

	var userIDInt int
	var claims jwt.MapClaims
	if err == nil && token != nil && token.Valid {
		claims, _ = token.Claims.(jwt.MapClaims)
		if s, ok := claims["id"]; ok {
			switch t := s.(type) {
			case float64:
				userIDInt = int(t)
			case int:
				userIDInt = t
			case string:
				if n, e := strconv.Atoi(t); e == nil {
					userIDInt = n
				}
			}
		} else if s, ok := claims["sub"]; ok {
			switch t := s.(type) {
			case float64:
				userIDInt = int(t)
			case int:
				userIDInt = t
			case string:
				if n, e := strconv.Atoi(t); e == nil {
					userIDInt = n
				}
			}
		}
	} else {
		userIDInt = 0
	}

	// prepare ip & ua pointers for rotation storage
	ipPtr := getIPPtr(c)
	uaPtr := getUserAgentPtr(c)

	// Try Redis flow if available
	if database.Rdb != nil {
		if userIDInt == 0 {
			// fallback to DB if token isn't JWT with user id
			goto DB_FALLBACK
		}

		userKey := fmt.Sprintf("refresh:%d", userIDInt)
		hash, err := database.Rdb.HGetAll(rdbCtx, userKey).Result()
		if err != nil || len(hash) == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token not found"})
		}

		storedEnc := hash["refresh_token"]
		storedRefresh, derr := utils.Decrypt(storedEnc)
		if derr != nil {
			// data korup -> delete key
			_ = database.Rdb.Del(rdbCtx, userKey)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "invalid stored refresh token"})
		}

		parentEnc := hash["parent_token"]
		parentDecrypted := ""
		if parentEnc != "" {
			if pd, perr := utils.Decrypt(parentEnc); perr == nil {
				parentDecrypted = pd
			}
		}

		expInt, _ := strconv.ParseInt(hash["exp"], 10, 64)

		// reuse detection
		if refreshStr == parentDecrypted && refreshStr != storedRefresh {
			_ = database.Rdb.Del(rdbCtx, userKey)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token mismatch, please login again"})
		}

		// expired?
		if time.Now().Unix() > expInt {
			_ = database.Rdb.Del(rdbCtx, userKey)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token expired"})
		}

		// success: generate new access token + rotate refresh
		name := ""
		role := ""
		if claims != nil {
			if v, ok := claims["name"].(string); ok {
				name = v
			}
			if v, ok := claims["role"].(string); ok {
				role = v
			}
		}

		accessTTL := 15
		if v := os.Getenv("ACCESS_TTL_MIN"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				accessTTL = n
			}
		}
		accessToken, exp, err := signAccessToken(uint(userIDInt), role, accessTTL)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed generate access token"})
		}

		// generate fresh refresh token via utils
		newSignedRefresh, err := utils.GenerateRefreshToken(c, userIDInt, name, role)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed generate refresh token"})
		}

		newRT := models.RefreshTokens{
			RefreshToken: newSignedRefresh,
			ParentToken:  refreshStr,
			UserID:       userIDInt,
			Exp:          time.Now().Add(7 * 24 * time.Hour).Unix(),
			IPAddress:    ipPtr,
			UserAgent:    uaPtr,
		}
		if err := utils.SaveRefreshToken(c, newRT); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed save refresh token"})
		}

		return c.JSON(fiber.Map{
			"access_token":       accessToken,
			"access_expires_at":  exp.Format(time.RFC3339),
			"refresh_token":      newSignedRefresh,
			"refresh_expires_at": time.Unix(newRT.Exp, 0).Format(time.RFC3339),
		})
	}

DB_FALLBACK:
	// fallback DB-based flow: lookup refresh token record by refreshStr
	var old models.RefreshTokens
	if err := database.DB.Where("refresh_token = ?", refreshStr).First(&old).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token not found"})
	}

	// expiry check
	if time.Now().Unix() > old.Exp {
		_ = database.DB.Where("refresh_token = ?", refreshStr).Delete(&models.RefreshTokens{})
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token expired"})
	}

	// reuse detection
	if old.ParentToken != "" && old.ParentToken == refreshStr && old.RefreshToken != refreshStr {
		_ = database.DB.Where("user_id = ?", old.UserID).Delete(&models.RefreshTokens{})
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "refresh token reuse detected"})
	}

	// load user
	var u models.User
	if err := database.DB.First(&u, old.UserID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "user not found"})
	}

	// generate access token
	accessTTL := 15
	if v := os.Getenv("ACCESS_TTL_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			accessTTL = n
		}
	}
	accessToken, exp, err := signAccessToken(uint(u.ID), u.Role, accessTTL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed generate access token"})
	}

	// rotate refresh token: create new record & delete old
	newRaw, err := generateRandomHex(32)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed create new refresh token"})
	}
	newRT := models.RefreshTokens{
		RefreshToken: newRaw,
		ParentToken:  old.RefreshToken,
		UserID:       old.UserID,
		Exp:          time.Now().Add(7 * 24 * time.Hour).Unix(),
		IPAddress:    ipPtr,
		UserAgent:    uaPtr,
	}
	tx := database.DB.Begin()
	if err := tx.Create(&newRT).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed save new refresh token"})
	}
	if err := tx.Where("refresh_token = ?", old.RefreshToken).Delete(&models.RefreshTokens{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed delete old refresh token"})
	}
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "failed commit transaction"})
	}

	return c.JSON(fiber.Map{
		"access_token":       accessToken,
		"access_expires_at":  exp.Format(time.RFC3339),
		"refresh_token":      newRaw,
		"refresh_expires_at": time.Unix(newRT.Exp, 0).Format(time.RFC3339),
	})
}

// Logout: POST /api/v1/auth/logout
// expects body { "refresh_token": "..." } or cookie "refresh_token"
// removes from Redis or DB
func Logout(c *fiber.Ctx) error {
	rt := c.Cookies("refresh_token", "")
	if rt == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.BodyParser(&body)
		rt = strings.TrimSpace(body.RefreshToken)
	}
	if rt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "refresh_token wajib"})
	}

	// try decrypt (if token was stored encrypted in cookie)
	dec, derr := utils.Decrypt(rt)
	if derr == nil {
		rt = dec
	}

	// If Redis available, try delete by user key (we need user id from token)
	if database.Rdb != nil {
		// try parse jwt to get user id (best-effort)
		if token, err := jwt.Parse(rt, func(t *jwt.Token) (interface{}, error) {
			secret := os.Getenv("JWT_SECRET")
			return []byte(secret), nil
		}); err == nil && token != nil {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				var userIDInt int
				if s, ok := claims["id"]; ok {
					switch t := s.(type) {
					case float64:
						userIDInt = int(t)
					case int:
						userIDInt = t
					case string:
						if n, e := strconv.Atoi(t); e == nil {
							userIDInt = n
						}
					}
				} else if s, ok := claims["sub"]; ok {
					switch t := s.(type) {
					case float64:
						userIDInt = int(t)
					case int:
						userIDInt = t
					case string:
						if n, e := strconv.Atoi(t); e == nil {
							userIDInt = n
						}
					}
				}
				if userIDInt != 0 {
					userKey := fmt.Sprintf("refresh:%d", userIDInt)
					_ = database.Rdb.Del(context.Background(), userKey)
					// Also delete any fallback DB record if exists
					_ = database.DB.Where("refresh_token = ?", rt).Delete(&models.RefreshTokens{}).Error
					return c.JSON(fiber.Map{"message": "logout sukses"})
				}
			}
		}
		// If we couldn't parse token to get user id, fallback to DB deletion (best-effort)
		_ = database.DB.Where("refresh_token = ?", rt).Delete(&models.RefreshTokens{}).Error
		return c.JSON(fiber.Map{"message": "logout sukses"})
	}

	// fallback DB: delete by refresh token value
	if err := database.DB.Where("refresh_token = ?", rt).Delete(&models.RefreshTokens{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menghapus token", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "logout sukses"})
}
