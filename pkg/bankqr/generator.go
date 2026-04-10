package bankqr

import (
	"fmt"

	"github.com/skip2/go-qrcode"
)

// GenerateEMVCoString tự build chuỗi BankQR chuẩn EMVCo TLV (Tag-Length-Value)
func GenerateEMVCoString(bin, accountNo string, amount int, message string) string {
	// Định dạng TLV (Tag + Length + Value)
	tlv := func(tag string, value string) string {
		return fmt.Sprintf("%s%02d%s", tag, len(value), value)
	}

	// 1. Beneficiary Info (GUID + BIN + Account)
	guid := tlv("00", "A000000727")
	
	// Napas Receiver (BIN + Account)
	receiverInfo := tlv("00", bin) + tlv("01", accountNo)
	napas := tlv("01", receiverInfo)
	
	// Service code QRIBFTTA (Chuyển khoản nhanh 24/7)
	serviceCode := tlv("02", "QRIBFTTA")
	
	merchantAccountInfo := guid + napas + serviceCode

	// 2. Cấu trúc toàn bộ Payload
	payload := ""
	payload += tlv("00", "01") // Payload Format Indicator
	if amount > 0 {
		payload += tlv("01", "12") // Dynamic QR (có số tiền)
	} else {
		payload += tlv("01", "11") // Static QR (không số tiền)
	}
	
	payload += tlv("38", merchantAccountInfo) // Merchant Account Information
	payload += tlv("53", "704")               // Currency Code (VND = 704)
	
	if amount > 0 {
		payload += tlv("54", fmt.Sprintf("%d", amount)) // Transaction Amount
	}
	
	payload += tlv("58", "VN") // Country Code
	
	if message != "" {
		additionalData := tlv("08", message) // Bill Number / Purpose
		payload += tlv("62", additionalData)
	}

	// 3. Chuẩn bị tính Checksum (Tag 63, length 04)
	payload += "6304"
	checksum := calculateCRC16CCITT(payload)
	
	return payload + checksum
}

// GenerateBase64QR tạo ra base64 hình ảnh QR từ chuỗi EMVCo
func GenerateBase64QR(bin, accountNo string, amount int, message string) (string, error) {
	bankQRString := GenerateEMVCoString(bin, accountNo, amount, message)
	
	// Tạo ảnh QR kích thước 256x256
	pngBytes, err := qrcode.Encode(bankQRString, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	
	// Có thể parse ra base64 nếu client cần ảnh base64 trực tiếp 
	// hoặc trả raw bytes tuỳ Framework UI
	// return base64.StdEncoding.EncodeToString(pngBytes), nil
	_ = pngBytes
	return bankQRString, nil
}

// Hàm tính CRC-16 (CCITT-FALSE) Polynomial 0x1021, Init 0xFFFF
func calculateCRC16CCITT(data string) string {
	crc := 0xFFFF
	for i := 0; i < len(data); i++ {
		crc ^= int(data[i]) << 8
		for j := 0; j < 8; j++ {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return fmt.Sprintf("%04X", crc&0xFFFF)
}
