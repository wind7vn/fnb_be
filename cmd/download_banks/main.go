package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	// 1. Phân giải đúng đường dẫn JSON (nếu chạy từ cmd/download_banks thì cần lùi lại)
	jsonPath := "../../data/momo_banks.json"
	fileBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Println("Lỗi đọc JSON, hãy đảm bảo bạn chạy lệnh này bên trong thư mục cmd/download_banks:", err)
		return
	}

	var banks map[string]map[string]interface{}
	if err := json.Unmarshal(fileBytes, &banks); err != nil {
		fmt.Println("Error unmarshaling:", err)
		return
	}

	// 2. Tạo thư mục public để host Local Icons
	targetDir := "../../web/bank_icons"
	os.MkdirAll(targetDir, os.ModePerm)

	// 3. Quét tất cả icon ngoài
	successCount := 0
	for id, bankInfo := range banks {
		logoUrlStr, ok := bankInfo["bankLogoUrl"].(string)
		if !ok || logoUrlStr == "" {
			continue
		}

		localFilename := fmt.Sprintf("%s.png", id)
		localPath := fmt.Sprintf("%s/%s", targetDir, localFilename)

		// Rút về máy, chặn request lên CDN Momo
		if strings.HasPrefix(logoUrlStr, "http") {
			fmt.Printf("Đang tải logo %s...\n", id)
			resp, err := http.Get(logoUrlStr)
			if err == nil && resp.StatusCode == 200 {
				out, errCreate := os.Create(localPath)
				if errCreate == nil {
					io.Copy(out, resp.Body)
					out.Close()
					successCount++
					
					// Thay url bằng Local Static Path
					bankInfo["bankLogoUrl"] = fmt.Sprintf("/bank_icons/%s", localFilename)
				}
				resp.Body.Close()
			} else {
				fmt.Printf("❌ Tải thất bại ngân hàng %s\n", id)
			}
		}
	}

	// 4. Cập nhật lại chính tệp JSON
	updatedBytes, _ := json.MarshalIndent(banks, "", "  ")
	os.WriteFile(jsonPath, updatedBytes, 0644)
	fmt.Printf("Xong! Đã tải %d icon cục bộ và ngưng sử dụng đường dẫn bên ngoài.\n", successCount)
}
