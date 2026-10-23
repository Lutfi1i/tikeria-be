package utils

// MaskNIK masks the NIK showing only the first 4 and last 4 characters.
// Example: 3174000000001234 -> 3174********1234
func MaskNIK(nik string) string {
	if len(nik) < 8 {
		return nik
	}
	prefix := nik[:4]
	suffix := nik[len(nik)-4:]
	maskedLen := len(nik) - 8
	var stars string
	for i := 0; i < maskedLen; i++ {
		stars += "*"
	}
	return prefix + stars + suffix
}
