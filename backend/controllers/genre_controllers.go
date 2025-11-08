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

// ListGenres godoc
// @Summary      List genre
// @Description  Mengambil daftar genre dengan opsi pencarian dan pagination.
// @Tags         Genres
// @Accept       json
// @Produce      json
// @Param        q      query     string  false  "Cari berdasarkan name atau slug"
// @Param        page   query     int     false  "Halaman (default 1)"
// @Param        limit  query     int     false  "Jumlah per halaman (default 50)"
// @Success      200    {object}  map[string]interface{}  "Daftar genre + meta pagination"
// @Failure      500    {object}  map[string]interface{}  "Gagal mengambil daftar genre"
// @Router       /genres [get]
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

// GetGenre godoc
// @Summary      Detail genre
// @Description  Mengambil detail genre berdasarkan ID atau slug.
// @Tags         Genres
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID atau slug genre"
// @Success      200  {object}  map[string]interface{}  "Detail genre"
// @Failure      404  {object}  map[string]interface{}  "Genre tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil genre"
// @Router       /genres/{id} [get]
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

// GetHomeGenres godoc
// @Summary      List genre untuk halaman utama
// @Description  Mengambil daftar genre yang ditandai tampil di halaman utama (show_on_home = true).
// @Tags         Genres
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Daftar genre untuk halaman utama"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil genre"
// @Router       /genres/home [get]
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

// GenreShowByAdmin godoc
// @Summary      Atur genre yang tampil di halaman utama
// @Description  Mengatur ulang daftar genre yang memiliki show_on_home = true berdasarkan genre_ids yang dikirim. Hanya untuk admin.
// @Tags         Genres
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "Body: {\"genre_ids\": [1,2,3]}"
// @Success      200   {object}  map[string]interface{}  "Genre untuk halaman utama berhasil diperbarui"
// @Failure      400   {object}  map[string]interface{}  "Payload tidak valid / genre_ids kosong / genre tidak ditemukan"
// @Failure      500   {object}  map[string]interface{}  "Gagal memperbarui genre halaman utama"
// @Router       /admin/genres/home [put]
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

// CreateGenre godoc
// @Summary      Buat genre baru
// @Description  Membuat genre baru dengan name dan slug unik.
// @Tags         Genres
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "Body: {\"name\": \"Action\", \"slug\": \"action\"}"
// @Success      201   {object}  map[string]interface{}  "Genre berhasil dibuat"
// @Failure      400   {object}  map[string]interface{}  "Payload tidak valid / field wajib kosong"
// @Failure      409   {object}  map[string]interface{}  "Name atau slug sudah digunakan"
// @Failure      500   {object}  map[string]interface{}  "Gagal membuat genre"
// @Router       /genres [post]
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

// UpdateGenre godoc
// @Summary      Update genre
// @Description  Mengubah name dan/atau slug genre berdasarkan ID atau slug. Wajib unik.
// @Tags         Genres
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "ID atau slug genre"
// @Param        body  body      object  true  "Body: {\"name\": \"Nama Baru\", \"slug\": \"slug-baru\"} (opsional per field)"
// @Success      200   {object}  map[string]interface{}  "Genre berhasil diperbarui"
// @Failure      400   {object}  map[string]interface{}  "Payload tidak valid / field kosong"
// @Failure      404   {object}  map[string]interface{}  "Genre tidak ditemukan"
// @Failure      409   {object}  map[string]interface{}  "Name atau slug sudah digunakan"
// @Failure      500   {object}  map[string]interface{}  "Gagal memperbarui genre"
// @Router       /genres/{id} [put]
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

// DeleteGenre godoc
// @Summary      Hapus genre
// @Description  Menghapus genre berdasarkan ID atau slug. Relasi many2many ke novel akan ikut terhapus karena constraint OnDelete:CASCADE di join table.
// @Tags         Genres
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      string  true  "ID atau slug genre"
// @Success      200  {object}  map[string]interface{}  "Genre berhasil dihapus"
// @Failure      404  {object}  map[string]interface{}  "Genre tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal menghapus genre"
// @Router       /genres/{id} [delete]
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
