package redis

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func HashFilters(filters interface{}) (string, error) {
	data, err := json.Marshal(filters)
	if err != nil {
		return "", fmt.Errorf("filterHasher: marshal error: %w", err)
	}

	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:]), nil
}
