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


func genreok(c *fiber.Ctx, status int, msg string, data any) error {
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

func genreokList(c *fiber.Ctx, msg string, data any, page, limit int, total int64) error {
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

func genrefail(c *fiber.Ctx, status int, msg string, err error) error {
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
		return genrefail(c, http.StatusInternalServerError, "Gagal menghitung total genre", err)
	}

	var genres []models.Genre
	if err := db.Order("name ASC").Limit(limit).Offset(offset).Find(&genres).Error; err != nil {
		return genrefail(c, http.StatusInternalServerError, "Gagal mengambil daftar genre", err)
	}

	return genreokList(c, "Daftar genre berhasil diambil", genres, page, limit, total)
}

func GetGenre(c *fiber.Ctx) error {
	param := c.Params("id")
	var g models.Genre

	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&g, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return genrefail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
			}
			return genrefail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
		}
		return genreok(c, http.StatusOK, "Genre berhasil diambil", g)
	}

	if err := database.DB.Where("slug = ?", param).First(&g).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return genrefail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
		}
		return genrefail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
	}
	return genreok(c, http.StatusOK, "Genre berhasil diambil", g)
}

func GetHomeGenres(c *fiber.Ctx) error {
	var genres []models.Genre
	if err := database.DB.
		Where("show_on_home = ?", true).
		Order("name ASC").
		Find(&genres).Error; err != nil {
		return genrefail(c, http.StatusInternalServerError, "Gagal mengambil genre untuk halaman utama", err)
	}

	return genreok(c, http.StatusOK, "Daftar genre untuk halaman utama", genres)
}

func GenreShowByAdmin(c *fiber.Ctx) error {
	var body struct {
		GenreIDs []uint `json:"genre_ids"`
	}

	if err := c.BodyParser(&body); err != nil {
		return genrefail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}
	if len(body.GenreIDs) == 0 {
		return genrefail(c, http.StatusBadRequest, "genre_ids wajib diisi (minimal 1 genre)", nil)
	}

	var count int64
	if err := database.DB.Model(&models.Genre{}).
		Where("id IN ?", body.GenreIDs).
		Count(&count).Error; err != nil {
		return genrefail(c, http.StatusInternalServerError, "Gagal memeriksa daftar genre", err)
	}
	if count != int64(len(body.GenreIDs)) {
		return genrefail(c, http.StatusBadRequest, "Salah satu genre_id tidak ditemukan", nil)
	}

	tx := database.DB.Begin()

	if err := tx.Model(&models.Genre{}).
		Where("show_on_home = ?", true).
		Update("show_on_home", false).Error; err != nil {
		tx.Rollback()
		return genrefail(c, http.StatusInternalServerError, "Gagal mereset daftar genre halaman utama", err)
	}

	if err := tx.Model(&models.Genre{}).
		Where("id IN ?", body.GenreIDs).
		Update("show_on_home", true).Error; err != nil {
		tx.Rollback()
		return genrefail(c, http.StatusInternalServerError, "Gagal mengatur genre halaman utama", err)
	}

	if err := tx.Commit().Error; err != nil {
		return genrefail(c, http.StatusInternalServerError, "Gagal menyimpan perubahan genre halaman utama", err)
	}

	var genres []models.Genre
	_ = database.DB.
		Where("show_on_home = ?", true).
		Order("name ASC").
		Find(&genres).Error

	return genreok(c, http.StatusOK, "Genre untuk halaman utama berhasil diperbarui", genres)
}

func CreateGenre(c *fiber.Ctx) error {
	var payload struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return genrefail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}
	payload.Name = strings.TrimSpace(payload.Name)
	payload.Slug = strings.TrimSpace(payload.Slug)
	if payload.Name == "" || payload.Slug == "" {
		return genrefail(c, http.StatusBadRequest, "Name dan slug wajib diisi", nil)
	}

	var cnt int64
	database.DB.Model(&models.Genre{}).
		Where("name = ? OR slug = ?", payload.Name, payload.Slug).
		Count(&cnt)
	if cnt > 0 {
		return genrefail(c, http.StatusConflict, "Name atau slug sudah digunakan", nil)
	}

	g := models.Genre{
		Name:      payload.Name,
		Slug:      payload.Slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := database.DB.Create(&g).Error; err != nil {
		return genrefail(c, http.StatusInternalServerError, "Gagal membuat genre", err)
	}
	return genreok(c, http.StatusCreated, "Genre berhasil dibuat", g)
}

func UpdateGenre(c *fiber.Ctx) error {
	param := c.Params("id")
	var g models.Genre

	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&g, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return genrefail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
			}
			return genrefail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
		}
	} else {
		if err := database.DB.Where("slug = ?", param).First(&g).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return genrefail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
			}
			return genrefail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
		}
	}

	var payload struct {
		Name *string `json:"name"`
		Slug *string `json:"slug"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return genrefail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}

	if payload.Name != nil {
		newName := strings.TrimSpace(*payload.Name)
		if newName == "" {
			return genrefail(c, http.StatusBadRequest, "Name tidak boleh kosong", nil)
		}
		var cnt int64
		database.DB.Model(&models.Genre{}).Where("name = ? AND id <> ?", newName, g.ID).Count(&cnt)
		if cnt > 0 {
			return genrefail(c, http.StatusConflict, "Name sudah digunakan", nil)
		}
		g.Name = newName
	}
	if payload.Slug != nil {
		newSlug := strings.TrimSpace(*payload.Slug)
		if newSlug == "" {
			return genrefail(c, http.StatusBadRequest, "Slug tidak boleh kosong", nil)
		}
		var cnt int64
		database.DB.Model(&models.Genre{}).Where("slug = ? AND id <> ?", newSlug, g.ID).Count(&cnt)
		if cnt > 0 {
			return genrefail(c, http.StatusConflict, "Slug sudah digunakan", nil)
		}
		g.Slug = newSlug
	}

	g.UpdatedAt = time.Now()
	if err := database.DB.Save(&g).Error; err != nil {
		return genrefail(c, http.StatusInternalServerError, "Gagal memperbarui genre", err)
	}
	return genreok(c, http.StatusOK, "Genre berhasil diperbarui", g)
}

func DeleteGenre(c *fiber.Ctx) error {
	param := c.Params("id")
	var g models.Genre

	if id, err := strconv.Atoi(param); err == nil {
		if err := database.DB.First(&g, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return genrefail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
			}
			return genrefail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
		}
	} else {
		if err := database.DB.Where("slug = ?", param).First(&g).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return genrefail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
			}
			return genrefail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
		}
	}

	if err := database.DB.Delete(&g).Error; err != nil {
		return genrefail(c, http.StatusInternalServerError, "Gagal menghapus genre", err)
	}
	return genreok(c, http.StatusOK, "Genre berhasil dihapus", nil)
}
