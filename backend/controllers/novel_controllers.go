package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"final_project/database"
	"final_project/models"
	"final_project/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

func novelOK(c *fiber.Ctx, status int, msg string, data any) error {
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

func novelListOK(c *fiber.Ctx, msg string, data any, page, limit int, total int64) error {
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

func novelFail(c *fiber.Ctx, status int, msg string, err error) error {
	resp := fiber.Map{
		"success": false,
		"message": msg,
	}
	if err != nil && os.Getenv("APP_ENV") != "production" {
		resp["error"] = err.Error()
	}
	return c.Status(status).JSON(resp)
}

func getAuthFromAccessCookieNovel(c *fiber.Ctx) (uint, string, error) {
	enc := c.Cookies("access_token", "")
	if enc == "" {
		return 0, "", fiber.ErrUnauthorized
	}

	tokenStr, err := utils.Decrypt(enc)
	if err != nil || strings.TrimSpace(tokenStr) == "" {
		return 0, "", fiber.ErrUnauthorized
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return 0, "", fiber.ErrUnauthorized
	}

	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(secret), nil
	})
	if err != nil || tok == nil || !tok.Valid {
		return 0, "", fiber.ErrUnauthorized
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", fiber.ErrUnauthorized
	}

	var uid uint
	switch v := claims["id"].(type) {
	case float64:
		uid = uint(v)
	case string:
		if n, e := strconv.Atoi(v); e == nil {
			uid = uint(n)
		} else {
			return 0, "", fiber.ErrUnauthorized
		}
	default:
		return 0, "", fiber.ErrUnauthorized
	}

	role := ""
	if r, ok := claims["role"].(string); ok {
		role = strings.ToLower(r)
	}

	return uid, role, nil
}

func normalizeNovelStatus(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "ongoing", "completed", "hiatus":
		return s, true
	default:
		return "", false
	}
}

func parseUintSliceFromString(s string) ([]uint, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return []uint{}, nil
	}
	if strings.HasPrefix(s, "[") {
		var arr []uint
		if err := json.Unmarshal([]byte(s), &arr); err == nil {
			return arr, nil
		}
	}
	parts := strings.Split(s, ",")
	out := make([]uint, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		out = append(out, uint(n))
	}
	return out, nil
}

func parseIDsFromForm(c *fiber.Ctx, keys ...string) ([]uint, bool, error) {
	for _, k := range keys {
		v := c.FormValue(k)
		if v != "" {
			arr, err := parseUintSliceFromString(v)
			return arr, true, err
		}
	}
	return nil, false, nil
}

