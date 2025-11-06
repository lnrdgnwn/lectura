// controllers/novel_controllers.go
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

func novelOK(c *fiber.Ctx, status int, msg string, data any) error {
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

func novelListOK(c *fiber.Ctx, msg string, data any, page, limit int, total int64) error {
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

func novelFail(c *fiber.Ctx, status int, msg string, err error) error {
	resp := fiber.Map{
		"success": false,
		"message": msg,
	}
	if err != nil && os.Getenv("APP_ENV") != "production" {
		resp["error"] = err.Error()
	}
	return c.Status(status).JSON(resp)
}

func getAuthFromAccessCookieNovel(c *fiber.Ctx) (uint, string, error) {
	enc := c.Cookies("access_token", "")
	if enc == "" {
		return 0, "", fiber.ErrUnauthorized
	}

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
		} else {
			return 0, "", fiber.ErrUnauthorized
		}
	default:
		return 0, "", fiber.ErrUnauthorized
	}

	role := ""
	if r, ok := claims["role"].(string); ok {
		role = strings.ToLower(r)
	}

	return uid, role, nil
}

func normalizeNovelStatus(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "ongoing", "complete", "hiatus":
		return s, true
	default:
		return "", false
	}
}

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

func GetNovel(c *fiber.Ctx) error {
	var novels []models.Novel
	if err := database.DB.Preload("Genres").Preload("Tags").
		Order("created_at DESC").Find(&novels).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel", err)
	}
	return novelOK(c, http.StatusOK, "Daftar novel", novels)
}

func GetNovelByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var novel models.Novel
	if err := database.DB.
		Preload("Genres").
		Preload("Tags").
		First(&novel, id).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return novelFail(c, http.StatusNotFound, "Novel tidak ditemukan", nil)
		}
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel", err)
	}

	uid, role, _ := getAuthFromAccessCookieNovel(c) 

	var chapters []models.Chapter

	chQuery := database.DB.
		Model(&models.Chapter{}).
		Where("novel_id = ?", novel.ID).
		Order("order_no ASC")

	if role == "admin" || uid == novel.AuthorID {
		if err := chQuery.Find(&chapters).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil daftar chapter", err)
		}
	} else {
		if err := chQuery.
			Where("published_at IS NOT NULL AND published_at <= ?", time.Now()).
			Find(&chapters).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil daftar chapter terbit", err)
		}
	}

	return novelOK(c, http.StatusOK, "Detail novel", fiber.Map{
		"novel":    novel,
		"chapters": chapters,
	})
}

func GetMyNovels(c *fiber.Ctx) error {
	uid, _, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return novelFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	q := strings.TrimSpace(c.Query("q", ""))
	status := strings.TrimSpace(c.Query("status", ""))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	countQB := database.DB.Model(&models.Novel{}).Where("author_id = ?", uid)

	if q != "" {
		like := "%" + q + "%"
		countQB = countQB.Where("title LIKE ? OR slug LIKE ?", like, like)
	}

	if status != "" {
		if ns, ok := normalizeNovelStatus(status); ok {
			countQB = countQB.Where("status = ?", ns)
		} else {
			return novelFail(c, http.StatusBadRequest, "Status tidak valid (gunakan: ongoing, complete, hiatus)", nil)
		}
	}

	var total int64
	if err := countQB.Count(&total).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal menghitung total novel milik user", err)
	}

	dataQB := database.DB.
		Model(&models.Novel{}).
		Where("author_id = ?", uid).
		Preload("Genres").
		Preload("Tags")

	if q != "" {
		like := "%" + q + "%"
		dataQB = dataQB.Where("title LIKE ? OR slug LIKE ?", like, like)
	}
	if status != "" {
		ns, _ := normalizeNovelStatus(status)
		dataQB = dataQB.Where("status = ?", ns)
	}

	var novels []models.Novel
	if err := dataQB.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&novels).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel milik user", err)
	}

	return novelListOK(c, "Daftar novel milik user", novels, page, limit, total)
}

