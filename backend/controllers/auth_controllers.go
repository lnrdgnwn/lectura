// controllers/auth_controllers.go
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

func Register(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Field tidak valid"})
	}
	body.Username = strings.TrimSpace(body.Username)
	body.Email = strings.TrimSpace(body.Email)
	if body.Username == "" || body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Username, Email, Password wajib diisi"})
	}

	// Cek duplicate username/email
	var cnt int64
	if err := database.DB.Model(&models.User{}).
		Where("username = ? OR email = ?", body.Username, body.Email).
		Count(&cnt).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal memeriksa duplikasi"})
	}
	if cnt > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Username atau Email sudah terpakai"})
	}

	pwHash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal memproses password"})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membuat user", "error": err.Error()})
	}

	u.PasswordHash = ""
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User berhasil terdaftar",
		"user":    u,
	})
}

func Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Field tidak valid"})
	}
	body.Email = strings.TrimSpace(body.Email)
	if body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Email dan password wajib diisi"})
	}

	// Cari user
	var u models.User
	if err := database.DB.Where("email = ?", body.Email).First(&u).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Kredensial salah"})
	}

	// Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Kredensial salah"})
	}

	accessJWT, err := utils.GenerateAccessToken(c, int(u.ID), u.Username, u.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membuat access token"})
	}

	refreshJWT, err := utils.GenerateRefreshToken(c, int(u.ID), u.Username, u.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membuat refresh token"})
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
		// pastikan single-active: buang key lama dulu
		_ = database.Rdb.Del(context.Background(), fmt.Sprintf("refresh:%d", rt.UserID)).Err()
		if err := utils.SaveRefreshToken(c, rt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menyimpan refresh token (redis)", "error": err.Error()})
		}
	} else {
		_ = database.DB.Where("user_id = ?", rt.UserID).Delete(&models.RefreshToken{}).Error
		if err := database.DB.Create(&rt).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menyimpan refresh token (db)", "error": err.Error()})
		}
	}

	return c.JSON(fiber.Map{
		"message":            "Login sukses",
		"access_token":       accessJWT,
		"access_expires_at":  time.Now().Add(15 * time.Minute).Format(time.RFC3339),
		"refresh_token":      refreshJWT,
		"refresh_expires_at": time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
	})
}

