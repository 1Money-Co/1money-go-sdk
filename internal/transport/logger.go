/*
 * Copyright 2025 1Money Co.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package transport

import (
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// loggerValue stores the logger instance using atomic.Value for thread-safe access
	loggerValue atomic.Value
	loggerOnce  sync.Once
)

// initLogger initializes the package-level logger.
// It uses sync.Once to ensure the logger is only created once.
// Logger is only enabled when ONEMONEY_DEBUG or ONEMONEY_LOG_LEVEL environment variable is set.
//
// Environment variables:
//   - ONEMONEY_DEBUG: Enable debug level logging
//   - ONEMONEY_LOG_LEVEL: Set log level (debug, info, warn, error)
//   - ONEMONEY_ENABLE_STACKTRACE: Enable stack trace output (default: disabled)
func initLogger() {
	loggerOnce.Do(func() {
		// Check if logging is enabled via environment variables
		debugEnv := os.Getenv("ONEMONEY_DEBUG")
		logLevelEnv := os.Getenv("ONEMONEY_LOG_LEVEL")
		stacktraceEnv := os.Getenv("ONEMONEY_ENABLE_STACKTRACE")

		var l *zap.Logger

		// Default to no-op logger (no logging)
		if debugEnv == "" && logLevelEnv == "" {
			l = zap.NewNop()
			loggerValue.Store(l)
			return
		}

		// Determine log level
		level := zapcore.InfoLevel
		if logLevelEnv != "" {
			switch strings.ToLower(logLevelEnv) {
			case "debug":
				level = zapcore.DebugLevel
			case "info":
				level = zapcore.InfoLevel
			case "warn", "warning":
				level = zapcore.WarnLevel
			case "error":
				level = zapcore.ErrorLevel
			default:
				level = zapcore.InfoLevel
			}
		} else if debugEnv != "" {
			// If ONEMONEY_DEBUG is set, enable debug level
			level = zapcore.DebugLevel
		}

		// Create development config for human-readable output
		config := zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(level)
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		// Disable stack trace by default
		// Only enable if ONEMONEY_ENABLE_STACKTRACE is explicitly set
		if stacktraceEnv == "" || strings.ToLower(stacktraceEnv) == "false" || stacktraceEnv == "0" {
			config.DisableStacktrace = true
		} else {
			config.DisableStacktrace = false
		}

		var err error
		l, err = config.Build()
		if err != nil {
			// Fallback to a no-op logger if initialization fails
			l = zap.NewNop()
		}

		loggerValue.Store(l)
	})
}

// getLogger returns the package-level logger instance.
// It initializes the logger on first call using sync.Once.
// This function is thread-safe and can be called concurrently.
func getLogger() *zap.Logger {
	initLogger()
	l := loggerValue.Load()
	if l == nil {
		// This should never happen, but return a no-op logger as a safety fallback
		return zap.NewNop()
	}
	return l.(*zap.Logger)
}

// SetLogger allows users to configure a custom logger for the transport package.
// This is useful for integrating with application-wide logging configuration.
//
// This function is thread-safe and can be called concurrently with getLogger()
// and other logging operations. The logger update is atomic and will be visible
// to all goroutines immediately.
//
// Note: It's recommended to call SetLogger during application initialization,
// before any transport operations begin, to ensure consistent logging behavior.
func SetLogger(l *zap.Logger) {
	if l != nil {
		loggerValue.Store(l)
	}
}
