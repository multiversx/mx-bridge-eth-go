package core

import (
    logger "github.com/ElrondNetwork/elrond-go-logger"
)

// loggerWithIdentifier is a decorator for the logger
type loggerWithIdentifier struct {
    logger     logger.Logger
    identifier string
}

// NewLoggerWithIdentifier creates a new loggerWithIdentifier instance
func NewLoggerWithIdentifier(logger logger.Logger, identifier string) *loggerWithIdentifier {
    if logger == nil {
        return nil
    }

    log := &loggerWithIdentifier{
        logger:     logger,
        identifier: identifier,
    }

    return log
}

// Trace outputs a tracing log message with optional provided arguments, preceded by the identifier
func (l *loggerWithIdentifier) Trace(message string, args ...interface{}) {
    l.logger.Trace(l.formatMessage(message), args...)
}

// Debug outputs a debugging log message with optional provided arguments, preceded by the identifier
func (l *loggerWithIdentifier) Debug(message string, args ...interface{}) {
    l.logger.Debug(l.formatMessage(message), args...)
}

// Info outputs an information log message with optional provided arguments, preceded by the identifier
func (l *loggerWithIdentifier) Info(message string, args ...interface{}) {
    l.logger.Info(l.formatMessage(message), args...)
}

// Warn outputs a warning log message with optional provided arguments, preceded by the identifier
func (l *loggerWithIdentifier) Warn(message string, args ...interface{}) {
    l.logger.Warn(l.formatMessage(message), args...)
}

// Error outputs an error log message with optional provided arguments, preceded by the identifier
func (l *loggerWithIdentifier) Error(message string, args ...interface{}) {
    l.logger.Error(l.formatMessage(message), args...)
}

// LogIfError outputs an error log message preceded by the identifier with optional provided arguments if the provided error parameter is not nil
func (l *loggerWithIdentifier) LogIfError(err error, args ...interface{}) {
    if err == nil {
        return
    }

    l.Error(err.Error(), args...)
}

// Log forwards the log line towards underlying log output handler, in respect with the identifier
func (l *loggerWithIdentifier) Log(line *logger.LogLine) {
    if line == nil {
        return
    }

    line.Message = l.formatMessage(line.Message)
    l.logger.Log(line)
}

// SetLevel sets the current level of the logger
func (l *loggerWithIdentifier) SetLevel(logLevel logger.LogLevel) {
    l.logger.SetLevel(logLevel)
}

// GetLevel gets the current level of the logger
func (l *loggerWithIdentifier) GetLevel() logger.LogLevel {
    return l.logger.GetLevel()
}

// IsInterfaceNil returns true if there is no value under the interface
func (l *loggerWithIdentifier) IsInterfaceNil() bool {
    return l == nil
}

func (l *loggerWithIdentifier) formatMessage(message string) string {
    return l.identifier + " " + message
}
