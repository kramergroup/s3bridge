package util

import (
	"log"
	"os"
	"strconv"
	"time"
)

// LookupEnvOrString obtains the value of the environment variable `key` or
// returns `defaultVal` if no environment variable is set
func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

// LookupEnvOrInt obtains the value of the environment variable `key` or
// returns `defaultVal` if no environment variable is set
func LookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

// LookupEnvOrDuration obtains the value of the environment variable `key` or
// returns `defaultVal` if no environment variable is set
func LookupEnvOrDuration(key string, defaultVal time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		v, err := time.ParseDuration(val)
		if err != nil {
			log.Fatalf("LookupEnvOrDuration[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}
