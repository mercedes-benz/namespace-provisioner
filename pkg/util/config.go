// SPDX-License-Identifier: MIT

package util

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
)

func getEnv(log logr.Logger, key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		log.V(1).Info(fmt.Sprintf("Using env %s=%s", key, value))
		return value
	}
	log.V(1).Info(fmt.Sprintf("Using default value %s=%s", key, fallback))
	return fallback
}

func GetConfigNamespace(log logr.Logger) string {
	return getEnv(log, "CONFIG_NAMESPACE", "default")
}

func GetLogLevel(log logr.Logger) string {
	return getEnv(log, "LOG_LEVEL", "INFO")
}
