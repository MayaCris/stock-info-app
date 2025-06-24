package logger

import (
	"context"
	"fmt"
	"time"
)

// IntegrityFileLogger implementa IntegrityLogger
type IntegrityFileLogger struct {
	Logger
	config *LogConfig
}

// NewIntegrityLogger crea un nuevo logger específico para integridad
func NewIntegrityLogger(baseLogger Logger, config *LogConfig) IntegrityLogger {
	return &IntegrityFileLogger{
		Logger: baseLogger.WithContext("INTEGRITY"),
		config: config,
	}
}

func (l *IntegrityFileLogger) LogValidationStart(ctx context.Context, validationType string, recordCount int) {
	l.Info(ctx, fmt.Sprintf("🔍 Starting %s validation", validationType),
		String("operation", "validation_start"),
		String("validation_type", validationType),
		Int("record_count", recordCount))
}

func (l *IntegrityFileLogger) LogValidationEnd(ctx context.Context, validationType string, issuesFound int, duration time.Duration) {
	level := "info"
	message := fmt.Sprintf("✅ %s validation completed", validationType)

	if issuesFound > 0 {
		if issuesFound < 5 {
			level = "warn"
			message = fmt.Sprintf("⚠️ %s validation completed with issues", validationType)
		} else {
			level = "error"
			message = fmt.Sprintf("❌ %s validation completed with critical issues", validationType)
		}
	}

	fields := []Field{
		String("operation", "validation_end"),
		String("validation_type", validationType),
		Int("issues_found", issuesFound),
		Duration("duration", duration),
	}

	switch level {
	case "warn":
		l.Warn(ctx, message, fields...)
	case "error":
		l.Error(ctx, message, nil, fields...)
	default:
		l.Info(ctx, message, fields...)
	}
}

func (l *IntegrityFileLogger) LogOrphanDetected(ctx context.Context, entityType string, entityID string, reason string, severity string) {
	fields := []Field{
		String("operation", "orphan_detected"),
		String("entity_type", entityType),
		String("entity_id", entityID),
		String("reason", reason),
		String("severity", severity),
	}

	if severity == "critical" {
		l.Error(ctx, fmt.Sprintf("💀 Orphaned %s detected: %s", entityType, entityID), nil, fields...)
	} else {
		l.Warn(ctx, fmt.Sprintf("⚠️ Orphaned %s detected: %s", entityType, entityID), fields...)
	}
}

func (l *IntegrityFileLogger) LogInconsistencyDetected(ctx context.Context, entityType string, field string, expected interface{}, actual interface{}) {
	l.Warn(ctx, fmt.Sprintf("⚡ Data inconsistency in %s.%s", entityType, field),
		String("operation", "inconsistency_detected"),
		String("entity_type", entityType),
		String("field", field),
		Any("expected", expected),
		Any("actual", actual))
}

func (l *IntegrityFileLogger) LogDuplicateDetected(ctx context.Context, entityType string, criteria string, count int) {
	l.Warn(ctx, fmt.Sprintf("🔄 Duplicate %s found", entityType),
		String("operation", "duplicate_detected"),
		String("entity_type", entityType),
		String("criteria", criteria),
		Int("duplicate_count", count))
}

func (l *IntegrityFileLogger) LogBusinessRuleViolation(ctx context.Context, rule string, entityType string, entityID string, details string) {
	l.Error(ctx, fmt.Sprintf("📋 Business rule violation: %s", rule), nil,
		String("operation", "business_rule_violation"),
		String("rule", rule),
		String("entity_type", entityType),
		String("entity_id", entityID),
		String("details", details))
}

func (l *IntegrityFileLogger) LogRepairAttempt(ctx context.Context, issueType string, entityID string, action string) {
	l.Info(ctx, fmt.Sprintf("🔧 Attempting to repair %s issue", issueType),
		String("operation", "repair_attempt"),
		String("issue_type", issueType),
		String("entity_id", entityID),
		String("action", action))
}

func (l *IntegrityFileLogger) LogRepairResult(ctx context.Context, issueType string, entityID string, success bool, err error) {
	fields := []Field{
		String("operation", "repair_result"),
		String("issue_type", issueType),
		String("entity_id", entityID),
		Bool("success", success),
	}

	if success {
		l.Info(ctx, fmt.Sprintf("✅ Successfully repaired %s issue", issueType), fields...)
	} else {
		l.Error(ctx, fmt.Sprintf("❌ Failed to repair %s issue", issueType), err, fields...)
	}
}

// CacheFileLogger implementa CacheLogger
type CacheFileLogger struct {
	Logger
	config *LogConfig
}

// NewCacheLogger crea un nuevo logger específico para cache
func NewCacheLogger(baseLogger Logger, config *LogConfig) CacheLogger {
	return &CacheFileLogger{
		Logger: baseLogger.WithContext("CACHE"),
		config: config,
	}
}

func (l *CacheFileLogger) LogCacheHit(ctx context.Context, key string, cacheType string) {
	l.Debug(ctx, fmt.Sprintf("🎯 Cache hit: %s", key),
		String("operation", "cache_hit"),
		String("key", key),
		String("cache_type", cacheType))
}

func (l *CacheFileLogger) LogCacheMiss(ctx context.Context, key string, cacheType string) {
	l.Debug(ctx, fmt.Sprintf("❌ Cache miss: %s", key),
		String("operation", "cache_miss"),
		String("key", key),
		String("cache_type", cacheType))
}

func (l *CacheFileLogger) LogCacheSet(ctx context.Context, key string, cacheType string, ttl time.Duration) {
	l.Debug(ctx, fmt.Sprintf("💾 Cache set: %s", key),
		String("operation", "cache_set"),
		String("key", key),
		String("cache_type", cacheType),
		Duration("ttl", ttl))
}

func (l *CacheFileLogger) LogCacheDelete(ctx context.Context, key string, cacheType string) {
	l.Debug(ctx, fmt.Sprintf("🗑️ Cache delete: %s", key),
		String("operation", "cache_delete"),
		String("key", key),
		String("cache_type", cacheType))
}

func (l *CacheFileLogger) LogCacheClear(ctx context.Context, cacheType string) {
	l.Info(ctx, fmt.Sprintf("🧹 Cache cleared: %s", cacheType),
		String("operation", "cache_clear"),
		String("cache_type", cacheType))
}

func (l *CacheFileLogger) LogCacheError(ctx context.Context, operation string, key string, err error) {
	l.Error(ctx, fmt.Sprintf("💥 Cache error during %s", operation), err,
		String("operation", "cache_error"),
		String("cache_operation", operation),
		String("key", key))
}

func (l *CacheFileLogger) LogCacheStats(ctx context.Context, cacheType string, hits int, misses int, hitRate float64) {
	l.Info(ctx, fmt.Sprintf("📊 Cache stats: %s", cacheType),
		String("operation", "cache_stats"),
		String("cache_type", cacheType),
		Int("hits", hits),
		Int("misses", misses),
		Float64("hit_rate", hitRate))
}
