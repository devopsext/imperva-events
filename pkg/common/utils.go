package common

import (
	"os"
	"strconv"
)

//func GetEnv(key, defaultValue string) string {
//	if value := os.Getenv(key); value != "" {
//		return value
//	}
//	return defaultValue
//}

func EnvGet(key string, defaultValue interface{}) interface{} {

	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}

	switch defaultValue.(type) {
	case int:
		r, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return r
	case bool:
		r, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return r
	default:
		return value
	}
}
