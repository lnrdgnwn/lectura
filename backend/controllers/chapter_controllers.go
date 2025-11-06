// controllers/chapter_controllers.go
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

func chapterOK(c *fiber.Ctx, status int, msg string, data any) error {
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

func chapterListOK(c *fiber.Ctx, msg string, data any, page, limit int, total int64) error {
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

func chapterFail(c *fiber.Ctx, status int, msg string, err error) error {
	resp := fiber.Map{
		"success": false,
		"message": msg,
	}
	if err != nil && os.Getenv("APP_ENV") != "production" {
		resp["error"] = err.Error()
	}
	return c.Status(status).JSON(resp)
}

func onlyPublished(db *gorm.DB) *gorm.DB {
	return db.Where("published_at IS NOT NULL AND published_at <= ?", time.Now())
}

func GetAllChapters(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	offset := (page - 1) * limit

	var novelID int
	if nv := strings.TrimSpace(c.Query("novel_id", "")); nv != "" {
		n, err := strconv.Atoi(nv)
		if err != nil || n <= 0 {
			return chapterFail(c, http.StatusBadRequest, "novel_id tidak valid", err)
		}
		novelID = n
	}

	_, role, _ := getAuthFromAccessCookieNovel(c)

	db := database.DB.Model(&models.Chapter{})

	if role != "admin" {
		db = db.Scopes(onlyPublished)
	}

	if novelID > 0 {
		db = db.Where("novel_id = ?", novelID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return chapterFail(c, http.StatusInternalServerError, "Gagal menghitung total chapter", err)
	}

	var chapters []models.Chapter
	if err := db.
		Order("novel_id ASC, order_no ASC").
		Limit(limit).Offset(offset).
		Find(&chapters).Error; err != nil {
		return chapterFail(c, http.StatusInternalServerError, "Gagal mengambil chapters", err)
	}

	msg := "Daftar chapter (terbit saja)"
	if role == "admin" {
		msg = "Daftar chapter (semua)"
	}
	return chapterListOK(c, msg, chapters, page, limit, total)
}

func AddChapter(c *fiber.Ctx) error {
	var payload struct {
		NovelID uint   `json:"novel_id"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Publish *bool  `json:"publish"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return chapterFail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}
	if payload.NovelID == 0 {
		return chapterFail(c, http.StatusBadRequest, "novel_id wajib diisi", nil)
	}
	if strings.TrimSpace(payload.Title) == "" {
		return chapterFail(c, http.StatusBadRequest, "title wajib diisi", nil)
	}

	uid, role, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return chapterFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	var novel models.Novel
	if err := database.DB.First(&novel, payload.NovelID).Error; err != nil {
		return chapterFail(c, http.StatusBadRequest, "Novel tidak ditemukan", err)
	}
	if novel.AuthorID != uid && role != "admin" {
		return chapterFail(c, http.StatusForbidden, "Tidak berwenang menambah chapter untuk novel ini", nil)
	}

	var last models.Chapter
	_ = database.DB.Where("novel_id = ?", payload.NovelID).
		Order("order_no DESC").
		First(&last)
	nextOrder := 1
	if last.ID != 0 {
		nextOrder = last.OrderNo + 1
	}

	var pubAt *time.Time
	if payload.Publish != nil && *payload.Publish {
		now := time.Now()
		pubAt = &now
	}

	title := strings.TrimSpace(payload.Title)
	content := payload.Content

	ch := models.Chapter{
		NovelID:     payload.NovelID,
		OrderNo:     nextOrder,
		Title:       &title,
		Content:     &content,
		PublishedAt: pubAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := database.DB.Create(&ch).Error; err != nil {
		return chapterFail(c, http.StatusInternalServerError, "Gagal menambah chapter", err)
	}
	statusMsg := "Chapter dibuat sebagai draft"
	if ch.PublishedAt != nil {
		statusMsg = "Chapter dibuat dan dipublish"
	}
	return chapterOK(c, http.StatusCreated, statusMsg, ch)
}

func ListChapters(c *fiber.Ctx) error {
	nid := c.Params("novel_id")
	nidInt, err := strconv.Atoi(nid)
	if err != nil || nidInt <= 0 {
		return chapterFail(c, http.StatusBadRequest, "novel_id tidak valid", err)
	}

	var novel models.Novel
	if err := database.DB.Select("id, author_id").First(&novel, nidInt).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return chapterFail(c, http.StatusNotFound, "Novel tidak ditemukan", nil)
		}
		return chapterFail(c, http.StatusInternalServerError, "Gagal mengambil novel", err)
	}

	uid, role, _ := getAuthFromAccessCookieNovel(c)

	db := database.DB.Where("novel_id = ?", nidInt)

	if role == "admin" || uid == novel.AuthorID {
	} else {
		db = db.Scopes(onlyPublished)
	}

	var chapters []models.Chapter
	if err := db.Order("order_no ASC").Find(&chapters).Error; err != nil {
		return chapterFail(c, http.StatusInternalServerError, "Gagal mengambil chapters", err)
	}

	msg := "Daftar chapter terbit"
	if role == "admin" || uid == novel.AuthorID {
		msg = "Daftar chapter (termasuk draft)"
	}
	return chapterOK(c, http.StatusOK, msg, chapters)
}

func GetChapter(c *fiber.Ctx) error {
	id := c.Params("id")

	var ch models.Chapter
	if err := database.DB.First(&ch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return chapterFail(c, http.StatusNotFound, "Chapter tidak ditemukan", nil)
		}
		return chapterFail(c, http.StatusInternalServerError, "Gagal mengambil chapter", err)
	}

	if ch.PublishedAt != nil && ch.PublishedAt.Before(time.Now().Add(1*time.Second)) {
		return chapterOK(c, http.StatusOK, "Detail chapter", ch)
	}

	uid, role, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return chapterFail(c, http.StatusForbidden, "Chapter ini masih draft", nil)
	}

	var novel models.Novel
	if err := database.DB.Select("id, author_id").First(&novel, ch.NovelID).Error; err != nil {
		return chapterFail(c, http.StatusInternalServerError, "Gagal memeriksa pemilik chapter", err)
	}

	if role == "admin" || uid == novel.AuthorID {
		return chapterOK(c, http.StatusOK, "Detail chapter (draft)", ch)
	}

	return chapterFail(c, http.StatusForbidden, "Tidak berwenang mengakses draft chapter ini", nil)
}

