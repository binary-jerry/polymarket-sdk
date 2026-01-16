package common

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

// TimestampMs 返回当前时间戳（毫秒）
func TimestampMs() int64 {
	return time.Now().UnixMilli()
}

// TimestampSec 返回当前时间戳（秒）
func TimestampSec() int64 {
	return time.Now().Unix()
}

// TimestampMsStr 返回当前时间戳字符串（毫秒）
func TimestampMsStr() string {
	return strconv.FormatInt(TimestampMs(), 10)
}

// TimestampSecStr 返回当前时间戳字符串（秒）
func TimestampSecStr() string {
	return strconv.FormatInt(TimestampSec(), 10)
}

// GenerateRandomHex 生成随机十六进制字符串
func GenerateRandomHex(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateSalt 生成订单盐值
// 使用与 Python SDK 一致的方式：timestamp * random
func GenerateSalt() (*big.Int, error) {
	now := time.Now().Unix()
	// 生成 0-1 之间的随机数
	randBytes := make([]byte, 8)
	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, err
	}
	// 将随机字节转换为 0-1 之间的浮点数
	randVal := float64(binary.BigEndian.Uint64(randBytes)) / float64(^uint64(0))
	// 计算 salt = now * random
	salt := int64(float64(now) * randVal)
	return big.NewInt(salt), nil
}

// GenerateNonce 生成订单 nonce
func GenerateNonce() (*big.Int, error) {
	return GenerateSalt()
}

// SaltToString 将盐值转换为字符串
func SaltToString(salt *big.Int) string {
	if salt == nil {
		return "0"
	}
	return salt.String()
}

// StringToBigInt 将字符串转换为大整数
func StringToBigInt(s string) (*big.Int, bool) {
	n := new(big.Int)
	return n.SetString(s, 10)
}

// HexToBigInt 将十六进制字符串转换为大整数
func HexToBigInt(s string) (*big.Int, bool) {
	// 移除 0x 前缀
	if len(s) >= 2 && s[:2] == "0x" {
		s = s[2:]
	}
	n := new(big.Int)
	return n.SetString(s, 16)
}

// BigIntToHex 将大整数转换为十六进制字符串（带 0x 前缀）
func BigIntToHex(n *big.Int) string {
	if n == nil {
		return "0x0"
	}
	return fmt.Sprintf("0x%x", n)
}

// NormalizeAddress 规范化以太坊地址（转为小写并添加 0x 前缀）
func NormalizeAddress(addr string) string {
	if len(addr) == 0 {
		return ""
	}
	// 移除 0x 前缀后转小写再添加回来
	if len(addr) >= 2 && addr[:2] == "0x" {
		addr = addr[2:]
	}
	return "0x" + addr
}

// IsValidAddress 检查地址格式是否有效
func IsValidAddress(addr string) bool {
	if len(addr) == 0 {
		return false
	}
	// 移除 0x 前缀
	if len(addr) >= 2 && addr[:2] == "0x" {
		addr = addr[2:]
	}
	// 检查长度（40 个十六进制字符 = 20 字节）
	if len(addr) != 40 {
		return false
	}
	// 检查是否全为十六进制字符
	for _, c := range addr {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// MinInt 返回两个整数中的较小值
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt 返回两个整数中的较大值
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ZeroAddress 零地址常量
const ZeroAddress = "0x0000000000000000000000000000000000000000"

// IsZeroAddress 检查是否为零地址
func IsZeroAddress(addr string) bool {
	normalized := NormalizeAddress(addr)
	return normalized == ZeroAddress
}