// GetNovel godoc
// @Summary      Ambil semua novel
// @Description  Mengambil semua novel dengan genres, tags, info author (username & profile_picture), dan chapters yang boleh diakses.
// @Description  - Admin: semua chapters.
// @Description  - User login: chapter published + semua chapter novel miliknya.
// @Description  - Tanpa login: hanya chapter yang sudah published.
// @Tags         Novels
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Daftar novel"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil novel"
// @Router       /novels [get]
func GetNovel(c *fiber.Ctx) error {
	var novels []models.Novel
	if err := database.DB.
		Preload("Genres").
		Preload("Tags").
		Order("created_at DESC").
		Find(&novels).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel", err)
	}

	if len(novels) == 0 {
		return novelOK(c, http.StatusOK, "Daftar novel", []any{})
	}

	uid, role, _ := getAuthFromAccessCookieNovel(c)

	authorIDSet := make(map[uint]struct{})
	novelIDSet := make(map[uint]struct{})
	ownedNovelIDSet := make(map[uint]struct{})

	for _, n := range novels {
		if n.AuthorID != 0 {
			authorIDSet[n.AuthorID] = struct{}{}
		}
		novelIDSet[n.ID] = struct{}{}
		if uid != 0 && n.AuthorID == uid {
			ownedNovelIDSet[n.ID] = struct{}{}
		}
	}

	authorIDs := make([]uint, 0, len(authorIDSet))
	for id := range authorIDSet {
		authorIDs = append(authorIDs, id)
	}

	novelIDs := make([]uint, 0, len(novelIDSet))
	for id := range novelIDSet {
		novelIDs = append(novelIDs, id)
	}

	ownedNovelIDs := make([]uint, 0, len(ownedNovelIDSet))
	for id := range ownedNovelIDSet {
		ownedNovelIDs = append(ownedNovelIDs, id)
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
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil data author", err)
		}

		for _, u := range users {
			authorsMap[u.ID] = AuthorMini{
				ID:             u.ID,
				Username:       u.Username,
				ProfilePicture: u.ProfilePicture,
			}
		}
	}

	chaptersByNovel := make(map[uint][]models.Chapter)

	if role == "admin" {
		var chs []models.Chapter
		if err := database.DB.
			Where("novel_id IN ?", novelIDs).
			Order("novel_id ASC, order_no ASC").
			Find(&chs).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil chapters", err)
		}
		for _, ch := range chs {
			chaptersByNovel[ch.NovelID] = append(chaptersByNovel[ch.NovelID], ch)
		}
	} else {
		now := time.Now()

		var published []models.Chapter
		if err := database.DB.
			Where("novel_id IN ? AND published_at IS NOT NULL AND published_at <= ?", novelIDs, now).
			Order("novel_id ASC, order_no ASC").
			Find(&published).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil chapters terbit", err)
		}
		for _, ch := range published {
			chaptersByNovel[ch.NovelID] = append(chaptersByNovel[ch.NovelID], ch)
		}

		if uid != 0 && len(ownedNovelIDs) > 0 {
			var ownedCh []models.Chapter
			if err := database.DB.
				Where("novel_id IN ?", ownedNovelIDs).
				Order("novel_id ASC, order_no ASC").
				Find(&ownedCh).Error; err != nil {
				return novelFail(c, http.StatusInternalServerError, "Gagal mengambil chapters milik author", err)
			}

			tempOwned := make(map[uint][]models.Chapter)
			for _, ch := range ownedCh {
				tempOwned[ch.NovelID] = append(tempOwned[ch.NovelID], ch)
			}
			for nid, list := range tempOwned {
				chaptersByNovel[nid] = list
			}
		}
	}

	type NovelItem struct {
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

	resp := make([]NovelItem, 0, len(novels))
	for _, n := range novels {
		var authorPtr *AuthorMini
		if am, ok := authorsMap[n.AuthorID]; ok {
			a := am
			authorPtr = &a
		}

		resp = append(resp, NovelItem{
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

	return novelOK(c, http.StatusOK, "Daftar novel", resp)
}

// GetNovelByID godoc
// @Summary      Detail novel
// @Description  Mengambil detail novel berdasarkan ID, termasuk genres, tags, author (username & profile_picture), dan daftar chapters.
// @Description  - Admin/Author: semua chapters (termasuk draft).
// @Description  - User lain/tanpa login: hanya chapter yang sudah published.
// @Tags         Novels
// @Produce      json
// @Param        id   path      int  true  "ID Novel"
// @Success      200  {object}  map[string]interface{}  "Detail novel"
// @Failure      404  {object}  map[string]interface{}  "Novel tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil novel"
// @Router       /novels/{id} [get]
func GetNovelByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.
		Preload("Genres").
		Preload("Tags").
		First(&novel, id).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return novelFail(c, http.StatusNotFound, "Novel tidak ditemukan", nil)
		}
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel", err)
	}

	var author models.User
	if err := database.DB.
		Select("id, username, profile_picture").
		First(&author, novel.AuthorID).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return novelFail(c, http.StatusInternalServerError, "Data author untuk novel ini tidak ditemukan", nil)
		}
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil data author", err)
	}

	type AuthorMini struct {
		ID             uint    `json:"id"`
		Username       string  `json:"username"`
		ProfilePicture *string `json:"profile_picture"`
	}

	authorMini := AuthorMini{
		ID:             author.ID,
		Username:       author.Username,
		ProfilePicture: author.ProfilePicture,
	}

	uid, role, _ := getAuthFromAccessCookieNovel(c)

	var chapters []models.Chapter

	chQuery := database.DB.
		Model(&models.Chapter{}).
		Where("novel_id = ?", novel.ID).
		Order("order_no ASC")

	if role == "admin" || uid == novel.AuthorID {
		if err := chQuery.Find(&chapters).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil daftar chapter", err)
		}
	} else {
		if err := chQuery.
			Where("published_at IS NOT NULL AND published_at <= ?", time.Now()).
			Find(&chapters).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil daftar chapter terbit", err)
		}
	}

	type NovelDetailResponse struct {
		ID         uint             `json:"id"`
		Title      string           `json:"title"`
		Slug       string           `json:"slug"`
		Synopsis   *string          `json:"synopsis"`
		CoverImage *string          `json:"cover_image"`
		Status     string           `json:"status"`
		Author     AuthorMini       `json:"author"`
		Genres     []models.Genre   `json:"genres"`
		Tags       []models.Tag     `json:"tags"`
		Chapters   []models.Chapter `json:"chapters"`
		CreatedAt  time.Time        `json:"created_at"`
		UpdatedAt  time.Time        `json:"updated_at"`
	}

	resp := NovelDetailResponse{
		ID:         novel.ID,
		Title:      novel.Title,
		Slug:       novel.Slug,
		Synopsis:   novel.Synopsis,
		CoverImage: novel.CoverImage,
		Status:     novel.Status,
		Author:     authorMini,
		Genres:     novel.Genres,
		Tags:       novel.Tags,
		Chapters:   chapters,
		CreatedAt:  novel.CreatedAt,
		UpdatedAt:  novel.UpdatedAt,
	}

	return novelOK(c, http.StatusOK, "Detail novel", resp)
}

