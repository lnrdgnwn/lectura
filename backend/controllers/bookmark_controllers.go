// controllers/bookmark_controllers.go
package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"final_project/database"
	"final_project/models"

	"github.com/gofiber/fiber/v2"
)

func bookmarkOK(c *fiber.Ctx, status int, msg string, data any) error {
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

func bookmarkListOK(c *fiber.Ctx, msg string, data any, page, limit int, total int64) error {
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

func bookmarkFail(c *fiber.Ctx, status int, msg string, err error) error {
	resp := fiber.Map{
		"success": false,
		"message": msg,
	}
	if err != nil && os.Getenv("APP_ENV") != "production" {
		resp["error"] = err.Error()
	}
	return c.Status(status).JSON(resp)
}

func uidFromCtx(c *fiber.Ctx) (uint, error) {
	v := c.Locals("user_id")
	if v == nil {
		return 0, errors.New("user_id not found in context")
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
		var n int
		_, err := fmt.Sscanf(t, "%d", &n)
		if err != nil {
			return 0, errors.New("cannot parse user_id string")
		}
		return uint(n), nil
	default:
		return 0, errors.New("unsupported user_id type")
	}
}

func AddBookmark(c *fiber.Ctx) error {
	var body struct {
		NovelID uint `json:"novel_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return bookmarkFail(c, http.StatusBadRequest, "Payload tidak valid", err)
	}
	if body.NovelID == 0 {
		return bookmarkFail(c, http.StatusBadRequest, "novel_id wajib diisi", nil)
	}

	uid, err := uidFromCtx(c)
	if err != nil {
		return bookmarkFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	var exist int64
	if err := database.DB.Model(&models.Bookmark{}).
		Where("user_id = ? AND novel_id = ?", uid, body.NovelID).
		Count(&exist).Error; err != nil {
		return bookmarkFail(c, http.StatusInternalServerError, "Gagal memeriksa bookmark", err)
	}
	if exist > 0 {
		return bookmarkFail(c, http.StatusConflict, "Novel sudah dibookmark", nil)
	}

	b := models.Bookmark{
		UserID:    uid,
		NovelID:   body.NovelID,
		CreatedAt: time.Now(),
	}
	if err := database.DB.Create(&b).Error; err != nil {
		return bookmarkFail(c, http.StatusInternalServerError, "Gagal menambahkan bookmark", err)
	}
	return bookmarkOK(c, http.StatusCreated, "Bookmark ditambahkan", b)
}

func RemoveBookmark(c *fiber.Ctx) error {
	novelIDParam := c.Params("novel_id")
	if novelIDParam == "" {
		return bookmarkFail(c, http.StatusBadRequest, "novel_id wajib diisi", nil)
	}

	var novelID uint
	if _, err := fmt.Sscanf(novelIDParam, "%d", &novelID); err != nil || novelID == 0 {
		return bookmarkFail(c, http.StatusBadRequest, "novel_id tidak valid", err)
	}

	uid, err := uidFromCtx(c)
	if err != nil {
		return bookmarkFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	res := database.DB.Where("user_id = ? AND novel_id = ?", uid, novelID).Delete(&models.Bookmark{})
	if res.Error != nil {
		return bookmarkFail(c, http.StatusInternalServerError, "Gagal menghapus bookmark", res.Error)
	}
	if res.RowsAffected == 0 {
		return bookmarkFail(c, http.StatusNotFound, "Bookmark tidak ditemukan", nil)
	}
	return bookmarkOK(c, http.StatusOK, "Bookmark dihapus", nil)
}

func ListBookmarks(c *fiber.Ctx) error {
	uid, err := uidFromCtx(c)
	if err != nil {
		return bookmarkFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	var bookmarks []models.Bookmark
	if err := database.DB.Where("user_id = ?", uid).
		Order("created_at DESC").
		Find(&bookmarks).Error; err != nil {
		return bookmarkFail(c, http.StatusInternalServerError, "Gagal mengambil bookmark", err)
	}

	type NovelMini struct {
		ID    uint   `json:"id"`
		Title string `json:"title"`
		Slug  string `json:"slug"`
	}
	novelIDs := make([]uint, 0, len(bookmarks))
	for _, b := range bookmarks {
		novelIDs = append(novelIDs, b.NovelID)
	}
	novelsMap := map[uint]NovelMini{}
	if len(novelIDs) > 0 {
		var novels []models.Novel
		if err := database.DB.Where("id IN ?", novelIDs).Find(&novels).Error; err == nil {
			for _, n := range novels {
				novelsMap[n.ID] = NovelMini{ID: n.ID, Title: n.Title, Slug: n.Slug}
			}
		}
	}

	resp := make([]fiber.Map, 0, len(bookmarks))
	for _, b := range bookmarks {
		resp = append(resp, fiber.Map{
			"id":         b.ID,
			"novel_id":   b.NovelID,
			"created_at": b.CreatedAt,
			"novel":      novelsMap[b.NovelID],
		})
	}

	return bookmarkListOK(c, "Daftar bookmark", resp, 1, len(resp), int64(len(resp)))
}
