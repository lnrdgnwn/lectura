package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SaveFileTo(c *fiber.Ctx, fieldName, subfolder string) (string, error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExt := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
	}

	if !allowedExt[ext] {
		return "", fmt.Errorf("invalid file type: %s", ext)
	}

	base := "assets"
	imgBase := filepath.Join(base, "IMG")
	var targetDir string
	if strings.TrimSpace(subfolder) == "" {
		targetDir = imgBase
	} else {
		cleanSub := filepath.Clean(subfolder)
		cleanSub = strings.Trim(cleanSub, string(filepath.Separator))
		targetDir = filepath.Join(imgBase, cleanSub)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	dstPath := filepath.Join(targetDir, filename)

	if err := c.SaveFile(file, dstPath); err != nil {
		return "", err
	}

	baseURL := os.Getenv("BASE_URL")
	publicPath := "/" + filepath.ToSlash(strings.TrimPrefix(dstPath, "./"))
	if baseURL == "" {
		return publicPath, nil
	}
	return strings.TrimRight(baseURL, "/") + publicPath, nil
}

func SaveFile(c *fiber.Ctx, fieldName string, opts ...string) (string, error) {
	subfolder := ""

	if len(opts) == 0 {
		subfolder = ""
	} else if len(opts) == 1 {
		v := strings.TrimSpace(opts[0])
		if v == "" || strings.EqualFold(v, "false") || v == "0" {
			subfolder = ""
		} else if strings.EqualFold(v, "true") {
			subfolder = "cover"
		} else {
			subfolder = v
		}
	} else {
		first := strings.TrimSpace(opts[0])
		second := strings.TrimSpace(opts[1])

		if strings.EqualFold(second, "true") {
			subfolder = "profile_picture"
		} else if strings.EqualFold(first, "true") {
			subfolder = "cover"
		} else if first != "" && !strings.EqualFold(first, "false") {
			subfolder = first
		} else {
			subfolder = ""
		}
	}

	return SaveFileTo(c, fieldName, subfolder)
}
