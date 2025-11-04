package controllers

import (
	"encoding/json"
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
	"gorm.io/gorm"
)

func getAuthFromAccessCookieNovel(c *fiber.Ctx) (uint, string, error) {
	enc := c.Cookies("access_token", "")
	if enc == "" {
		return 0, "", fiber.ErrUnauthorized
	}
	// decrypt cookie -> plaintext JWT
	tokenStr, err := utils.Decrypt(enc)
	if err != nil || strings.TrimSpace(tokenStr) == "" {
		return 0, "", fiber.ErrUnauthorized
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return 0, "", fiber.ErrUnauthorized
	}
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(secret), nil
	})
	if err != nil || tok == nil || !tok.Valid {
		return 0, "", fiber.ErrUnauthorized
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", fiber.ErrUnauthorized
	}

	var uid uint
	switch v := claims["id"].(type) {
	case float64:
		uid = uint(v)
	case string:
		if n, e := strconv.Atoi(v); e == nil {
			uid = uint(n)
		}
	default:
		return 0, "", fiber.ErrUnauthorized
	}

	role := ""
	if r, ok := claims["role"].(string); ok {
		role = r
	}

	return uid, strings.ToUpper(role), nil
}

/* ===================== helpers parsing kecil ===================== */

// parse comma-separated "1,2,3" or JSON array "[1,2]" or single "1"
func parseUintSliceFromString(s string) ([]uint, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return []uint{}, nil
	}
	if strings.HasPrefix(s, "[") {
		var arr []uint
		if err := json.Unmarshal([]byte(s), &arr); err == nil {
			return arr, nil
		}
	}
	parts := strings.Split(s, ",")
	out := make([]uint, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		out = append(out, uint(n))
	}
	return out, nil
}

// try multiple form keys for arrays: "genre_ids", "genre_ids[]"
func parseIDsFromForm(c *fiber.Ctx, keys ...string) ([]uint, bool, error) {
	for _, k := range keys {
		v := c.FormValue(k)
		if v != "" {
			arr, err := parseUintSliceFromString(v)
			return arr, true, err
		}
	}
	return nil, false, nil
}

/* ===================== handlers ===================== */

// Get all novels (public)
func GetNovels(c *fiber.Ctx) error {
	var novels []models.Novel
	if err := database.DB.Preload("Genres").Preload("Tags").
		Order("created_at DESC").Find(&novels).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil novel"})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"data": novels})
}

// Get single novel by id (public)
func GetNovel(c *fiber.Ctx) error {
	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.Preload("Genres").Preload("Tags").First(&novel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Novel tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil novel"})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"data": novel})
}

// Create novel (author must be authenticated via cookie access_token)
func PostNovel(c *fiber.Ctx) error {
	uid, _, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	var payload struct {
		Title    string  `json:"title"`
		Slug     string  `json:"slug"`
		Synopsis *string `json:"synopsis"`
		CoverURL *string `json:"cover_url"`
		Status   *string `json:"status"`
		GenreIDs []uint  `json:"genre_ids"`
		TagIDs   []uint  `json:"tag_ids"`
	}

	ct := c.Get("Content-Type")
	if strings.Contains(ct, "application/json") || c.Is("json") {
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Payload JSON tidak valid"})
		}
	} else {
		payload.Title = c.FormValue("title")
		payload.Slug = c.FormValue("slug")
		if s := c.FormValue("synopsis"); s != "" {
			payload.Synopsis = &s
		}
		if cv := c.FormValue("cover_image"); cv != "" {
			payload.CoverURL = &cv
		}
		if st := c.FormValue("status"); st != "" {
			payload.Status = &st
		}
		if arr, present, perr := parseIDsFromForm(c, "genre_ids", "genre_ids[]"); perr != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "genre_ids tidak valid", "error": perr.Error()})
		} else if present {
			payload.GenreIDs = arr
		}
		if arr, present, perr := parseIDsFromForm(c, "tag_ids", "tag_ids[]"); perr != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "tag_ids tidak valid", "error": perr.Error()})
		} else if present {
			payload.TagIDs = arr
		}
		if _, ferr := c.FormFile("cover_image"); ferr == nil {
			if path, err := utils.SaveFile(c, "cover_image", "cover"); err == nil {
				payload.CoverURL = &path
			} else {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "gagal menyimpan cover", "error": err.Error()})
			}
		}
	}

	if strings.TrimSpace(payload.Title) == "" || strings.TrimSpace(payload.Slug) == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "title dan slug wajib diisi"})
	}
	if len(payload.GenreIDs) > 1 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "hanya boleh memasukkan maksimal 1 genre"})
	}

	n := models.Novel{
		AuthorID:   uid,
		Title:      strings.TrimSpace(payload.Title),
		Slug:       strings.TrimSpace(payload.Slug),
		Synopsis:   payload.Synopsis,
		CoverImage: payload.CoverURL,
		Status:     "DRAFT",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if payload.Status != nil && *payload.Status != "" {
		n.Status = *payload.Status
	}

	if len(payload.GenreIDs) == 1 {
		var genre models.Genre
		if err := database.DB.First(&genre, payload.GenreIDs[0]).Error; err == nil {
			n.Genres = []models.Genre{genre}
		} else {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "genre tidak ditemukan", "genre_id": payload.GenreIDs[0]})
		}
	}
	if len(payload.TagIDs) > 0 {
		var tags []models.Tag
		if err := database.DB.Where("id IN ?", payload.TagIDs).Find(&tags).Error; err == nil {
			if len(tags) != len(payload.TagIDs) {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "salah satu tag_id tidak ditemukan"})
			}
			n.Tags = tags
		} else {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil tags", "error": err.Error()})
		}
	}

	if err := database.DB.Create(&n).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membuat novel", "error": err.Error()})
	}
	_ = database.DB.Preload("Genres").Preload("Tags").First(&n, n.ID)
	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "Novel dibuat", "novel": n})
}

