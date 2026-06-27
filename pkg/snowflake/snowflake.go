package snowflake

import (
	"strconv"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	node    *snowflake.Node
	once    sync.Once
	initErr error
)

// Init 初始化雪花算法节点（workerID=1，单机部署）
func Init() error {
	once.Do(func() {
		node, initErr = snowflake.NewNode(1)
	})
	return initErr
}

// GenerateID 生成标准雪花ID字符串（≈19位数字）
func GenerateID() string {
	if node == nil {
		_ = Init()
	}
	return strconv.FormatInt(node.Generate().Int64(), 10)
}

// GenerateShortID 生成指定位数范围的短ID字符串（用于 User/AdminUser 等特殊表）
// 从雪花ID中取后 digits 位，不足时补齐以保证位数
func GenerateShortID(digits int) string {
	sid := node.Generate().Int64()
	var mod int64 = 1
	for i := 0; i < digits; i++ {
		mod *= 10
	}
	short := sid % mod
	minVal := int64(1)
	for i := 1; i < digits; i++ {
		minVal *= 10
	}
	if short < minVal {
		short += minVal
	}
	return strconv.FormatInt(short, 10)
}