func PostNovel(c *fiber.Ctx) error {
	uid, _, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return novelFail(c, http.StatusUnauthorized, "Unauthorized", err)
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
			return novelFail(c, http.StatusBadRequest, "Payload JSON tidak valid", err)
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
			return novelFail(c, http.StatusBadRequest, "genre_ids tidak valid", perr)
		} else if present {
			payload.GenreIDs = arr
		}
		if arr, present, perr := parseIDsFromForm(c, "tag_ids", "tag_ids[]"); perr != nil {
			return novelFail(c, http.StatusBadRequest, "tag_ids tidak valid", perr)
		} else if present {
			payload.TagIDs = arr
		}
		if _, ferr := c.FormFile("cover_image"); ferr == nil {
			if path, err := utils.SaveFile(c, "cover_image", "cover"); err == nil {
				payload.CoverURL = &path
			} else {
				return novelFail(c, http.StatusBadRequest, "Gagal menyimpan cover", err)
			}
		}
	}

	if strings.TrimSpace(payload.Title) == "" || strings.TrimSpace(payload.Slug) == "" {
		return novelFail(c, http.StatusBadRequest, "title dan slug wajib diisi", nil)
	}
	if len(payload.GenreIDs) > 1 {
		return novelFail(c, http.StatusBadRequest, "hanya boleh memasukkan maksimal 1 genre", nil)
	}

	status := "ongoing"
	if payload.Status != nil && strings.TrimSpace(*payload.Status) != "" {
		if ns, ok := normalizeNovelStatus(*payload.Status); ok {
			status = ns
		} else {
			return novelFail(c, http.StatusBadRequest, "Status tidak valid (gunakan: ongoing, complete, hiatus)", nil)
		}
	}

	n := models.Novel{
		AuthorID:   uid,
		Title:      strings.TrimSpace(payload.Title),
		Slug:       strings.TrimSpace(payload.Slug),
		Synopsis:   payload.Synopsis,
		CoverImage: payload.CoverURL,
		Status:     status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if len(payload.GenreIDs) == 1 {
		var genre models.Genre
		if err := database.DB.First(&genre, payload.GenreIDs[0]).Error; err != nil {
			return novelFail(c, http.StatusBadRequest, "genre tidak ditemukan", err)
		}
		n.Genres = []models.Genre{genre}
	}

	if len(payload.TagIDs) > 0 {
		var tags []models.Tag
		if err := database.DB.Where("id IN ?", payload.TagIDs).Find(&tags).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil tags", err)
		}
		if len(tags) != len(payload.TagIDs) {
			return novelFail(c, http.StatusBadRequest, "salah satu tag_id tidak ditemukan", nil)
		}
		n.Tags = tags
	}

	if err := database.DB.Create(&n).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal membuat novel", err)
	}
	_ = database.DB.Preload("Genres").Preload("Tags").First(&n, n.ID)

	return novelOK(c, http.StatusCreated, "Novel dibuat", n)
}