func RefreshToken(c *fiber.Ctx) error {
    rdbCtx := context.Background()

    encRefresh := c.Cookies("refresh_token")
    if encRefresh == "" {
        return c.Status(401).JSON(fiber.Map{"message": "refresh token missing"})
    }

    // decrypt cookie -> plaintext JWT lama yang dikirim klien
    refreshStr, err := utils.Decrypt(encRefresh)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"message": "invalid refresh token1"})
    }

    // parse JWT untuk ambil claims (id, name, role, exp)
    token, err := jwt.Parse(refreshStr, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SECRET")), nil
    })
    if err != nil || !token.Valid {
        return c.Status(401).JSON(fiber.Map{"message": "invalid refresh token2"})
    }

    claims := token.Claims.(jwt.MapClaims)
    userID := int(claims["id"].(float64))

    // ambil state aktif dari Redis
    userKey := fmt.Sprintf("refresh:%d", userID)
    hash, err := database.Rdb.HGetAll(rdbCtx, userKey).Result()
    if err != nil || len(hash) == 0 {
        return c.Status(401).JSON(fiber.Map{"message": "refresh token not found"})
    }

    // decrypt nilai yg tersimpan di Redis (token aktif resmi)
    hashRefreshtoken := hash["refresh_token"]
    refreshDecrypted, err := utils.Decrypt(hashRefreshtoken)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"message": "invalid refresh token3"})
    }

    hashParentToken := hash["parent_token"]
    parentDecrypted := ""
    if hashParentToken != "" {
        parentDecrypted, err = utils.Decrypt(hashParentToken)
        if err != nil {
            return c.Status(401).JSON(fiber.Map{"message": "invalid refresh token4"})
        }
    }

    expInt, _ := strconv.ParseInt(hash["exp"], 10, 64)

    // deteksi reuse / mismatch
    if refreshStr == parentDecrypted && refreshStr != refreshDecrypted {
        database.Rdb.Del(rdbCtx, userKey)
        return c.Status(401).JSON(fiber.Map{"message": "refresh token mismatch, please login again"})
    }

    // expired?
    if time.Now().Unix() > expInt {
        database.Rdb.Del(rdbCtx, userKey)
        return c.Status(401).JSON(fiber.Map{"message": "refresh token expired"})
    }

    // generate access baru
    if _, err := utils.GenerateAccessToken(c, userID, claims["name"].(string), claims["role"].(string)); err != nil {
        return c.Status(500).JSON(fiber.Map{"message": "failed generate access token"})
    }

    // generate refresh baru (JWT -> JWT)
    signedRefresh, err := utils.GenerateRefreshToken(c, userID, claims["name"].(string), claims["role"].(string))
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"message": "Gagal buat refresh token"})
    }

    ipPtr := getIPPtr(c)
    uaPtr := getUserAgentPtr(c)

    // PENTING: parent ambil dari token AKTIF DI REDIS (refreshDecrypted)
    newState := models.RefreshToken{
        RefreshToken: signedRefresh,
        ParentToken:  refreshDecrypted,
        UserID:       userID,
        Exp:          time.Now().Add(7 * 24 * time.Hour).Unix(),
        IPAddress:    ipPtr,
        UserAgent:    uaPtr,
    }

    // simpan ke Redis
    if err := utils.SaveRefreshToken(c, newState); err != nil {
        return c.Status(500).JSON(fiber.Map{"message": "Redis error"})
    }

    // --- DB mirror (hapus jika SaveRefreshToken-mu sudah melakukan ini) ---
    if database.DB != nil {
        // single-active di DB
        _ = database.DB.Where("user_id = ?", newState.UserID).Delete(&models.RefreshToken{}).Error
        if err := database.DB.Create(&newState).Error; err != nil {
            return c.Status(500).JSON(fiber.Map{"message": "DB mirror error", "error": err.Error()})
        }
    }
    // ---------------------------------------------------------------------

    return c.JSON(fiber.Map{
        "message": "refresh success",
    })
}

func Logout(c *fiber.Ctx) error {
	// 1) Ambil refresh token dari cookie atau body
	rtEnc := c.Cookies("refresh_token", "")
	if rtEnc == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.BodyParser(&body)
		rtEnc = strings.TrimSpace(body.RefreshToken)
	}
	if rtEnc == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "refresh_token wajib"})
	}

	// 2) Jika datang dari cookie, itu terenkripsi (utils.SetTokenCookie) → decrypt ke plaintext JWT
	rtPlain, derr := utils.Decrypt(rtEnc)
	if derr != nil {
		// Bisa jadi user memang mengirim plaintext JWT di body, jadi gunakan apa adanya
		rtPlain = rtEnc
	}

	// 3) Parse JWT untuk ambil userID (claim "id"). Jika gagal, kita tetap akan mencoba hapus via DB.
	var userIDInt int
	if token, err := jwt.Parse(rtPlain, func(t *jwt.Token) (interface{}, error) {
		// Pastikan algoritma HMAC
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	}); err == nil && token != nil && token.Valid {
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

	// 4) Invalidate di Redis (hapus key refresh:<userID>) bila userID diketahui & Redis ada
	if database.Rdb != nil && userIDInt != 0 {
		userKey := fmt.Sprintf("refresh:%d", userIDInt)
		_ = database.Rdb.Del(context.Background(), userKey)
	}

	// 5) Fallback/Best-effort: hapus baris di DB berdasarkan nilai refresh_token (plaintext JWT)
	_ = database.DB.Where("refresh_token = ?", rtPlain).Delete(&models.RefreshToken{}).Error

	return c.JSON(fiber.Map{"message": "logout sukses"})
}
