package utils

import (
	"strings"
)

// 使用一个全局固定的、已经打乱过的 62 字符字典。
// 编码和解码必须使用同一个字典。
const customBase62Charset = "vPh7zG2c1TklA4yD0xNq8rE3f5jW9uI6BmVnSoYpObCtwRFLHQKdZJgXaMesiU"

// EncodeBase62 将数字 ID 转换为 Base62 字符串
func EncodeBase62(id uint64) string {
	if id == 0 {
		return string(customBase62Charset[0])
	}

	var chars []byte
	base := uint64(62)

	// 辗转相除法
	for id > 0 {
		remainder := id % base
		chars = append(chars, customBase62Charset[remainder])
		id = id / base
	}

	// 因为收集的余数是反向的，需要反转切片
	reverseBytes(chars)

	return string(chars)
}

// DecodeBase62 将 Base62 字符串还原为数字 ID
func DecodeBase62(shortKey string) uint64 {
	var id uint64
	base := uint64(62)

	// 从左到右遍历字符，使用纯整数乘法累加，避免 float64 精度丢失问题
	for _, char := range shortKey {
		index := strings.IndexRune(customBase62Charset, char)
		if index == -1 {
			// 如果输入了非法字符，根据你的业务逻辑处理（此处简化为返回 0，建议实际业务中返回 error）
			return 0
		}
		id = id*base + uint64(index)
	}

	return id
}

// reverseBytes 辅助函数：反转字节切片
func reverseBytes(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}
