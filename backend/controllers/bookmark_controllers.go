package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
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

// AddBookmark godoc
// @Summary      Tambah bookmark novel
// @Description  Menambahkan novel ke daftar bookmark user yang sedang login.
// @Tags         Bookmarks
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        body  body      map[string]uint  true  "novel_id, contoh: {\"novel_id\": 1}"
// @Success      201   {object}  map[string]interface{}  "Bookmark ditambahkan"
// @Failure      400   {object}  map[string]interface{}  "Payload tidak valid / novel_id kosong"
// @Failure      401   {object}  map[string]interface{}  "Unauthorized"
// @Failure      409   {object}  map[string]interface{}  "Novel sudah dibookmark"
// @Failure      500   {object}  map[string]interface{}  "Gagal menambahkan bookmark"
// @Router       /bookmarks [post]
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

// GetNovelByBookmarks godoc
// @Summary      Daftar novel dari bookmark user
// @Description  Mengambil daftar novel yang dibookmark oleh user login, lengkap dengan author & chapters.
// @Tags         Bookmarks
// @Security     CookieAuth
// @Produce      json
// @Param        page   query     int  false  "Halaman (default 1)"
// @Param        limit  query     int  false  "Jumlah per halaman (default 20)"
// @Success      200    {object}  map[string]interface{}  "Daftar novel dari bookmark dengan meta pagination"
// @Failure      401    {object}  map[string]interface{}  "Unauthorized"
// @Failure      500    {object}  map[string]interface{}  "Gagal mengambil data"
// @Router       /bookmarks/novels [get]
func GetNovelByBookmarks(c *fiber.Ctx) error {
	// pakai JWT dari cookie supaya dapat role juga
	uid, role, err := getAuthFromAccessCookieNovel(c)
	if err != nil || uid == 0 {
		return bookmarkFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	// pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	// hitung total bookmark user
	var total int64
	if err := database.DB.
		Model(&models.Bookmark{}).
		Where("user_id = ?", uid).
		Count(&total).Error; err != nil {
		return bookmarkFail(c, http.StatusInternalServerError, "Gagal menghitung total bookmark", err)
	}

	if total == 0 {
		return bookmarkListOK(c, "Daftar novel dari bookmark", []any{}, page, limit, 0)
	}

	// ambil bookmark untuk halaman ini
	var bookmarks []models.Bookmark
	if err := database.DB.
		Where("user_id = ?", uid).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&bookmarks).Error; err != nil {
		return bookmarkFail(c, http.StatusInternalServerError, "Gagal mengambil bookmark", err)
	}
	if len(bookmarks) == 0 {
		return bookmarkListOK(c, "Daftar novel dari bookmark", []any{}, page, limit, total)
	}

	// kumpulkan novel_id unik
	novelIDSet := make(map[uint]struct{})
	for _, b := range bookmarks {
		if b.NovelID != 0 {
			novelIDSet[b.NovelID] = struct{}{}
		}
	}
	novelIDs := make([]uint, 0, len(novelIDSet))
	for id := range novelIDSet {
		novelIDs = append(novelIDs, id)
	}

	// ambil novel terkait
	var novels []models.Novel
	if err := database.DB.
		Preload("Genres").
		Preload("Tags").
		Where("id IN ?", novelIDs).
		Order("created_at DESC").
		Find(&novels).Error; err != nil {
		return bookmarkFail(c, http.StatusInternalServerError, "Gagal mengambil novel dari bookmark", err)
	}
	if len(novels) == 0 {
		return bookmarkListOK(c, "Daftar novel dari bookmark", []any{}, page, limit, total)
	}

	// kumpulkan author_id & map novel->author
	authorIDSet := make(map[uint]struct{})
	novelAuthorMap := make(map[uint]uint)
	for _, n := range novels {
		if n.AuthorID != 0 {
			authorIDSet[n.AuthorID] = struct{}{}
		}
		novelAuthorMap[n.ID] = n.AuthorID
	}

	authorIDs := make([]uint, 0, len(authorIDSet))
	for id := range authorIDSet {
		authorIDs = append(authorIDs, id)
	}

	type AuthorMini struct {
		ID             uint    `json:"id"`
		Username       string  `json:"username"`
		ProfilePicture *string `json:"profile_picture"`
	}

	authorsMap := make(map[uint]AuthorMini)
	if len(authorIDs) > 0 {
		var users []models.User
		if err := database.DB.
			Select("id, username, profile_picture").
			Where("id IN ?", authorIDs).
			Find(&users).Error; err != nil {
			return bookmarkFail(c, http.StatusInternalServerError, "Gagal mengambil data author", err)
		}
		for _, u := range users {
			authorsMap[u.ID] = AuthorMini{
				ID:             u.ID,
				Username:       u.Username,
				ProfilePicture: u.ProfilePicture,
			}
		}
	}

	// siapkan chapters: tergantung role & kepemilikan
	now := time.Now()
	chaptersByNovel := make(map[uint][]models.Chapter)

	if role == "admin" {
		// admin lihat semua chapter
		var chs []models.Chapter
		if err := database.DB.
			Where("novel_id IN ?", novelIDs).
			Order("novel_id ASC, order_no ASC").
			Find(&chs).Error; err != nil {
			return bookmarkFail(c, http.StatusInternalServerError, "Gagal mengambil chapters", err)
		}
		for _, ch := range chs {
			chaptersByNovel[ch.NovelID] = append(chaptersByNovel[ch.NovelID], ch)
		}
	} else {
		// user biasa: ambil chapter terbit
		var pubChs []models.Chapter
		if err := database.DB.
			Where("novel_id IN ? AND published_at IS NOT NULL AND published_at <= ?", novelIDs, now).
			Order("novel_id ASC, order_no ASC").
			Find(&pubChs).Error; err != nil {
			return bookmarkFail(c, http.StatusInternalServerError, "Gagal mengambil chapters terbit", err)
		}
		for _, ch := range pubChs {
			chaptersByNovel[ch.NovelID] = append(chaptersByNovel[ch.NovelID], ch)
		}

		// kalau novel itu milik user → override: lihat semua chapter
		ownedNovelIDs := make([]uint, 0)
		for nid, aid := range novelAuthorMap {
			if aid == uid {
				ownedNovelIDs = append(ownedNovelIDs, nid)
			}
		}
		if len(ownedNovelIDs) > 0 {
			var ownedChs []models.Chapter
			if err := database.DB.
				Where("novel_id IN ?", ownedNovelIDs).
				Order("novel_id ASC, order_no ASC").
				Find(&ownedChs).Error; err != nil {
				return bookmarkFail(c, http.StatusInternalServerError, "Gagal mengambil chapters milik author", err)
			}
			tmp := make(map[uint][]models.Chapter)
			for _, ch := range ownedChs {
				tmp[ch.NovelID] = append(tmp[ch.NovelID], ch)
			}
			for nid, list := range tmp {
				chaptersByNovel[nid] = list
			}
		}
	}

	// map novel by id untuk jaga urutan sesuai bookmark
	novelMap := make(map[uint]models.Novel)
	for _, n := range novels {
		novelMap[n.ID] = n
	}

	type BookmarkNovelItem struct {
		ID         uint             `json:"id"`
		Title      string           `json:"title"`
		Slug       string           `json:"slug"`
		Synopsis   *string          `json:"synopsis"`
		CoverImage *string          `json:"cover_image"`
		Status     string           `json:"status"`
		Author     *AuthorMini      `json:"author"`
		Genres     []models.Genre   `json:"genres"`
		Tags       []models.Tag     `json:"tags"`
		Chapters   []models.Chapter `json:"chapters"`
		CreatedAt  time.Time        `json:"created_at"`
		UpdatedAt  time.Time        `json:"updated_at"`
	}

	resp := make([]BookmarkNovelItem, 0, len(bookmarks))
	for _, b := range bookmarks {
		n, ok := novelMap[b.NovelID]
		if !ok {
			continue
		}

		var authorPtr *AuthorMini
		if am, ok := authorsMap[n.AuthorID]; ok {
			a := am
			authorPtr = &a
		}

		resp = append(resp, BookmarkNovelItem{
			ID:         n.ID,
			Title:      n.Title,
			Slug:       n.Slug,
			Synopsis:   n.Synopsis,
			CoverImage: n.CoverImage,
			Status:     n.Status,
			Author:     authorPtr,
			Genres:     n.Genres,
			Tags:       n.Tags,
			Chapters:   chaptersByNovel[n.ID],
			CreatedAt:  n.CreatedAt,
			UpdatedAt:  n.UpdatedAt,
		})
	}

	return bookmarkListOK(c, "Daftar novel dari bookmark", resp, page, limit, total)
}

// RemoveBookmark godoc
// @Summary      Hapus bookmark novel
// @Description  Menghapus bookmark novel tertentu milik user yang sedang login.
// @Tags         Bookmarks
// @Security     CookieAuth
// @Produce      json
// @Param        novel_id  path      int  true  "ID Novel yang dibookmark"
// @Success      200       {object}  map[string]interface{}  "Bookmark dihapus"
// @Failure      400       {object}  map[string]interface{}  "novel_id tidak valid"
// @Failure      401       {object}  map[string]interface{}  "Unauthorized"
// @Failure      404       {object}  map[string]interface{}  "Bookmark tidak ditemukan"
// @Failure      500       {object}  map[string]interface{}  "Gagal menghapus bookmark"
// @Router       /bookmarks/{novel_id} [delete]
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

// ListBookmarks godoc
// @Summary      List bookmark milik user
// @Description  Mengambil daftar bookmark milik user yang sedang login, termasuk info singkat novel.
// @Tags         Bookmarks
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Daftar bookmark dengan meta pagination"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil bookmark"
// @Router       /bookmarks [get]
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