// GetNovelByGenreID godoc
// @Summary      Daftar novel berdasarkan genre
// @Description  Mengambil daftar novel untuk genre tertentu, termasuk genres, tags, author (username & profile_picture), dan jumlah chapter terbit.
// @Tags         Novels
// @Produce      json
// @Param        genre_id  query     int  true   "ID Genre (bisa juga dari path: /novels/genre/{genre_id})"
// @Param        page      query     int  false  "Halaman (default: 1)"
// @Param        limit     query     int  false  "Jumlah item per halaman (default: 20)"
// @Success      200       {object}  map[string]interface{}  "Daftar novel berdasarkan genre dengan meta pagination"
// @Failure      400       {object}  map[string]interface{}  "genre_id tidak valid / wajib diisi"
// @Failure      404       {object}  map[string]interface{}  "Genre tidak ditemukan"
// @Failure      500       {object}  map[string]interface{}  "Gagal mengambil data genre / novel"
// @Router       /novels/by-genre/:genre_id [get]
func GetNovelByGenreID(c *fiber.Ctx) error {
	gidStr := c.Params("genre_id")
	if gidStr == "" {
		gidStr = c.Params("id")
	}
	if gidStr == "" {
		gidStr = c.Query("genre_id", "")
	}
	if strings.TrimSpace(gidStr) == "" {
		return novelFail(c, http.StatusBadRequest, "genre_id wajib diisi", nil)
	}

	genreIDInt, err := strconv.Atoi(gidStr)
	if err != nil || genreIDInt <= 0 {
		return novelFail(c, http.StatusBadRequest, "genre_id tidak valid", err)
	}
	genreID := uint(genreIDInt)

	var g models.Genre
	if err := database.DB.First(&g, genreID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return novelFail(c, http.StatusNotFound, "Genre tidak ditemukan", nil)
		}
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil data genre", err)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	base := database.DB.
		Model(&models.Novel{}).
		Joins("JOIN novel_genres ng ON ng.novel_id = novels.id").
		Where("ng.genre_id = ?", genreID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal menghitung total novel untuk genre ini", err)
	}

	var novels []models.Novel
	if err := base.
		Preload("Genres").
		Preload("Tags").
		Order("novels.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&novels).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel untuk genre ini", err)
	}

	if len(novels) == 0 {
		return novelListOK(c, "Daftar novel berdasarkan genre", []any{}, page, limit, total)
	}

	authorIDSet := make(map[uint]struct{})
	novelIDs := make([]uint, 0, len(novels))
	for _, n := range novels {
		if n.AuthorID != 0 {
			authorIDSet[n.AuthorID] = struct{}{}
		}
		novelIDs = append(novelIDs, n.ID)
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
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil data author", err)
		}
		for _, u := range users {
			authorsMap[u.ID] = AuthorMini{
				ID:             u.ID,
				Username:       u.Username,
				ProfilePicture: u.ProfilePicture,
			}
		}
	}

	type ChapterCountRow struct {
		NovelID uint  `json:"novel_id"`
		Count   int64 `json:"count"`
	}

	chapterCounts := make(map[uint]int64)
	if len(novelIDs) > 0 {
		var rows []ChapterCountRow
		if err := database.DB.
			Model(&models.Chapter{}).
			Select("novel_id, COUNT(*) as count").
			Where("novel_id IN ? AND published_at IS NOT NULL AND published_at <= ?", novelIDs, time.Now()).
			Group("novel_id").
			Scan(&rows).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal menghitung jumlah chapter", err)
		}
		for _, r := range rows {
			chapterCounts[r.NovelID] = r.Count
		}
	}

	type GenreNovelItem struct {
		ID           uint           `json:"id"`
		Title        string         `json:"title"`
		Slug         string         `json:"slug"`
		Synopsis     *string        `json:"synopsis"`
		CoverImage   *string        `json:"cover_image"`
		Status       string         `json:"status"`
		Author       *AuthorMini    `json:"author"`
		Genres       []models.Genre `json:"genres"`
		Tags         []models.Tag   `json:"tags"`
		ChapterCount int64          `json:"chapter_count"`
		CreatedAt    time.Time      `json:"created_at"`
		UpdatedAt    time.Time      `json:"updated_at"`
	}

	resp := make([]GenreNovelItem, 0, len(novels))
	for _, n := range novels {
		var authorPtr *AuthorMini
		if am, ok := authorsMap[n.AuthorID]; ok {
			a := am
			authorPtr = &a
		}

		resp = append(resp, GenreNovelItem{
			ID:           n.ID,
			Title:        n.Title,
			Slug:         n.Slug,
			Synopsis:     n.Synopsis,
			CoverImage:   n.CoverImage,
			Status:       n.Status,
			Author:       authorPtr,
			Genres:       n.Genres,
			Tags:         n.Tags,
			ChapterCount: chapterCounts[n.ID],
			CreatedAt:    n.CreatedAt,
			UpdatedAt:    n.UpdatedAt,
		})
	}

	return novelListOK(c, "Daftar novel berdasarkan genre", resp, page, limit, total)
}

