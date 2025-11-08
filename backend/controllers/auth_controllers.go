package controllers

import (
	"context"
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

// Response helpers
func authok(c *fiber.Ctx, status int, msg string, data any) error {
	if data == nil {
		return c.Status(status).JSON(fiber.Map{
			"success": true,
			"message": msg,
		})
	}
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": msg,
		"data":    data,
	})
}

func authfail(c *fiber.Ctx, status int, msg string, err error) error {
	resp := fiber.Map{
		"success": false,
		"message": msg,
	}
	if err != nil && os.Getenv("APP_ENV") != "production" {
		resp["error"] = err.Error()
	}
	return c.Status(status).JSON(resp)
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

// Register godoc
// @Summary      Register user baru
// @Description  Membuat akun baru dengan username, email, dan password. Role default: user.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      models.RegisterRequest  true  "Data registrasi"
// @Success      201   {object}  map[string]interface{}  "User berhasil terdaftar"
// @Failure      400   {object}  map[string]interface{}
// @Failure      409   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /auth/register [post]
func Register(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return authfail(c, fiber.StatusBadRequest, "Payload tidak valid", err)
	}
	body.Username = strings.TrimSpace(body.Username)
	body.Email = strings.TrimSpace(body.Email)
	if body.Username == "" || body.Email == "" || body.Password == "" {
		return authfail(c, fiber.StatusBadRequest, "Username, email, dan password wajib diisi", nil)
	}

	var cnt int64
	if err := database.DB.Model(&models.User{}).
		Where("username = ? OR email = ?", body.Username, body.Email).
		Count(&cnt).Error; err != nil {
		return authfail(c, fiber.StatusInternalServerError, "Gagal memeriksa duplikasi akun", err)
	}
	if cnt > 0 {
		return authfail(c, fiber.StatusConflict, "Username atau email sudah digunakan", nil)
	}

	pwHash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return authfail(c, fiber.StatusInternalServerError, "Gagal memproses password", err)
	}

	u := models.User{
		Username:     body.Username,
		Email:        body.Email,
		PasswordHash: string(pwHash),
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := database.DB.Create(&u).Error; err != nil {
		return authfail(c, fiber.StatusInternalServerError, "Gagal membuat user", err)
	}

	u.PasswordHash = ""
	return authok(c, fiber.StatusCreated, "User berhasil terdaftar", fiber.Map{"user": u})
}

