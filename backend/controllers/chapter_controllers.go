// controllers/chapter_controllers.go
package controllers

import (
	"net/http"
	"os"
	"strconv"
	"time"
	"strings"

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
	// Tampilkan detail error saat non-production
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
    if page < 1 { page = 1 }
    if limit < 1 || limit > 200 { limit = 50 }
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
    if strings.ToUpper(role) != "ADMIN" {
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
    if strings.ToUpper(role) == "ADMIN" {
        msg = "Daftar chapter (semua)"
    }
    return chapterListOK(c, msg, chapters, page, limit, total)
}


func AddChapter(c *fiber.Ctx) error {
	var payload struct {
		NovelID     uint    `json:"novel_id"`
		Title       string  `json:"title"`
		Content     string  `json:"content"`
		PublishNow  *bool   `json:"publish_now"`
		PublishedAt *string `json:"published_at"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return chapterFail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}
	if payload.NovelID == 0 {
		return chapterFail(c, http.StatusBadRequest, "novel_id wajib diisi", nil)
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
	_ = database.DB.Where("novel_id = ?", payload.NovelID).Order("order_no DESC").First(&last)
	nextOrder := 1
	if last.ID != 0 {
		nextOrder = last.OrderNo + 1
	}

	var pubAt *time.Time
	if payload.PublishNow != nil && *payload.PublishNow {
		now := time.Now()
		pubAt = &now
	} else if payload.PublishedAt != nil && *payload.PublishedAt != "" {
		if t, err := time.Parse(time.RFC3339, *payload.PublishedAt); err == nil {
			pubAt = &t
		} else {
			return chapterFail(c, http.StatusBadRequest, "published_at harus RFC3339 (contoh: 2006-01-02T15:04:05Z)", err)
		}
	}

	title := payload.Title
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
	return chapterOK(c, http.StatusCreated, "Chapter ditambahkan", ch)
}

func ListChapters(c *fiber.Ctx) error {
	nid := c.Params("novel_id")
	nidInt, err := strconv.Atoi(nid)
	if err != nil {
		return chapterFail(c, http.StatusBadRequest, "novel_id tidak valid", err)
	}

	var chapters []models.Chapter
	if err := database.DB.
		Where("novel_id = ?", nidInt).
		Scopes(onlyPublished).
		Order("order_no ASC").
		Find(&chapters).Error; err != nil {
		return chapterFail(c, http.StatusInternalServerError, "Gagal mengambil chapters", err)
	}
	return chapterOK(c, http.StatusOK, "Daftar chapter terbit", chapters)
}

func GetChapter(c *fiber.Ctx) error {
	id := c.Params("id")
	var ch models.Chapter
	if err := database.DB.Scopes(onlyPublished).First(&ch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return chapterFail(c, http.StatusNotFound, "Chapter tidak ditemukan atau belum terbit", nil)
		}
		return chapterFail(c, http.StatusInternalServerError, "Gagal mengambil chapter", err)
	}
	return chapterOK(c, http.StatusOK, "Detail chapter", ch)
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
		return chapterFail(c, http.StatusForbidden, "Tidak berwenang mengubah/menghapus chapter ini", nil)
	}

	var payload struct {
		Title       *string `json:"title"`
		Content     *string `json:"content"`
		OrderNo     *int    `json:"order_no"`
		PublishNow  *bool   `json:"publish_now"`
		PublishedAt *string `json:"published_at"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return chapterFail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}

	if payload.Title != nil {
		ch.Title = payload.Title
	}
	if payload.Content != nil {
		ch.Content = payload.Content
	}
	if payload.OrderNo != nil {
		ch.OrderNo = *payload.OrderNo
	}

	if payload.PublishNow != nil && *payload.PublishNow {
		now := time.Now()
		ch.PublishedAt = &now
	} else if payload.PublishedAt != nil {
		if *payload.PublishedAt == "" {
			ch.PublishedAt = nil
		} else {
			if t, err := time.Parse(time.RFC3339, *payload.PublishedAt); err == nil {
				ch.PublishedAt = &t
			} else {
				return chapterFail(c, http.StatusBadRequest, "published_at harus RFC3339 (contoh: 2006-01-02T15:04:05Z)", err)
			}
		}
	}

	ch.UpdatedAt = time.Now()
	if err := database.DB.Save(&ch).Error; err != nil {
		return chapterFail(c, http.StatusInternalServerError, "Gagal memperbarui chapter", err)
	}
	return chapterOK(c, http.StatusOK, "Chapter diperbarui", ch)
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
		return chapterFail(c, http.StatusForbidden, "Tidak berwenang mengubah/menghapus chapter ini", nil)
	}

	if err := database.DB.Delete(&ch).Error; err != nil {
		return chapterFail(c, http.StatusInternalServerError, "Gagal menghapus chapter", err)
	}
	return chapterOK(c, http.StatusOK, "Chapter dihapus", nil)
}