func UpdateNovel(c *fiber.Ctx) error {
	uid, role, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return novelFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.Preload("Genres").Preload("Tags").First(&novel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return novelFail(c, http.StatusNotFound, "Novel tidak ditemukan", nil)
		}
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel", err)
	}

	if novel.AuthorID != uid && role != "admin" {
		return novelFail(c, http.StatusForbidden, "Tidak berwenang mengubah novel ini", nil)
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
			return novelFail(c, http.StatusBadRequest, "Payload JSON tidak valid", err)
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
			return novelFail(c, http.StatusBadRequest, "genre_ids tidak valid", perr)
		} else if present {
			payload.GenreIDs = &arr
		}
		if arr, present, perr := parseIDsFromForm(c, "tag_ids", "tag_ids[]"); perr != nil {
			return novelFail(c, http.StatusBadRequest, "tag_ids tidak valid", perr)
		} else if present {
			payload.TagIDs = &arr
		}
		if _, ferr := c.FormFile("cover_image"); ferr == nil {
			if path, err := utils.SaveFile(c, "cover_image", "cover"); err == nil {
				payload.CoverImage = &path
			} else {
				return novelFail(c, http.StatusBadRequest, "Gagal menyimpan cover", err)
			}
		}
	}

	if payload.GenreIDs != nil && len(*payload.GenreIDs) > 1 {
		return novelFail(c, http.StatusBadRequest, "hanya boleh memasukkan maksimal 1 genre", nil)
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
	if payload.Status != nil {
		s := strings.TrimSpace(*payload.Status)
		if s == "" {
			return novelFail(c, http.StatusBadRequest, "Status tidak boleh kosong", nil)
		}
		if ns, ok := normalizeNovelStatus(s); ok {
			novel.Status = ns
		} else {
			return novelFail(c, http.StatusBadRequest, "Status tidak valid (gunakan: ongoing, complete, hiatus)", nil)
		}
	}

	if payload.GenreIDs != nil {
		if len(*payload.GenreIDs) == 0 {
			_ = database.DB.Model(&novel).Association("Genres").Clear()
		} else {
			var genres []models.Genre
			if err := database.DB.Where("id IN ?", *payload.GenreIDs).Find(&genres).Error; err != nil {
				return novelFail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
			}
			if len(genres) != len(*payload.GenreIDs) {
				return novelFail(c, http.StatusBadRequest, "salah satu genre_id tidak ditemukan", nil)
			}
			_ = database.DB.Model(&novel).Association("Genres").Replace(&genres)
		}
	}

	if payload.TagIDs != nil {
		if len(*payload.TagIDs) == 0 {
			_ = database.DB.Model(&novel).Association("Tags").Clear()
		} else {
			var tags []models.Tag
			if err := database.DB.Where("id IN ?", *payload.TagIDs).Find(&tags).Error; err != nil {
				return novelFail(c, http.StatusInternalServerError, "Gagal mengambil tags", err)
			}
			if len(tags) != len(*payload.TagIDs) {
				return novelFail(c, http.StatusBadRequest, "salah satu tag_id tidak ditemukan", nil)
			}
			_ = database.DB.Model(&novel).Association("Tags").Replace(&tags)
		}
	}

	novel.UpdatedAt = time.Now()
	if err := database.DB.Save(&novel).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal menyimpan perubahan", err)
	}
	_ = database.DB.Preload("Genres").Preload("Tags").First(&novel, novel.ID)

	return novelOK(c, http.StatusOK, "Novel diperbarui", novel)
}

func DeleteNovel(c *fiber.Ctx) error {
	uid, role, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return novelFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.First(&novel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return novelFail(c, http.StatusNotFound, "Novel tidak ditemukan", nil)
		}
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel", err)
	}

	if novel.AuthorID != uid && role != "admin" {
		return novelFail(c, http.StatusForbidden, "Tidak berwenang menghapus novel ini", nil)
	}

	if err := database.DB.Delete(&novel).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal menghapus novel", err)
	}
	return novelOK(c, http.StatusOK, "Novel dihapus", nil)
}

func SearchNovels(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q", ""))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	countQB := database.DB.Model(&models.Novel{})
	if q != "" {
		like := "%" + q + "%"
		countQB = countQB.Where("title LIKE ? OR slug LIKE ?", like, like)
	}

	var total int64
	if err := countQB.Count(&total).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal menghitung total novel", err)
	}

	dataQB := database.DB.Model(&models.Novel{}).Preload("Genres").Preload("Tags")
	if q != "" {
		like := "%" + q + "%"
		dataQB = dataQB.Where("title LIKE ? OR slug LIKE ?", like, like)
	}

	var novels []models.Novel
	if err := dataQB.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&novels).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel", err)
	}

	return novelListOK(c, "Hasil pencarian novel", novels, page, limit, total)
}
