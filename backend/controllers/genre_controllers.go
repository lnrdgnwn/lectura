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

// ListGenres - GET /genres?q=&page=&limit=
// Mendukung pencarian sederhana pada name/slug dan pagination
func ListGenres(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q", ""))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	offset := (page - 1) * limit

	db := database.DB.Model(&models.Genre{})
	if q != "" {
		like := "%" + q + "%"
		db = db.Where("name LIKE ? OR slug LIKE ?", like, like)
	}

	var genres []models.Genre
	if err := db.Order("name ASC").Limit(limit).Offset(offset).Find(&genres).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil daftar genre", "error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"page":  page,
		"limit": limit,
		"data":  genres,
	})
}

// GetGenre - GET /genres/:id
// :id dapat berupa numeric id atau slug
func GetGenre(c *fiber.Ctx) error {
	param := c.Params("id")
	var g models.Genre

	// coba numeric id dulu
	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&g, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "genre tidak ditemukan"})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil genre", "error": err.Error()})
		}
		return c.Status(http.StatusOK).JSON(fiber.Map{"data": g})
	}

	// fallback: cari berdasarkan slug
	if err := database.DB.Where("slug = ?", param).First(&g).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "genre tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil genre", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"data": g})
}

// CreateGenre - POST /genres
// body JSON: { "name": "Romance", "slug": "romance" }
func CreateGenre(c *fiber.Ctx) error {
	var payload struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "payload tidak valid"})
	}
	payload.Name = strings.TrimSpace(payload.Name)
	payload.Slug = strings.TrimSpace(payload.Slug)

	if payload.Name == "" || payload.Slug == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "name dan slug wajib diisi"})
	}

	// cek duplicate name/slug
	var cnt int64
	database.DB.Model(&models.Genre{}).
		Where("name = ? OR slug = ?", payload.Name, payload.Slug).
		Count(&cnt)
	if cnt > 0 {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"message": "name atau slug sudah digunakan"})
	}

	g := models.Genre{
		Name:      payload.Name,
		Slug:      payload.Slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := database.DB.Create(&g).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal membuat genre", "error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "genre dibuat", "genre": g})
}

// UpdateGenre - PUT /genres/:id
// body JSON: { "name": "...", "slug": "..." } (partial allowed)
func UpdateGenre(c *fiber.Ctx) error {
	param := c.Params("id")
	var g models.Genre

	// ambil genre (by id or slug)
	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&g, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "genre tidak ditemukan"})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil genre", "error": err.Error()})
		}
	} else {
		if err := database.DB.Where("slug = ?", param).First(&g).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "genre tidak ditemukan"})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil genre", "error": err.Error()})
		}
	}

	var payload struct {
		Name *string `json:"name"`
		Slug *string `json:"slug"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "payload tidak valid"})
	}

	// jika mengubah name/slug, cek unique (kecuali milik record ini)
	if payload.Name != nil {
		newName := strings.TrimSpace(*payload.Name)
		if newName == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "name tidak boleh kosong"})
		}
		var cnt int64
		database.DB.Model(&models.Genre{}).Where("name = ? AND id <> ?", newName, g.ID).Count(&cnt)
		if cnt > 0 {
			return c.Status(http.StatusConflict).JSON(fiber.Map{"message": "name sudah digunakan"})
		}
		g.Name = newName
	}
	if payload.Slug != nil {
		newSlug := strings.TrimSpace(*payload.Slug)
		if newSlug == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "slug tidak boleh kosong"})
		}
		var cnt int64
		database.DB.Model(&models.Genre{}).Where("slug = ? AND id <> ?", newSlug, g.ID).Count(&cnt)
		if cnt > 0 {
			return c.Status(http.StatusConflict).JSON(fiber.Map{"message": "slug sudah digunakan"})
		}
		g.Slug = newSlug
	}

	g.UpdatedAt = time.Now()
	if err := database.DB.Save(&g).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal memperbarui genre", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "genre diperbarui", "genre": g})
}

// DeleteGenre - DELETE /genres/:id
func DeleteGenre(c *fiber.Ctx) error {
	param := c.Params("id")
	var g models.Genre

	// ambil genre
	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&g, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "genre tidak ditemukan"})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil genre", "error": err.Error()})
		}
	} else {
		if err := database.DB.Where("slug = ?", param).First(&g).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "genre tidak ditemukan"})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil genre", "error": err.Error()})
		}
	}

	if err := database.DB.Delete(&g).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menghapus genre", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "genre dihapus"})
}