// Update novel (author sendiri atau ADMIN) — auth via cookie
func UpdateNovel(c *fiber.Ctx) error {
	uid, role, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.Preload("Genres").Preload("Tags").First(&novel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Novel tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil novel"})
	}

	// cek ownership atau admin
	if novel.AuthorID != uid && role != "ADMIN" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "Tidak berwenang mengubah novel ini"})
	}

	var payload struct {
		Title      *string `json:"title"`
		Synopsis   *string `json:"synopsis"`
		CoverImage *string `json:"cover_image"`
		Status     *string `json:"status"`
		GenreIDs   *[]uint `json:"genre_ids"`
		TagIDs     *[]uint `json:"tag_ids"`
	}

	ct := c.Get("Content-Type")
	if strings.Contains(ct, "application/json") || c.Is("json") {
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Payload JSON tidak valid"})
		}
	} else {
		if v := c.FormValue("title"); v != "" {
			payload.Title = &v
		}
		if v := c.FormValue("synopsis"); v != "" {
			payload.Synopsis = &v
		}
		if v := c.FormValue("cover_image"); v != "" {
			payload.CoverImage = &v
		}
		if v := c.FormValue("status"); v != "" {
			payload.Status = &v
		}
		if arr, present, perr := parseIDsFromForm(c, "genre_ids", "genre_ids[]"); perr != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "genre_ids tidak valid", "error": perr.Error()})
		} else if present {
			payload.GenreIDs = &arr
		}
		if arr, present, perr := parseIDsFromForm(c, "tag_ids", "tag_ids[]"); perr != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "tag_ids tidak valid", "error": perr.Error()})
		} else if present {
			payload.TagIDs = &arr
		}
		if _, ferr := c.FormFile("cover_image"); ferr == nil {
			if path, err := utils.SaveFile(c, "cover_image", "cover"); err == nil {
				payload.CoverImage = &path
			} else {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "gagal menyimpan cover", "error": err.Error()})
			}
		}
	}

	if payload.GenreIDs != nil && len(*payload.GenreIDs) > 1 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "hanya boleh memasukkan maksimal 1 genre"})
	}

	if payload.Title != nil {
		novel.Title = strings.TrimSpace(*payload.Title)
	}
	if payload.Synopsis != nil {
		novel.Synopsis = payload.Synopsis
	}
	if payload.CoverImage != nil {
		novel.CoverImage = payload.CoverImage
	}
	if payload.Status != nil && *payload.Status != "" {
		novel.Status = *payload.Status
	}

	// Genre
	if payload.GenreIDs != nil {
		if len(*payload.GenreIDs) == 0 {
			_ = database.DB.Model(&novel).Association("Genres").Clear()
		} else {
			var genres []models.Genre
			if err := database.DB.Where("id IN ?", *payload.GenreIDs).Find(&genres).Error; err == nil {
				if len(genres) != len(*payload.GenreIDs) {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "salah satu genre_id tidak ditemukan"})
				}
				_ = database.DB.Model(&novel).Association("Genres").Replace(&genres)
			} else {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil genre", "error": err.Error()})
			}
		}
	}

	// Tags
	if payload.TagIDs != nil {
		if len(*payload.TagIDs) == 0 {
			_ = database.DB.Model(&novel).Association("Tags").Clear()
		} else {
			var tags []models.Tag
			if err := database.DB.Where("id IN ?", *payload.TagIDs).Find(&tags).Error; err == nil {
				if len(tags) != len(*payload.TagIDs) {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "salah satu tag_id tidak ditemukan"})
				}
				_ = database.DB.Model(&novel).Association("Tags").Replace(&tags)
			} else {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil tags", "error": err.Error()})
			}
		}
	}

	novel.UpdatedAt = time.Now()
	if err := database.DB.Save(&novel).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menyimpan perubahan", "error": err.Error()})
	}
	_ = database.DB.Preload("Genres").Preload("Tags").First(&novel, novel.ID)

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Novel diperbarui", "novel": novel})
}

// Delete novel — author sendiri atau ADMIN — auth via cookie
func DeleteNovel(c *fiber.Ctx) error {
	uid, role, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.First(&novel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Novel tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil novel"})
	}

	if novel.AuthorID != uid && role != "ADMIN" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "Tidak berwenang menghapus novel ini"})
	}

	if err := database.DB.Delete(&novel).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menghapus novel"})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Novel dihapus"})
}

// Simple list with search + pagination (public)
// GET /novels?q=kata&page=1&limit=20
func SearchNovels(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q", ""))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	db := database.DB.Model(&models.Novel{}).Preload("Genres").Preload("Tags")
	if q != "" {
		like := "%" + q + "%"
		db = db.Where("title LIKE ? OR slug LIKE ?", like, like)
	}

	var novels []models.Novel
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&novels).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil novel"})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"page": page, "limit": limit, "data": novels})
}
