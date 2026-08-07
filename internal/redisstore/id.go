package redisstore

import (
	"fmt"
	"strconv"
	"strings"
)

// NormalizeID turns "1", "0001", or "customer:1" into the canonical "0001".
func NormalizeID(id string) string {
	id = strings.TrimSpace(id)
	id = strings.TrimPrefix(id, "customer:")
	if n, err := strconv.Atoi(id); err == nil && n >= 0 {
		return fmt.Sprintf("%04d", n)
	}
	return id
}