func UpdateChapter(c *fiber.Ctx) error {
	id := c.Params("id")

	var ch models.Chapter
	if err := database.DB.First(&ch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return chapterFail(c, http.StatusNotFound, "Chapter tidak ditemukan", nil)
		}
		return chapterFail(c, http.StatusInternalServerError, "Gagal mengambil chapter", err)
	}

	uid, role, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return chapterFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	var novel models.Novel
	if err := database.DB.Select("id, author_id").First(&novel, ch.NovelID).Error; err != nil {
		return chapterFail(c, http.StatusBadRequest, "Novel tidak ditemukan", err)
	}
	if novel.AuthorID != uid && role != "admin" {
		return chapterFail(c, http.StatusForbidden, "Tidak berwenang mengubah chapter ini", nil)
	}

	var payload struct {
		Title   *string `json:"title"`
		Content *string `json:"content"`
		OrderNo *int    `json:"order_no"`
		Publish *bool   `json:"publish"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return chapterFail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}

	if payload.Title != nil {
		t := strings.TrimSpace(*payload.Title)
		if t == "" {
			return chapterFail(c, http.StatusBadRequest, "title tidak boleh kosong", nil)
		}
		ch.Title = &t
	}
	if payload.Content != nil {
		ch.Content = payload.Content
	}
	if payload.OrderNo != nil {
		if *payload.OrderNo <= 0 {
			return chapterFail(c, http.StatusBadRequest, "order_no harus lebih dari 0", nil)
		}
		ch.OrderNo = *payload.OrderNo
	}

	if payload.Publish != nil {
		if *payload.Publish {
			now := time.Now()
			ch.PublishedAt = &now
		} else {
			ch.PublishedAt = nil
		}
	}

	ch.UpdatedAt = time.Now()
	if err := database.DB.Save(&ch).Error; err != nil {
		return chapterFail(c, http.StatusInternalServerError, "Gagal memperbarui chapter", err)
	}

	statusMsg := "Chapter diperbarui"
	if ch.PublishedAt == nil {
		statusMsg += " sebagai draft"
	} else {
		statusMsg += " dan sudah terbit"
	}
	return chapterOK(c, http.StatusOK, statusMsg, ch)
}

func DeleteChapter(c *fiber.Ctx) error {
	id := c.Params("id")

	var ch models.Chapter
	if err := database.DB.First(&ch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return chapterFail(c, http.StatusNotFound, "Chapter tidak ditemukan", nil)
		}
		return chapterFail(c, http.StatusInternalServerError, "Gagal mengambil chapter", err)
	}

	uid, role, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return chapterFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	var novel models.Novel
	if err := database.DB.Select("id, author_id").First(&novel, ch.NovelID).Error; err != nil {
		return chapterFail(c, http.StatusBadRequest, "Novel tidak ditemukan", err)
	}
	if novel.AuthorID != uid && role != "admin" {
		return chapterFail(c, http.StatusForbidden, "Tidak berwenang menghapus chapter ini", nil)
	}

	if err := database.DB.Delete(&ch).Error; err != nil {
		return chapterFail(c, http.StatusInternalServerError, "Gagal menghapus chapter", err)
	}
	return chapterOK(c, http.StatusOK, "Chapter dihapus", nil)
}
