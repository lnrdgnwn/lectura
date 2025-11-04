package controllers

import (
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"final_project/database"
	"final_project/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func tagOK(c *fiber.Ctx, status int, msg string, data any) error {
	resp := fiber.Map{
		"success": true,
		"message": msg,
	}
	if data != nil {
		resp["data"] = data
	}
	return c.Status(status).JSON(resp)
}

func tagOKList(c *fiber.Ctx, msg string, data any, page, limit int, total int64) error {
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

func tagFail(c *fiber.Ctx, status int, msg string, err error) error {
	resp := fiber.Map{
		"success": false,
		"message": msg,
	}
	if err != nil && os.Getenv("APP_ENV") != "production" {
		resp["error"] = err.Error()
	}
	return c.Status(status).JSON(resp)
}

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
	countDB := database.DB.Model(&models.Tag{})

	if q != "" {
		like := "%" + q + "%"
		db = db.Where("name LIKE ? OR slug LIKE ?", like, like)
		countDB = countDB.Where("name LIKE ? OR slug LIKE ?", like, like)
	}

	var total int64
	if err := countDB.Count(&total).Error; err != nil {
		return tagFail(c, http.StatusInternalServerError, "Gagal menghitung total tag", err)
	}

	var tags []models.Tag
	if err := db.Order("name ASC").Limit(limit).Offset(offset).Find(&tags).Error; err != nil {
		return tagFail(c, http.StatusInternalServerError, "Gagal mengambil daftar tag", err)
	}

	return tagOKList(c, "Daftar tag berhasil diambil", tags, page, limit, total)
}

func GetTag(c *fiber.Ctx) error {
	param := c.Params("id")
	var t models.Tag

	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&t, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return tagFail(c, http.StatusNotFound, "Tag tidak ditemukan", nil)
			}
			return tagFail(c, http.StatusInternalServerError, "Gagal mengambil tag", err)
		}
		return tagOK(c, http.StatusOK, "Detail tag berhasil diambil", t)
	}

	if err := database.DB.Where("slug = ?", param).First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return tagFail(c, http.StatusNotFound, "Tag tidak ditemukan", nil)
		}
		return tagFail(c, http.StatusInternalServerError, "Gagal mengambil tag", err)
	}
	return tagOK(c, http.StatusOK, "Detail tag berhasil diambil", t)
}

func CreateTag(c *fiber.Ctx) error {
	var payload struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return tagFail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}
	payload.Name = strings.TrimSpace(payload.Name)
	payload.Slug = strings.TrimSpace(payload.Slug)

	if payload.Name == "" {
		return tagFail(c, http.StatusBadRequest, "Field name wajib diisi", nil)
	}
	if payload.Slug == "" {
		payload.Slug = generateSlug(payload.Name)
	} else {
		payload.Slug = generateSlug(payload.Slug)
	}

	var cnt int64
	if err := database.DB.Model(&models.Tag{}).
		Where("name = ? OR slug = ?", payload.Name, payload.Slug).
		Count(&cnt).Error; err != nil {
		return tagFail(c, http.StatusInternalServerError, "Gagal memeriksa duplikasi tag", err)
	}
	if cnt > 0 {
		return tagFail(c, http.StatusConflict, "Name atau slug sudah digunakan", nil)
	}

	tag := models.Tag{
		Name:      payload.Name,
		Slug:      payload.Slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := database.DB.Create(&tag).Error; err != nil {
		return tagFail(c, http.StatusInternalServerError, "Gagal membuat tag", err)
	}
	return tagOK(c, http.StatusCreated, "Tag berhasil dibuat", tag)
}

func UpdateTag(c *fiber.Ctx) error {
	param := c.Params("id")
	var t models.Tag

	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&t, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return tagFail(c, http.StatusNotFound, "Tag tidak ditemukan", nil)
			}
			return tagFail(c, http.StatusInternalServerError, "Gagal mengambil tag", err)
		}
	} else {
		if err := database.DB.Where("slug = ?", param).First(&t).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return tagFail(c, http.StatusNotFound, "Tag tidak ditemukan", nil)
			}
			return tagFail(c, http.StatusInternalServerError, "Gagal mengambil tag", err)
		}
	}

	var payload struct {
		Name *string `json:"name"`
		Slug *string `json:"slug"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return tagFail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}

	if payload.Name != nil {
		newName := strings.TrimSpace(*payload.Name)
		if newName == "" {
			return tagFail(c, http.StatusBadRequest, "Field name tidak boleh kosong", nil)
		}
		var cnt int64
		if err := database.DB.Model(&models.Tag{}).
			Where("name = ? AND id <> ?", newName, t.ID).
			Count(&cnt).Error; err != nil {
			return tagFail(c, http.StatusInternalServerError, "Gagal memeriksa duplikasi name", err)
		}
		if cnt > 0 {
			return tagFail(c, http.StatusConflict, "Name sudah digunakan", nil)
		}
		t.Name = newName
	}
	if payload.Slug != nil {
		newSlug := generateSlug(strings.TrimSpace(*payload.Slug))
		if newSlug == "" {
			return tagFail(c, http.StatusBadRequest, "Field slug tidak boleh kosong", nil)
		}
		var cnt int64
		if err := database.DB.Model(&models.Tag{}).
			Where("slug = ? AND id <> ?", newSlug, t.ID).
			Count(&cnt).Error; err != nil {
			return tagFail(c, http.StatusInternalServerError, "Gagal memeriksa duplikasi slug", err)
		}
		if cnt > 0 {
			return tagFail(c, http.StatusConflict, "Slug sudah digunakan", nil)
		}
		t.Slug = newSlug
	}

	t.UpdatedAt = time.Now()
	if err := database.DB.Save(&t).Error; err != nil {
		return tagFail(c, http.StatusInternalServerError, "Gagal memperbarui tag", err)
	}
	return tagOK(c, http.StatusOK, "Tag berhasil diperbarui", t)
}

func DeleteTag(c *fiber.Ctx) error {
	param := c.Params("id")
	var t models.Tag

	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&t, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return tagFail(c, http.StatusNotFound, "Tag tidak ditemukan", nil)
			}
			return tagFail(c, http.StatusInternalServerError, "Gagal mengambil tag", err)
		}
	} else {
		if err := database.DB.Where("slug = ?", param).First(&t).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return tagFail(c, http.StatusNotFound, "Tag tidak ditemukan", nil)
			}
			return tagFail(c, http.StatusInternalServerError, "Gagal mengambil tag", err)
		}
	}

	if err := database.DB.Delete(&t).Error; err != nil {
		return tagFail(c, http.StatusInternalServerError, "Gagal menghapus tag", err)
	}
	return tagOK(c, http.StatusOK, "Tag berhasil dihapus", nil)
}
