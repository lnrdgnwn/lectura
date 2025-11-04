package controllers

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"final_project/database"
	"final_project/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)


func ok(c *fiber.Ctx, status int, msg string, data any) error {
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

func okList(c *fiber.Ctx, msg string, data any, page, limit int, total int64) error {
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

func fail(c *fiber.Ctx, status int, msg string, err error) error {
	resp := fiber.Map{
		"success": false,
		"message": msg,
	}
	if err != nil && os.Getenv("APP_ENV") != "production" {
		resp["error"] = err.Error()
	}
	return c.Status(status).JSON(resp)
}

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

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "Gagal menghitung total genre", err)
	}

	var genres []models.Genre
	if err := db.Order("name ASC").Limit(limit).Offset(offset).Find(&genres).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "Gagal mengambil daftar genre", err)
	}

	return okList(c, "Daftar genre berhasil diambil", genres, page, limit, total)
}

func GetGenre(c *fiber.Ctx) error {
	param := c.Params("id")
	var g models.Genre

	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&g, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
			}
			return fail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
		}
		return ok(c, http.StatusOK, "Genre berhasil diambil", g)
	}

	if err := database.DB.Where("slug = ?", param).First(&g).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
		}
		return fail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
	}
	return ok(c, http.StatusOK, "Genre berhasil diambil", g)
}

func CreateGenre(c *fiber.Ctx) error {
	var payload struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}
	payload.Name = strings.TrimSpace(payload.Name)
	payload.Slug = strings.TrimSpace(payload.Slug)
	if payload.Name == "" || payload.Slug == "" {
		return fail(c, http.StatusBadRequest, "Name dan slug wajib diisi", nil)
	}

	var cnt int64
	database.DB.Model(&models.Genre{}).
		Where("name = ? OR slug = ?", payload.Name, payload.Slug).
		Count(&cnt)
	if cnt > 0 {
		return fail(c, http.StatusConflict, "Name atau slug sudah digunakan", nil)
	}

	g := models.Genre{
		Name:      payload.Name,
		Slug:      payload.Slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := database.DB.Create(&g).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "Gagal membuat genre", err)
	}
	return ok(c, http.StatusCreated, "Genre berhasil dibuat", g)
}

func UpdateGenre(c *fiber.Ctx) error {
	param := c.Params("id")
	var g models.Genre

	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&g, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
			}
			return fail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
		}
	} else {
		if err := database.DB.Where("slug = ?", param).First(&g).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
			}
			return fail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
		}
	}

	var payload struct {
		Name *string `json:"name"`
		Slug *string `json:"slug"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return fail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}

	if payload.Name != nil {
		newName := strings.TrimSpace(*payload.Name)
		if newName == "" {
			return fail(c, http.StatusBadRequest, "Name tidak boleh kosong", nil)
		}
		var cnt int64
		database.DB.Model(&models.Genre{}).Where("name = ? AND id <> ?", newName, g.ID).Count(&cnt)
		if cnt > 0 {
			return fail(c, http.StatusConflict, "Name sudah digunakan", nil)
		}
		g.Name = newName
	}
	if payload.Slug != nil {
		newSlug := strings.TrimSpace(*payload.Slug)
		if newSlug == "" {
			return fail(c, http.StatusBadRequest, "Slug tidak boleh kosong", nil)
		}
		var cnt int64
		database.DB.Model(&models.Genre{}).Where("slug = ? AND id <> ?", newSlug, g.ID).Count(&cnt)
		if cnt > 0 {
			return fail(c, http.StatusConflict, "Slug sudah digunakan", nil)
		}
		g.Slug = newSlug
	}

	g.UpdatedAt = time.Now()
	if err := database.DB.Save(&g).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "Gagal memperbarui genre", err)
	}
	return ok(c, http.StatusOK, "Genre berhasil diperbarui", g)
}

func DeleteGenre(c *fiber.Ctx) error {
	param := c.Params("id")
	var g models.Genre

	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&g, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
			}
			return fail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
		}
	} else {
		if err := database.DB.Where("slug = ?", param).First(&g).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
			}
			return fail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
		}
	}

	if err := database.DB.Delete(&g).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "Gagal menghapus genre", err)
	}
	return ok(c, http.StatusOK, "Genre berhasil dihapus", nil)
}