// GetMyNovels godoc
// @Summary      Daftar novel milik user login
// @Description  Mengambil daftar novel yang ditulis oleh user yang sedang login, termasuk genres, tags, author (username & profile_picture), dan jumlah chapter terbit.
// @Tags         Novels
// @Security     CookieAuth
// @Produce      json
// @Param        q       query     string  false  "Filter judul/slug (pencarian bebas)"
// @Param        status  query     string  false  "Filter status (ongoing, completed, hiatus)"
// @Param        page    query     int     false  "Halaman (default: 1)"
// @Param        limit   query     int     false  "Jumlah item per halaman (default: 20)"
// @Success      200     {object}  map[string]interface{}  "Daftar novel milik user dengan meta pagination"
// @Failure      400     {object}  map[string]interface{}  "Status tidak valid"
// @Failure      401     {object}  map[string]interface{}  "Unauthorized"
// @Failure      500     {object}  map[string]interface{}  "Gagal mengambil novel milik user"
// @Router       /users/me/novels [get]
func GetMyNovels(c *fiber.Ctx) error {
	uid, _, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return novelFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	q := strings.TrimSpace(c.Query("q", ""))
	status := strings.TrimSpace(c.Query("status", ""))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	countQB := database.DB.
		Model(&models.Novel{}).
		Where("author_id = ?", uid)

	if q != "" {
		like := "%" + q + "%"
		countQB = countQB.Where("title LIKE ? OR slug LIKE ?", like, like)
	}

	if status != "" {
		if ns, ok := normalizeNovelStatus(status); ok {
			countQB = countQB.Where("status = ?", ns)
		} else {
			return novelFail(c, http.StatusBadRequest, "Status tidak valid (gunakan: ongoing, completed, hiatus)", nil)
		}
	}

	var total int64
	if err := countQB.Count(&total).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal menghitung total novel milik user", err)
	}

	dataQB := database.DB.
		Model(&models.Novel{}).
		Where("author_id = ?", uid).
		Preload("Genres").
		Preload("Tags")

	if q != "" {
		like := "%" + q + "%"
		dataQB = dataQB.Where("title LIKE ? OR slug LIKE ?", like, like)
	}
	if status != "" {
		ns, _ := normalizeNovelStatus(status)
		dataQB = dataQB.Where("status = ?", ns)
	}

	var novels []models.Novel
	if err := dataQB.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&novels).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel milik user", err)
	}

	if len(novels) == 0 {
		return novelListOK(c, "Daftar novel milik user", []any{}, page, limit, total)
	}

	type AuthorMini struct {
		ID             uint    `json:"id"`
		Username       string  `json:"username"`
		ProfilePicture *string `json:"profile_picture"`
	}

	var author models.User
	if err := database.DB.
		Select("id, username, profile_picture").
		First(&author, uid).Error; err != nil {

		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil data author", err)
	}

	authorMini := AuthorMini{
		ID:             author.ID,
		Username:       author.Username,
		ProfilePicture: author.ProfilePicture,
	}

	novelIDs := make([]uint, 0, len(novels))
	for _, n := range novels {
		novelIDs = append(novelIDs, n.ID)
	}

	type ChapterCountRow struct {
		NovelID uint  `json:"novel_id"`
		Count   int64 `json:"count"`
	}

	chapterCounts := make(map[uint]int64)

	if len(novelIDs) > 0 {
		var rows []ChapterCountRow
		if err := database.DB.
			Model(&models.Chapter{}).
			Select("novel_id, COUNT(*) as count").
			Where("novel_id IN ? AND published_at IS NOT NULL AND published_at <= ?", novelIDs, time.Now()).
			Group("novel_id").
			Scan(&rows).Error; err != nil {

			return novelFail(c, http.StatusInternalServerError, "Gagal menghitung jumlah chapter", err)
		}

		for _, r := range rows {
			chapterCounts[r.NovelID] = r.Count
		}
	}

	type MyNovelItem struct {
		ID            uint           `json:"id"`
		Title         string         `json:"title"`
		Slug          string         `json:"slug"`
		Synopsis      *string        `json:"synopsis"`
		CoverImage    *string        `json:"cover_image"`
		Status        string         `json:"status"`
		Author        AuthorMini     `json:"author"`
		Genres        []models.Genre `json:"genres"`
		Tags          []models.Tag   `json:"tags"`
		ChapterCount  int64          `json:"chapter_count"`
		CreatedAt     time.Time      `json:"created_at"`
		UpdatedAt     time.Time      `json:"updated_at"`
	}

	resp := make([]MyNovelItem, 0, len(novels))
	for _, n := range novels {
		resp = append(resp, MyNovelItem{
			ID:           n.ID,
			Title:        n.Title,
			Slug:         n.Slug,
			Synopsis:     n.Synopsis,
			CoverImage:   n.CoverImage,
			Status:       n.Status,
			Author:       authorMini,
			Genres:       n.Genres,
			Tags:         n.Tags,
			ChapterCount: chapterCounts[n.ID],
			CreatedAt:    n.CreatedAt,
			UpdatedAt:    n.UpdatedAt,
		})
	}

	return novelListOK(c, "Daftar novel milik user", resp, page, limit, total)
}