// Login godoc
// @Summary      Login
// @Description  Login dengan email & password. Token dikirim via cookie HttpOnly.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      models.LoginRequest  true  "Data login"
// @Success      200   {object}  map[string]interface{}  "Login sukses"
// @Failure      400   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /auth/login [post]
func Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return authfail(c, fiber.StatusBadRequest, "Payload tidak valid", err)
	}
	body.Email = strings.TrimSpace(body.Email)
	if body.Email == "" || body.Password == "" {
		return authfail(c, fiber.StatusBadRequest, "Email dan password wajib diisi", nil)
	}

	var u models.User
	if err := database.DB.Where("email = ?", body.Email).First(&u).Error; err != nil {
		return authfail(c, fiber.StatusUnauthorized, "Kredensial salah", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)); err != nil {
		return authfail(c, fiber.StatusUnauthorized, "Kredensial salah", nil)
	}

	if _, err := utils.GenerateAccessToken(c, int(u.ID), u.Username, u.Role); err != nil {
		return authfail(c, fiber.StatusInternalServerError, "Gagal membuat access token", err)
	}

	refreshJWT, err := utils.GenerateRefreshToken(c, int(u.ID), u.Username, u.Role)
	if err != nil {
		return authfail(c, fiber.StatusInternalServerError, "Gagal membuat refresh token", err)
	}

	ipPtr := getIPPtr(c)
	uaPtr := getUserAgentPtr(c)

	rt := models.RefreshToken{
		RefreshToken: refreshJWT,
		ParentToken:  "",
		UserID:       int(u.ID),
		Exp:          time.Now().Add(7 * 24 * time.Hour).Unix(),
		IPAddress:    ipPtr,
		UserAgent:    uaPtr,
	}

	if database.Rdb != nil {
		_ = database.Rdb.Del(context.Background(), fmt.Sprintf("refresh:%d", rt.UserID)).Err()
		if err := utils.SaveRefreshToken(c, rt); err != nil {
			return authfail(c, fiber.StatusInternalServerError, "Gagal menyimpan refresh token (redis)", err)
		}
	} else {
		_ = database.DB.Where("user_id = ?", rt.UserID).Delete(&models.RefreshToken{}).Error
		if err := database.DB.Create(&rt).Error; err != nil {
			return authfail(c, fiber.StatusInternalServerError, "Gagal menyimpan refresh token (db)", err)
		}
	}

	return authok(c, fiber.StatusOK, "Login sukses", nil)
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Menggunakan refresh_token (cookie) untuk membuat access_token baru (dan rotate refresh_token).
// @Tags         Auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Refresh token sukses, token baru diset di cookie"
// @Failure      401  {object}  map[string]interface{}  "Refresh token tidak tersedia/invalid/kedaluwarsa"
// @Failure      500  {object}  map[string]interface{}  "Gagal membuat/simpan token baru"
// @Router       /auth/refresh [post]
func RefreshToken(c *fiber.Ctx) error {
	rdbCtx := context.Background()

	encRefresh := c.Cookies("refresh_token")
	if encRefresh == "" {
		return authfail(c, fiber.StatusUnauthorized, "Refresh token tidak tersedia", nil)
	}

	refreshStr, err := utils.Decrypt(encRefresh)
	if err != nil {
		return authfail(c, fiber.StatusUnauthorized, "Refresh token tidak valid", err)
	}

	token, err := jwt.Parse(refreshStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return authfail(c, fiber.StatusUnauthorized, "Refresh token tidak valid", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := int(claims["id"].(float64))

	userKey := fmt.Sprintf("refresh:%d", userID)
	hash, err := database.Rdb.HGetAll(rdbCtx, userKey).Result()
	if err != nil || len(hash) == 0 {
		return authfail(c, fiber.StatusUnauthorized, "Refresh token tidak ditemukan", err)
	}

	hashRefreshtoken := hash["refresh_token"]
	refreshDecrypted, err := utils.Decrypt(hashRefreshtoken)
	if err != nil {
		return authfail(c, fiber.StatusUnauthorized, "Refresh token tidak valid", err)
	}

	hashParentToken := hash["parent_token"]
	parentDecrypted := ""
	if hashParentToken != "" {
		parentDecrypted, err = utils.Decrypt(hashParentToken)
		if err != nil {
			return authfail(c, fiber.StatusUnauthorized, "Refresh token tidak valid", err)
		}
	}

	expInt, _ := strconv.ParseInt(hash["exp"], 10, 64)

	if refreshStr == parentDecrypted && refreshStr != refreshDecrypted {
		database.Rdb.Del(rdbCtx, userKey)
		return authfail(c, fiber.StatusUnauthorized, "Refresh token terdeteksi reuse, silakan login kembali", nil)
	}

	if time.Now().Unix() > expInt {
		database.Rdb.Del(rdbCtx, userKey)
		return authfail(c, fiber.StatusUnauthorized, "Refresh token kedaluwarsa", nil)
	}

	if _, err := utils.GenerateAccessToken(c, userID, claims["name"].(string), claims["role"].(string)); err != nil {
		return authfail(c, fiber.StatusInternalServerError, "Gagal membuat access token", err)
	}

	signedRefresh, err := utils.GenerateRefreshToken(c, userID, claims["name"].(string), claims["role"].(string))
	if err != nil {
		return authfail(c, fiber.StatusInternalServerError, "Gagal membuat refresh token", err)
	}

	ipPtr := getIPPtr(c)
	uaPtr := getUserAgentPtr(c)

	newState := models.RefreshToken{
		RefreshToken: signedRefresh,
		ParentToken:  refreshDecrypted,
		UserID:       userID,
		Exp:          time.Now().Add(7 * 24 * time.Hour).Unix(),
		IPAddress:    ipPtr,
		UserAgent:    uaPtr,
	}

	if err := utils.SaveRefreshToken(c, newState); err != nil {
		return authfail(c, fiber.StatusInternalServerError, "Gagal menyimpan state refresh token", err)
	}

	if database.DB != nil {
		_ = database.DB.Where("user_id = ?", newState.UserID).Delete(&models.RefreshToken{}).Error
		if err := database.DB.Create(&newState).Error; err != nil {
			return authfail(c, fiber.StatusInternalServerError, "Gagal menyimpan refresh token ke database", err)
		}
	}

	return authok(c, fiber.StatusOK, "Refresh token sukses", nil)
}

// Logout godoc
// @Summary      Logout
// @Description  Logout user. Menghapus refresh token dari Redis/DB dan mengosongkan cookie.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      models.LogoutRequest  false  "Opsional jika tidak pakai cookie"
// @Success      200   {object}  map[string]interface{}  "Logout sukses"
// @Router       /auth/logout [post]
func Logout(c *fiber.Ctx) error {
	rtEnc := c.Cookies("refresh_token", "")
	if rtEnc == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.BodyParser(&body)
		rtEnc = strings.TrimSpace(body.RefreshToken)
	}

	if rtEnc != "" {
		rtPlain, derr := utils.Decrypt(rtEnc)
		if derr != nil {
			rtPlain = rtEnc
		}

		var userIDInt int
		if token, err := jwt.Parse(rtPlain, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		}); err == nil && token != nil {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				switch s := claims["id"].(type) {
				case float64:
					userIDInt = int(s)
				case int:
					userIDInt = s
				case string:
					if n, e := strconv.Atoi(s); e == nil {
						userIDInt = n
					}
				}
			}
		}

		if database.Rdb != nil && userIDInt != 0 {
			userKey := fmt.Sprintf("refresh:%d", userIDInt)
			_ = database.Rdb.Del(context.Background(), userKey)
		}
		if database.DB != nil {
			_ = database.DB.Where("refresh_token = ?", rtPlain).Delete(&models.RefreshToken{}).Error
		}
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
		HTTPOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: "Lax",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
		HTTPOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: "Lax",
	})

	return authok(c, fiber.StatusOK, "Logout sukses", nil)
}
