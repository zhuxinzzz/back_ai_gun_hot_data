package utils

import (
	"fmt"
	"runtime"
	"strings"
)

func PrintStartInfo(coinCount int) {
	fmt.Printf("Found %d coins to process\n", coinCount)
}

func stack() []byte {
	buf := make([]byte, 8192) // 2^13 8KB
	n := runtime.Stack(buf, false)
	return buf[:n]
}

func GetStack() string {
	return fmt.Sprintf("%s", stack())
}

func GetFuncName() string {
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		return runtime.FuncForPC(pc).Name()
	} else {
		return ""
	}
}

// NormalizeName 标准化名称（小写去空格）
func NormalizeName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", ""))
}
