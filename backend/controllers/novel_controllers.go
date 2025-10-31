package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"final_project/database"
	"final_project/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ----------------- helpers -----------------

// ambil user id dari c.Locals("user_id")
func getUserIDFromLocals(c *fiber.Ctx) (uint, error) {
	v := c.Locals("user_id")
	if v == nil {
		return 0, fiber.ErrUnauthorized
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
	return 0, fiber.ErrUnauthorized
}

// ambil role dari c.Locals("role")
func getRoleFromLocals(c *fiber.Ctx) string {
	v := c.Locals("role")
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// ----------------- handlers -----------------

// Get all novels (simple)
func GetNovels(c *fiber.Ctx) error {
	var novels []models.Novel
	if err := database.DB.Preload("Genres").Order("created_at DESC").Find(&novels).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil novel"})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"data": novels})
}

// Get single novel by id (params :id)
func GetNovel(c *fiber.Ctx) error {
	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.Preload("Genres").First(&novel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Novel tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil novel"})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"data": novel})
}

// Create novel (JSON body). GenreIDs optional
// NOTE: author diambil dari token (user yang login) — client tidak boleh memasukkan author_id
func PostNovel(c *fiber.Ctx) error {
	uid, err := getUserIDFromLocals(c)
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
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Payload tidak valid"})
	}
	if strings.TrimSpace(payload.Title) == "" || strings.TrimSpace(payload.Slug) == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "title dan slug wajib diisi"})
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

	// attach genres jika ada
	if len(payload.GenreIDs) > 0 {
		var genres []models.Genre
		if err := database.DB.Where("id IN ?", payload.GenreIDs).Find(&genres).Error; err == nil {
			n.Genres = genres
		}
	}

	if err := database.DB.Create(&n).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membuat novel", "error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "Novel dibuat", "novel": n})
}

// Update novel (partial) — hanya author atau admin
func UpdateNovel(c *fiber.Ctx) error {
	uid, err := getUserIDFromLocals(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}
	role := strings.ToUpper(getRoleFromLocals(c))

	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.Preload("Genres").First(&novel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Novel tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil novel"})
	}

	// cek ownership atau admin
	if novel.AuthorID != uid && role != "admin" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "Tidak berwenang mengubah novel ini"})
	}

	var payload struct {
		Title    *string `json:"title"`
		Synopsis *string `json:"synopsis"`
		CoverURL *string `json:"cover_url"`
		Status   *string `json:"status"`
		GenreIDs []uint  `json:"genre_ids"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Payload tidak valid"})
	}

	if payload.Title != nil {
		novel.Title = strings.TrimSpace(*payload.Title)
	}
	if payload.Synopsis != nil {
		novel.Synopsis = payload.Synopsis
	}
	if payload.CoverURL != nil {
		novel.CoverImage = payload.CoverURL
	}
	if payload.Status != nil && *payload.Status != "" {
		novel.Status = *payload.Status
	}

	// update genres (replace association) bila dikirim
	if payload.GenreIDs != nil {
		var genres []models.Genre
		if err := database.DB.Where("id IN ?", payload.GenreIDs).Find(&genres).Error; err == nil {
			if err := database.DB.Model(&novel).Association("Genres").Replace(&genres); err != nil {
				// ignore association error (opsional: log)
			}
		}
	}

	novel.UpdatedAt = time.Now()
	if err := database.DB.Save(&novel).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menyimpan perubahan", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Novel diperbarui", "novel": novel})
}

// Delete novel — hanya author atau admin
func DeleteNovel(c *fiber.Ctx) error {
	uid, err := getUserIDFromLocals(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}
	role := strings.ToUpper(getRoleFromLocals(c))

	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.First(&novel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Novel tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil novel"})
	}

	// cek ownership atau admin
	if novel.AuthorID != uid && role != "admin" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"message": "Tidak berwenang menghapus novel ini"})
	}

	if err := database.DB.Delete(&novel).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menghapus novel"})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Novel dihapus"})
}

// Simple list with search + pagination
// GET /novels?q=kata&page=1&limit=20
func SearchNovels(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q", ""))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	db := database.DB.Model(&models.Novel{}).Preload("Genres")
	if q != "" {
		like := "%" + q + "%"
		db = db.Where("title LIKE ? OR synopsis LIKE ?", like, like)
	}

	var novels []models.Novel
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&novels).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil novel"})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"page": page, "limit": limit, "data": novels})
}
