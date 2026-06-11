package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/response"
	"github.com/wind7vn/fnb_be/pkg/config"
)

func UploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Chưa đính kèm file hình", err))
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
		return response.Error(c, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Định dạng không hỗ trợ (chỉ hình ảnh)", nil))
	}

	// Ensure uploads directory exists
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		return response.Error(c, errors.NewInternalServer(err))
	}

	filename := fmt.Sprintf("img_%d%s", time.Now().UnixNano(), ext)
	savePath := fmt.Sprintf("./uploads/%s", filename)

	if err := c.SaveFile(file, savePath); err != nil {
		return response.Error(c, errors.NewInternalServer(err))
	}

	domain := config.AppConfig.AppDomain
	if domain == "" {
		domain = c.BaseURL()
	}

	// Ensure domain starts with a protocol
	urlPrefix := domain
	if len(urlPrefix) > 0 && !strings.HasPrefix(urlPrefix, "http://") && !strings.HasPrefix(urlPrefix, "https://") {
		urlPrefix = c.Protocol() + "://" + urlPrefix
	}

	url := fmt.Sprintf("%s/uploads/%s", urlPrefix, filename)

	return response.SuccessWithMessage(c, "Tải lên thành công", map[string]string{
		"url": url,
	})
}
