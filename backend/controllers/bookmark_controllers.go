package controllers

import (
	"errors"
	"net/http"
	"time"
	"fmt"

	"final_project/database"
	"final_project/models"

	"github.com/gofiber/fiber/v2"
)

// uidFromCtx mengonversi c.Locals("user_id") ke uint
// (nama berbeda agar tidak konflik dengan helper lain di package)
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
		// parse string if necessary
		// fiber middleware sometimes puts numeric claims as string
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

// AddBookmark - POST /bookmarks
// Body JSON: { "novel_id": <uint> }
func AddBookmark(c *fiber.Ctx) error {
	var body struct {
		NovelID uint `json:"novel_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "payload tidak valid"})
	}
	if body.NovelID == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "novel_id wajib diisi"})
	}

	uid, err := uidFromCtx(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	// Cek apakah bookmark sudah ada
	var exist int64
	if err := database.DB.Model(&models.Bookmark{}).
		Where("user_id = ? AND novel_id = ?", uid, body.NovelID).
		Count(&exist).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal memeriksa bookmark"})
	}
	if exist > 0 {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"message": "novel sudah dibookmark"})
	}

	b := models.Bookmark{
		UserID:    uid,
		NovelID:   body.NovelID,
		CreatedAt: time.Now(),
	}
	if err := database.DB.Create(&b).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menambahkan bookmark"})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "bookmark ditambahkan", "bookmark": b})
}

// RemoveBookmark - DELETE /bookmarks/:novel_id
// Menghapus bookmark untuk user yang sedang login pada novel tertentu
func RemoveBookmark(c *fiber.Ctx) error {
	novelIDParam := c.Params("novel_id")
	if novelIDParam == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "novel_id required"})
	}
	// parse manual to uint
	var novelID uint
	_, err := fmt.Sscanf(novelIDParam, "%d", &novelID)
	if err != nil || novelID == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "novel_id tidak valid"})
	}

	uid, err := uidFromCtx(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	// hapus bookmark spesifik
	res := database.DB.Where("user_id = ? AND novel_id = ?", uid, novelID).Delete(&models.Bookmark{})
	if res.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menghapus bookmark"})
	}
	if res.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "bookmark tidak ditemukan"})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "bookmark dihapus"})
}

// ListBookmarks - GET /users/me/bookmarks
// Mengembalikan daftar bookmark user (bersama data novel minimal)
func ListBookmarks(c *fiber.Ctx) error {
	uid, err := uidFromCtx(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
	}

	var bookmarks []models.Bookmark
	if err := database.DB.Where("user_id = ?", uid).Order("created_at DESC").Find(&bookmarks).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil bookmark"})
	}

	// Ambil daftar novel terkait (batch) supaya response lebih informatif
	novelIDs := make([]uint, 0, len(bookmarks))
	for _, b := range bookmarks {
		novelIDs = append(novelIDs, b.NovelID)
	}

	type NovelMini struct {
		ID    uint   `json:"id"`
		Title string `json:"title"`
		Slug  string `json:"slug"`
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

	// compose response
	resp := make([]fiber.Map, 0, len(bookmarks))
	for _, b := range bookmarks {
		item := fiber.Map{
			"id":         b.ID,
			"novel_id":   b.NovelID,
			"created_at": b.CreatedAt,
			"novel":      novelsMap[b.NovelID], // jika kosong, akan berupa zero value
		}
		resp = append(resp, item)
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"bookmarks": resp})
}
