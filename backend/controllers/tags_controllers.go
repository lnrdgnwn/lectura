package controllers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"final_project/database"
	"final_project/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// generateSlug sederhana
func generateSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	re2 := regexp.MustCompile(`-+`)
	s = re2.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 100 {
		s = s[:100]
		s = strings.Trim(s, "-")
	}
	if s == "" {
		return "tag"
	}
	return s
}

// ListTags - GET /tags?q=&page=&limit=
func ListTags(c *fiber.Ctx) error {
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

	db := database.DB.Model(&models.Tag{})
	if q != "" {
		like := "%" + q + "%"
		db = db.Where("name LIKE ? OR slug LIKE ?", like, like)
	}

	var tags []models.Tag
	if err := db.Order("name ASC").Limit(limit).Offset(offset).Find(&tags).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil daftar tag", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"page": page, "limit": limit, "data": tags})
}

// GetTag - GET /tags/:id (id numeric atau slug)
func GetTag(c *fiber.Ctx) error {
	param := c.Params("id")
	var t models.Tag

	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&t, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "tag tidak ditemukan"})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil tag", "error": err.Error()})
		}
		return c.Status(http.StatusOK).JSON(fiber.Map{"data": t})
	}

	if err := database.DB.Where("slug = ?", param).First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "tag tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil tag", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"data": t})
}

// CreateTag - POST /tags  (admin)
func CreateTag(c *fiber.Ctx) error {
	var payload struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "payload tidak valid"})
	}
	payload.Name = strings.TrimSpace(payload.Name)
	payload.Slug = strings.TrimSpace(payload.Slug)

	if payload.Name == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "name wajib diisi"})
	}
	if payload.Slug == "" {
		payload.Slug = generateSlug(payload.Name)
	} else {
		payload.Slug = generateSlug(payload.Slug)
	}

	// unik check
	var cnt int64
	database.DB.Model(&models.Tag{}).Where("name = ? OR slug = ?", payload.Name, payload.Slug).Count(&cnt)
	if cnt > 0 {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"message": "name atau slug sudah digunakan"})
	}

	tag := models.Tag{
		Name:      payload.Name,
		Slug:      payload.Slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := database.DB.Create(&tag).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal membuat tag", "error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "tag dibuat", "tag": tag})
}

// UpdateTag - PUT /tags/:id  (admin)
func UpdateTag(c *fiber.Ctx) error {
	param := c.Params("id")
	var t models.Tag

	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&t, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "tag tidak ditemukan"})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil tag", "error": err.Error()})
		}
	} else {
		if err := database.DB.Where("slug = ?", param).First(&t).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "tag tidak ditemukan"})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil tag", "error": err.Error()})
		}
	}

	var payload struct {
		Name *string `json:"name"`
		Slug *string `json:"slug"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "payload tidak valid"})
	}

	if payload.Name != nil {
		newName := strings.TrimSpace(*payload.Name)
		if newName == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "name tidak boleh kosong"})
		}
		var cnt int64
		database.DB.Model(&models.Tag{}).Where("name = ? AND id <> ?", newName, t.ID).Count(&cnt)
		if cnt > 0 {
			return c.Status(http.StatusConflict).JSON(fiber.Map{"message": "name sudah digunakan"})
		}
		t.Name = newName
	}
	if payload.Slug != nil {
		newSlug := generateSlug(strings.TrimSpace(*payload.Slug))
		if newSlug == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "slug tidak boleh kosong"})
		}
		var cnt int64
		database.DB.Model(&models.Tag{}).Where("slug = ? AND id <> ?", newSlug, t.ID).Count(&cnt)
		if cnt > 0 {
			return c.Status(http.StatusConflict).JSON(fiber.Map{"message": "slug sudah digunakan"})
		}
		t.Slug = newSlug
	}

	t.UpdatedAt = time.Now()
	if err := database.DB.Save(&t).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal memperbarui tag", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "tag diperbarui", "tag": t})
}

// DeleteTag - DELETE /tags/:id  (admin)
func DeleteTag(c *fiber.Ctx) error {
	param := c.Params("id")
	var t models.Tag

	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&t, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "tag tidak ditemukan"})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil tag", "error": err.Error()})
		}
	} else {
		if err := database.DB.Where("slug = ?", param).First(&t).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "tag tidak ditemukan"})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil tag", "error": err.Error()})
		}
	}

	if err := database.DB.Delete(&t).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menghapus tag", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "tag dihapus"})
}

