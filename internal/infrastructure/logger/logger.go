// Package logger provides a configured structured logger.
package logger

import "go.uber.org/zap"

// New returns a production-grade structured logger.
func New() (*zap.Logger, error) {
	return zap.NewProduction()
}