// PostNovel godoc
// @Summary      Buat novel baru
// @Description  Membuat novel baru milik user login, dengan optional genres & tags. Status hanya boleh ongoing, completed, atau hiatus.
// @Tags         Novels
// @Security     CookieAuth
// @Accept       json
// @Accept       multipart/form-data
// @Produce      json
// @Param        title       formData  string  true   "Judul novel"             example("Judul Keren")
// @Param        slug        formData  string  true   "Slug unik novel"         example("judul-keren")
// @Param        synopsis    formData  string  false  "Sinopsis novel"
// @Param        cover_url   formData  string  false  "URL cover (jika tidak upload file)"
// @Param        cover_image formData  file    false  "File cover (opsional, multipart)"
// @Param        status      formData  string  false  "Status (ongoing, completed, hiatus). Default: ongoing"
// @Param        genre_ids   formData  []int   false  "Array ID genre (maksimal 1 genre)"
// @Param        tag_ids     formData  []int   false  "Array ID tag"
// @Success      201         {object}  map[string]interface{}  "Novel dibuat"
// @Failure      400         {object}  map[string]interface{}  "Payload tidak valid / field wajib kosong / status tidak valid"
// @Failure      401         {object}  map[string]interface{}  "Unauthorized"
// @Failure      500         {object}  map[string]interface{}  "Gagal membuat novel"
// @Router       /novels [post]
func PostNovel(c *fiber.Ctx) error {
	uid, _, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return novelFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	var payload struct {
		Title    string  `json:"title"`
		Slug     string  `json:"slug"`
		Synopsis *string `json:"synopsis"`
		CoverURL *string `json:"cover_url"`
		Status   *string `json:"status"`
		GenreIDs []uint  `json:"genre_ids"`
		TagIDs   []uint  `json:"tag_ids"`
	}

	ct := c.Get("Content-Type")
	if strings.Contains(ct, "application/json") || c.Is("json") {
		if err := c.BodyParser(&payload); err != nil {
			return novelFail(c, http.StatusBadRequest, "Payload JSON tidak valid", err)
		}
	} else {
		payload.Title = c.FormValue("title")
		payload.Slug = c.FormValue("slug")
		if s := c.FormValue("synopsis"); s != "" {
			payload.Synopsis = &s
		}
		if cv := c.FormValue("cover_image"); cv != "" {
			payload.CoverURL = &cv
		}
		if st := c.FormValue("status"); st != "" {
			payload.Status = &st
		}
		if arr, present, perr := parseIDsFromForm(c, "genre_ids", "genre_ids[]"); perr != nil {
			return novelFail(c, http.StatusBadRequest, "genre_ids tidak valid", perr)
		} else if present {
			payload.GenreIDs = arr
		}
		if arr, present, perr := parseIDsFromForm(c, "tag_ids", "tag_ids[]"); perr != nil {
			return novelFail(c, http.StatusBadRequest, "tag_ids tidak valid", perr)
		} else if present {
			payload.TagIDs = arr
		}
		if _, ferr := c.FormFile("cover_image"); ferr == nil {
			if path, err := utils.SaveFile(c, "cover_image", "cover"); err == nil {
				payload.CoverURL = &path
			} else {
				return novelFail(c, http.StatusBadRequest, "Gagal menyimpan cover", err)
			}
		}
	}

	if strings.TrimSpace(payload.Title) == "" || strings.TrimSpace(payload.Slug) == "" {
		return novelFail(c, http.StatusBadRequest, "title dan slug wajib diisi", nil)
	}
	if len(payload.GenreIDs) > 1 {
		return novelFail(c, http.StatusBadRequest, "hanya boleh memasukkan maksimal 1 genre", nil)
	}

	status := "ongoing"
	if payload.Status != nil && strings.TrimSpace(*payload.Status) != "" {
		if ns, ok := normalizeNovelStatus(*payload.Status); ok {
			status = ns
		} else {
			return novelFail(c, http.StatusBadRequest, "Status tidak valid (gunakan: ongoing, completed, hiatus)", nil)
		}
	}

	n := models.Novel{
		AuthorID:   uid,
		Title:      strings.TrimSpace(payload.Title),
		Slug:       strings.TrimSpace(payload.Slug),
		Synopsis:   payload.Synopsis,
		CoverImage: payload.CoverURL,
		Status:     status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if len(payload.GenreIDs) == 1 {
		var genre models.Genre
		if err := database.DB.First(&genre, payload.GenreIDs[0]).Error; err != nil {
			return novelFail(c, http.StatusBadRequest, "genre tidak ditemukan", err)
		}
		n.Genres = []models.Genre{genre}
	}

	if len(payload.TagIDs) > 0 {
		var tags []models.Tag
		if err := database.DB.Where("id IN ?", payload.TagIDs).Find(&tags).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil tags", err)
		}
		if len(tags) != len(payload.TagIDs) {
			return novelFail(c, http.StatusBadRequest, "salah satu tag_id tidak ditemukan", nil)
		}
		n.Tags = tags
	}

	if err := database.DB.Create(&n).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal membuat novel", err)
	}
	_ = database.DB.Preload("Genres").Preload("Tags").First(&n, n.ID)

	return novelOK(c, http.StatusCreated, "Novel dibuat", n)
}

