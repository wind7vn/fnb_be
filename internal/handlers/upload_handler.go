package handlers

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/response"
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

	filename := fmt.Sprintf("img_%d%s", time.Now().UnixNano(), ext)
	savePath := fmt.Sprintf("./uploads/%s", filename)

	if err := c.SaveFile(file, savePath); err != nil {
		return response.Error(c, errors.NewInternalServer(err))
	}

	url := fmt.Sprintf("http://localhost:8080/uploads/%s", filename)

	return response.SuccessWithMessage(c, "Tải lên thành công", map[string]string{
		"url": url,
	})
}
