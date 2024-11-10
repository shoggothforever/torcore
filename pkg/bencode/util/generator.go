package util

import (
	"crypto/sha1"
	"time"
)

const HashLength = 20

// GeneratePeerID 根据用户信息生成固定的 20 字节 peerID
func GeneratePeerID(userInfo string) [HashLength]byte {
	// 创建 SHA-1 哈希
	hash := sha1.New()
	hash.Write([]byte(userInfo + time.Now().String()))

	// 计算哈希并将其填充到 20 字节数组
	var peerID [HashLength]byte
	copy(peerID[:], hash.Sum(nil)[:HashLength])

	return peerID
}