// UpdateNovel godoc
// @Summary      Perbarui novel
// @Description  Memperbarui novel milik author (atau admin), termasuk title, synopsis, cover, status, genre, dan tag. Mengembalikan detail lengkap novel + author + chapters.
// @Tags         Novels
// @Security     CookieAuth
// @Accept       json
// @Accept       multipart/form-data
// @Produce      json
// @Param        id           path      int      true   "ID Novel"
// @Param        title        formData  string   false  "Judul novel"
// @Param        synopsis     formData  string   false  "Sinopsis novel"
// @Param        cover_image  formData  string   false  "URL cover baru atau path file yang diupload"
// @Param        cover_image  formData  file     false  "File cover (opsional, multipart)"
// @Param        status       formData  string   false  "Status (ongoing, completed, hiatus)"
// @Param        genre_ids    formData  []int    false  "Array ID genre (maksimal 1 genre). Jika kosong [] akan menghapus semua genre."
// @Param        tag_ids      formData  []int    false  "Array ID tag. Jika kosong [] akan menghapus semua tag."
// @Success      200          {object}  map[string]interface{}  "Novel diperbarui (detail lengkap)"
// @Failure      400          {object}  map[string]interface{}  "Payload tidak valid / status tidak valid / genre/tag tidak ditemukan"
// @Failure      401          {object}  map[string]interface{}  "Unauthorized"
// @Failure      403          {object}  map[string]interface{}  "Tidak berwenang mengubah novel"
// @Failure      404          {object}  map[string]interface{}  "Novel tidak ditemukan"
// @Failure      500          {object}  map[string]interface{}  "Gagal menyimpan perubahan / memuat ulang novel"
// @Router       /novels/{id} [put]
func UpdateNovel(c *fiber.Ctx) error {
	uid, role, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return novelFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.Preload("Genres").Preload("Tags").First(&novel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return novelFail(c, http.StatusNotFound, "Novel tidak ditemukan", nil)
		}
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel", err)
	}

	if novel.AuthorID != uid && role != "admin" {
		return novelFail(c, http.StatusForbidden, "Tidak berwenang mengubah novel ini", nil)
	}

	var payload struct {
		Title      *string `json:"title"`
		Synopsis   *string `json:"synopsis"`
		CoverImage *string `json:"cover_image"`
		Status     *string `json:"status"`
		GenreIDs   *[]uint `json:"genre_ids"`
		TagIDs     *[]uint `json:"tag_ids"`
	}

	ct := c.Get("Content-Type")
	if strings.Contains(ct, "application/json") || c.Is("json") {
		if err := c.BodyParser(&payload); err != nil {
			return novelFail(c, http.StatusBadRequest, "Payload JSON tidak valid", err)
		}
	} else {
		if v := c.FormValue("title"); v != "" {
			payload.Title = &v
		}
		if v := c.FormValue("synopsis"); v != "" {
			payload.Synopsis = &v
		}
		if v := c.FormValue("cover_image"); v != "" {
			payload.CoverImage = &v
		}
		if v := c.FormValue("status"); v != "" {
			payload.Status = &v
		}
		if arr, present, perr := parseIDsFromForm(c, "genre_ids", "genre_ids[]"); perr != nil {
			return novelFail(c, http.StatusBadRequest, "genre_ids tidak valid", perr)
		} else if present {
			payload.GenreIDs = &arr
		}
		if arr, present, perr := parseIDsFromForm(c, "tag_ids", "tag_ids[]"); perr != nil {
			return novelFail(c, http.StatusBadRequest, "tag_ids tidak valid", perr)
		} else if present {
			payload.TagIDs = &arr
		}
		if _, ferr := c.FormFile("cover_image"); ferr == nil {
			if path, err := utils.SaveFile(c, "cover_image", "cover"); err == nil {
				payload.CoverImage = &path
			} else {
				return novelFail(c, http.StatusBadRequest, "Gagal menyimpan cover", err)
			}
		}
	}

	if payload.GenreIDs != nil && len(*payload.GenreIDs) > 1 {
		return novelFail(c, http.StatusBadRequest, "hanya boleh memasukkan maksimal 1 genre", nil)
	}

	if payload.Title != nil {
		novel.Title = strings.TrimSpace(*payload.Title)
	}
	if payload.Synopsis != nil {
		novel.Synopsis = payload.Synopsis
	}
	if payload.CoverImage != nil {
		novel.CoverImage = payload.CoverImage
	}
	if payload.Status != nil {
		s := strings.TrimSpace(*payload.Status)
		if s == "" {
			return novelFail(c, http.StatusBadRequest, "Status tidak boleh kosong", nil)
		}
		if ns, ok := normalizeNovelStatus(s); ok {
			novel.Status = ns
		} else {
			return novelFail(c, http.StatusBadRequest, "Status tidak valid (gunakan: ongoing, completed, hiatus)", nil)
		}
	}

	if payload.GenreIDs != nil {
		if len(*payload.GenreIDs) == 0 {
			_ = database.DB.Model(&novel).Association("Genres").Clear()
		} else {
			var genres []models.Genre
			if err := database.DB.Where("id IN ?", *payload.GenreIDs).Find(&genres).Error; err != nil {
				return novelFail(c, http.StatusInternalServerError, "Gagal mengambil genre", err)
			}
			if len(genres) != len(*payload.GenreIDs) {
				return novelFail(c, http.StatusBadRequest, "salah satu genre_id tidak ditemukan", nil)
			}
			_ = database.DB.Model(&novel).Association("Genres").Replace(&genres)
		}
	}

	if payload.TagIDs != nil {
		if len(*payload.TagIDs) == 0 {
			_ = database.DB.Model(&novel).Association("Tags").Clear()
		} else {
			var tags []models.Tag
			if err := database.DB.Where("id IN ?", *payload.TagIDs).Find(&tags).Error; err != nil {
				return novelFail(c, http.StatusInternalServerError, "Gagal mengambil tags", err)
			}
			if len(tags) != len(*payload.TagIDs) {
				return novelFail(c, http.StatusBadRequest, "salah satu tag_id tidak ditemukan", nil)
			}
			_ = database.DB.Model(&novel).Association("Tags").Replace(&tags)
		}
	}

	novel.UpdatedAt = time.Now()
	if err := database.DB.Save(&novel).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal menyimpan perubahan", err)
	}

	if err := database.DB.
		Preload("Genres").
		Preload("Tags").
		First(&novel, novel.ID).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal memuat ulang novel", err)
	}

	type AuthorMini struct {
		ID             uint    `json:"id"`
		Username       string  `json:"username"`
		ProfilePicture *string `json:"profile_picture"`
	}

	var author models.User
	if err := database.DB.
		Select("id, username, profile_picture").
		First(&author, novel.AuthorID).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil data author", err)
	}

	authorMini := AuthorMini{
		ID:             author.ID,
		Username:       author.Username,
		ProfilePicture: author.ProfilePicture,
	}

	var chapters []models.Chapter
	chQuery := database.DB.
		Model(&models.Chapter{}).
		Where("novel_id = ?", novel.ID).
		Order("order_no ASC")

	if role == "admin" || uid == novel.AuthorID {
		if err := chQuery.Find(&chapters).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil daftar chapter", err)
		}
	} else {
		if err := chQuery.
			Where("published_at IS NOT NULL AND published_at <= ?", time.Now()).
			Find(&chapters).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil daftar chapter terbit", err)
		}
	}

	type NovelDetailResponse struct {
		ID         uint             `json:"id"`
		Title      string           `json:"title"`
		Slug       string           `json:"slug"`
		Synopsis   *string          `json:"synopsis"`
		CoverImage *string          `json:"cover_image"`
		Status     string           `json:"status"`
		Author     AuthorMini       `json:"author"`
		Genres     []models.Genre   `json:"genres"`
		Tags       []models.Tag     `json:"tags"`
		Chapters   []models.Chapter `json:"chapters"`
		CreatedAt  time.Time        `json:"created_at"`
		UpdatedAt  time.Time        `json:"updated_at"`
	}

	resp := NovelDetailResponse{
		ID:         novel.ID,
		Title:      novel.Title,
		Slug:       novel.Slug,
		Synopsis:   novel.Synopsis,
		CoverImage: novel.CoverImage,
		Status:     novel.Status,
		Author:     authorMini,
		Genres:     novel.Genres,
		Tags:       novel.Tags,
		Chapters:   chapters,
		CreatedAt:  novel.CreatedAt,
		UpdatedAt:  novel.UpdatedAt,
	}

	return novelOK(c, http.StatusOK, "Novel diperbarui", resp)
}

