package controllers

import (
	"net/http"
	"strconv"
	"time"

	"final_project/database"
	"final_project/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AddChapter - POST /chapters
// Body JSON: { "novel_id": uint, "title": "string", "content":"string" }
func AddChapter(c *fiber.Ctx) error {
	var payload struct {
		NovelID uint   `json:"novel_id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "payload tidak valid"})
	}
	if payload.NovelID == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "novel_id wajib diisi"})
	}

	// cari last order_no untuk novel tersebut (append)
	var last models.Chapter
	database.DB.Where("novel_id = ?", payload.NovelID).Order("order_no DESC").First(&last)
	nextOrder := 1
	if last.ID != 0 {
		nextOrder = last.OrderNo + 1
	}

	ch := models.Chapter{
		NovelID:   payload.NovelID,
		OrderNo:   nextOrder,
		Title:     &payload.Title,
		Content:   &payload.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.DB.Create(&ch).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menambah chapter", "error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "chapter ditambahkan", "chapter": ch})
}

// ListChapters - GET /novels/:novel_id/chapters
// Mengembalikan semua chapter untuk sebuah novel (urut berdasarkan order_no)
func ListChapters(c *fiber.Ctx) error {
	nid := c.Params("novel_id")
	nidInt, err := strconv.Atoi(nid)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "novel_id tidak valid"})
	}

	var chapters []models.Chapter
	if err := database.DB.Where("novel_id = ?", nidInt).Order("order_no ASC").Find(&chapters).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil chapters", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"chapters": chapters})
}

// GetChapter - GET /chapters/:id
func GetChapter(c *fiber.Ctx) error {
	id := c.Params("id")
	var ch models.Chapter
	if err := database.DB.First(&ch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "chapter tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil chapter", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"chapter": ch})
}

// UpdateChapter - PUT /chapters/:id
// Body JSON (partial allowed): { "title": "string", "content":"string", "order_no": int, "published_at": "RFC3339 string" }
func UpdateChapter(c *fiber.Ctx) error {
	id := c.Params("id")
	var ch models.Chapter
	if err := database.DB.First(&ch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "chapter tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil chapter", "error": err.Error()})
	}

	var payload struct {
		Title       *string `json:"title"`
		Content     *string `json:"content"`
		OrderNo     *int    `json:"order_no"`
		PublishedAt *string `json:"published_at"` // expect RFC3339 if provided
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "payload tidak valid"})
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
	if payload.PublishedAt != nil && *payload.PublishedAt != "" {
		if t, err := time.Parse(time.RFC3339, *payload.PublishedAt); err == nil {
			ch.PublishedAt = &t
		} else {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "published_at harus format RFC3339 (contoh: 2006-01-02T15:04:05Z)"})
		}
	}

	ch.UpdatedAt = time.Now()
	if err := database.DB.Save(&ch).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal memperbarui chapter", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "chapter diperbarui", "chapter": ch})
}

// DeleteChapter - DELETE /chapters/:id
func DeleteChapter(c *fiber.Ctx) error {
	id := c.Params("id")
	var ch models.Chapter
	if err := database.DB.First(&ch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "chapter tidak ditemukan"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal mengambil chapter", "error": err.Error()})
	}

	if err := database.DB.Delete(&ch).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "gagal menghapus chapter", "error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "chapter dihapus"})
}
