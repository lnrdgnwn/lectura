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

// Get all novels (simple)
func GetNovels(c *fiber.Ctx) error {
	var novels []models.Novel
	if err := database.DB.Preload("Genres").Find(&novels).Error; err != nil {
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

// Create novel (JSON body). optional: genre_ids []uint
func PostNovel(c *fiber.Ctx) error {
	var payload struct {
		AuthorID uint    `json:"author_id"`
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
	if payload.AuthorID == 0 || strings.TrimSpace(payload.Title) == "" || strings.TrimSpace(payload.Slug) == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "author_id, title, dan slug wajib diisi"})
	}

	n := models.Novel{
		AuthorID:   payload.AuthorID,
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

	// jika ada genre ids, preload genres dan set association
	if len(payload.GenreIDs) > 0 {
		var genres []models.Genre
		if err := database.DB.Where("id IN ?", payload.GenreIDs).Find(&genres).Error; err == nil {
			n.Genres = genres
		}
		// jika gagal load genres, kita tetap create novel (opsional: bisa stop)
	}

	if err := database.DB.Create(&n).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membuat novel", "error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "Novel dibuat", "novel": n})
}

// Update novel (partial)
func UpdateNovel(c *fiber.Ctx) error {
	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.Preload("Genres").First(&novel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Novel tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil novel"})
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

	// update genres jika dikirim list genre ids (replace association)
	if payload.GenreIDs != nil {
		var genres []models.Genre
		if err := database.DB.Where("id IN ?", payload.GenreIDs).Find(&genres).Error; err == nil {
			if err := database.DB.Model(&novel).Association("Genres").Replace(&genres); err != nil {
				// ignore error association replace, tapi bisa ditangani bila mau
			}
		}
	}

	novel.UpdatedAt = time.Now()
	if err := database.DB.Save(&novel).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menyimpan perubahan", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Novel diperbarui", "novel": novel})
}

// Delete novel
func DeleteNovel(c *fiber.Ctx) error {
	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.First(&novel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Novel tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil novel"})
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