// DeleteNovel godoc
// @Summary      Hapus novel
// @Description  Menghapus novel milik author atau oleh admin.
// @Tags         Novels
// @Security     CookieAuth
// @Produce      json
// @Param        id   path      int  true  "ID Novel"
// @Success      200  {object}  map[string]interface{}  "Novel dihapus"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Tidak berwenang menghapus novel"
// @Failure      404  {object}  map[string]interface{}  "Novel tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal menghapus novel"
// @Router       /novels/{id} [delete]
func DeleteNovel(c *fiber.Ctx) error {
	uid, role, err := getAuthFromAccessCookieNovel(c)
	if err != nil {
		return novelFail(c, http.StatusUnauthorized, "Unauthorized", err)
	}

	id := c.Params("id")
	var novel models.Novel
	if err := database.DB.First(&novel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return novelFail(c, http.StatusNotFound, "Novel tidak ditemukan", nil)
		}
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel", err)
	}

	if novel.AuthorID != uid && role != "admin" {
		return novelFail(c, http.StatusForbidden, "Tidak berwenang menghapus novel ini", nil)
	}

	if err := database.DB.Delete(&novel).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal menghapus novel", err)
	}
	return novelOK(c, http.StatusOK, "Novel dihapus", nil)
}

// SearchNovels godoc
// @Summary      Cari novel
// @Description  Mencari novel berdasarkan title atau slug. Mengembalikan list novel dengan genres, tags, author (username & profile_picture), dan chapters yang terlihat sesuai role.
// @Tags         Novels
// @Produce      json
// @Param        q      query     string  false  "Kata kunci (cari di title atau slug)"
// @Param        page   query     int     false  "Halaman (default: 1)"
// @Param        limit  query     int     false  "Jumlah item per halaman (default: 20)"
// @Success      200    {object}  map[string]interface{}  "Hasil pencarian novel dengan meta pagination"
// @Failure      500    {object}  map[string]interface{}  "Gagal mengambil hasil pencarian"
// @Router       /novels/search [get]
func SearchNovels(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q", ""))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	countQB := database.DB.Model(&models.Novel{})
	if q != "" {
		like := "%" + q + "%"
		countQB = countQB.Where("title LIKE ? OR slug LIKE ?", like, like)
	}

	var total int64
	if err := countQB.Count(&total).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal menghitung total novel", err)
	}

	dataQB := database.DB.
		Model(&models.Novel{}).
		Preload("Genres").
		Preload("Tags")

	if q != "" {
		like := "%" + q + "%"
		dataQB = dataQB.Where("title LIKE ? OR slug LIKE ?", like, like)
	}

	var novels []models.Novel
	if err := dataQB.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&novels).Error; err != nil {
		return novelFail(c, http.StatusInternalServerError, "Gagal mengambil novel", err)
	}

	if len(novels) == 0 {
		return novelListOK(c, "Hasil pencarian novel", []any{}, page, limit, total)
	}

	uid, role, _ := getAuthFromAccessCookieNovel(c)

	authorIDSet := make(map[uint]struct{})
	novelIDSet := make(map[uint]struct{})
	ownedNovelIDSet := make(map[uint]struct{})

	for _, n := range novels {
		if n.AuthorID != 0 {
			authorIDSet[n.AuthorID] = struct{}{}
		}
		novelIDSet[n.ID] = struct{}{}
		if uid != 0 && n.AuthorID == uid {
			ownedNovelIDSet[n.ID] = struct{}{}
		}
	}

	authorIDs := make([]uint, 0, len(authorIDSet))
	for id := range authorIDSet {
		authorIDs = append(authorIDs, id)
	}

	novelIDs := make([]uint, 0, len(novelIDSet))
	for id := range novelIDSet {
		novelIDs = append(novelIDs, id)
	}

	ownedNovelIDs := make([]uint, 0, len(ownedNovelIDSet))
	for id := range ownedNovelIDSet {
		ownedNovelIDs = append(ownedNovelIDs, id)
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
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil data author", err)
		}

		for _, u := range users {
			authorsMap[u.ID] = AuthorMini{
				ID:             u.ID,
				Username:       u.Username,
				ProfilePicture: u.ProfilePicture,
			}
		}
	}

	chaptersByNovel := make(map[uint][]models.Chapter)

	if role == "admin" {
		var chs []models.Chapter
		if err := database.DB.
			Where("novel_id IN ?", novelIDs).
			Order("novel_id ASC, order_no ASC").
			Find(&chs).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil chapters", err)
		}
		for _, ch := range chs {
			chaptersByNovel[ch.NovelID] = append(chaptersByNovel[ch.NovelID], ch)
		}
	} else {
		now := time.Now()

		var published []models.Chapter
		if err := database.DB.
			Where("novel_id IN ? AND published_at IS NOT NULL AND published_at <= ?", novelIDs, now).
			Order("novel_id ASC, order_no ASC").
			Find(&published).Error; err != nil {
			return novelFail(c, http.StatusInternalServerError, "Gagal mengambil chapters terbit", err)
		}
		for _, ch := range published {
			chaptersByNovel[ch.NovelID] = append(chaptersByNovel[ch.NovelID], ch)
		}

		if uid != 0 && len(ownedNovelIDs) > 0 {
			var ownedCh []models.Chapter
			if err := database.DB.
				Where("novel_id IN ?", ownedNovelIDs).
				Order("novel_id ASC, order_no ASC").
				Find(&ownedCh).Error; err != nil {
				return novelFail(c, http.StatusInternalServerError, "Gagal mengambil chapters milik author", err)
			}

			tmp := make(map[uint][]models.Chapter)
			for _, ch := range ownedCh {
				tmp[ch.NovelID] = append(tmp[ch.NovelID], ch)
			}
			for nid, list := range tmp {
				chaptersByNovel[nid] = list
			}
		}
	}

	type NovelItem struct {
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

	resp := make([]NovelItem, 0, len(novels))
	for _, n := range novels {
		var authorPtr *AuthorMini
		if am, ok := authorsMap[n.AuthorID]; ok {
			a := am
			authorPtr = &a
		}

		resp = append(resp, NovelItem{
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

	return novelListOK(c, "Hasil pencarian novel", resp, page, limit, total)
}
