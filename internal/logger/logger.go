package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

/* ------------------------------------------------------------------ *
|  1. Configuration & functional‑options                              |
* -------------------------------------------------------------------*/

type Config struct {
	Level      string
	FilePath   string
	Format     string
	Version    string
	Component  string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

type Option func(*Config)

func WithLevel(lvl string) Option      { return func(c *Config) { c.Level = lvl } }
func WithFormat(fmt string) Option     { return func(c *Config) { c.Format = fmt } }
func WithFile(path string) Option      { return func(c *Config) { c.FilePath = path } }
func WithVersion(v string) Option      { return func(c *Config) { c.Version = v } }
func WithComponent(comp string) Option { return func(c *Config) { c.Component = comp } }
func WithRotation(size, backups, age int) Option {
	return func(c *Config) {
		c.MaxSize, c.MaxBackups, c.MaxAge = size, backups, age
	}
}

/* ------------------------------------------------------------------ *
|  2. Package‑level state                                             |
* -------------------------------------------------------------------*/

var (
	core        zapcore.Core
	atomicLevel zap.AtomicLevel
	root        *zap.Logger

	active bool
	mu     sync.RWMutex
)

/* ------------------------------------------------------------------ *
|  3. Init / Shutdown                                                 |
* -------------------------------------------------------------------*/

// Init builds the global zap core. Calling Init twice replaces the old core.
func Init(opts ...Option) error {
	cfg := defaultConfig()
	for _, apply := range opts {
		apply(cfg)
	}

	enc, err := buildEncoder(cfg.Format)
	if err != nil {
		return err
	}
	ws, isFile, err := buildWriter(cfg)
	if err != nil {
		return err
	}
	lvl, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	atomicLevel = lvl

	newCore := zapcore.NewCore(enc, ws, atomicLevel)

	mu.Lock()
	defer mu.Unlock()

	// Flush previous file writer (if any)
	if active && root != nil && isFile {
		_ = root.Sync()
	}

	core = newCore
	root = zap.New(core,
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("version", cfg.Version),
			zap.String("component", cfg.Component),
		),
	)
	active = true
	return nil
}

// Shutdown flushes logs when the writer is a file.
func Shutdown() error {
	mu.RLock()
	defer mu.RUnlock()

	if !active || root == nil {
		return fmt.Errorf("logger not initialized")
	}
	if err := root.Sync(); err != nil && !isPathErr(err) {
		return err
	}
	active = false
	return nil
}

/* ------------------------------------------------------------------ *
|  4. Helpers                                                         |
* -------------------------------------------------------------------*/

func defaultConfig() *Config {
	return &Config{
		Level:      "info",
		Format:     "console",
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
	}
}

func buildEncoder(format string) (zapcore.Encoder, error) {
	switch format {
	case "json":
		return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), nil
	case "console":
		cfg := zap.NewDevelopmentEncoderConfig()
		cfg.EncodeTime = zapcore.ISO8601TimeEncoder
		return zapcore.NewConsoleEncoder(cfg), nil
	default:
		return nil, fmt.Errorf("unknown log format %q", format)
	}
}

func buildWriter(cfg *Config) (zapcore.WriteSyncer, bool, error) {
	if cfg.FilePath == "" {
		return zapcore.AddSync(os.Stdout), false, nil
	}
	if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0o755); err != nil {
		return nil, false, fmt.Errorf("create log dir: %w", err)
	}
	ws := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   true,
	})
	return ws, true, nil
}

func isPathErr(err error) bool {
	_, ok := err.(*os.PathError)
	return ok
}

/* ------------------------------------------------------------------ *
|  5. Context helpers & child loggers                                 |
* -------------------------------------------------------------------*/

type ctxKey struct{ name string }

type loggerKey struct{}

var (
	requestIDKey = &ctxKey{"reqID"}
	traceIDKey   = &ctxKey{"traceID"}
)

// WithLogger attaches a *zap.Logger to a context.
func WithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// FromContext returns a logger with request / trace IDs if present.
func FromContext(ctx context.Context) *zap.Logger {
	if !active {
		return zap.NewNop()
	}
	if l, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
		return l
	}

	fields := make([]zap.Field, 0, 2)
	if v := ctx.Value(requestIDKey); v != nil {
		fields = append(fields, zap.String("request_id", v.(string)))
	}
	if v := ctx.Value(traceIDKey); v != nil {
		fields = append(fields, zap.String("trace_id", v.(string)))
	}
	if len(fields) == 0 {
		return root
	}
	return root.With(fields...)
}

// New returns a component‑scoped child logger.
func New(component string) *zap.Logger {
	if !active {
		return zap.NewNop()
	}
	return root.With(zap.String("component", component))
}

/* ------------------------------------------------------------------ *
|  6. Convenience wrappers                                            |
* -------------------------------------------------------------------*/

func Debug(msg string, fields ...zap.Field) {
	if active {
		root.Debug(msg, fields...)
	}
}
func Info(msg string, fields ...zap.Field) {
	if active {
		root.Info(msg, fields...)
	}
}
func Warn(msg string, fields ...zap.Field) {
	if active {
		root.Warn(msg, fields...)
	}
}
func Error(msg string, fields ...zap.Field) {
	if active {
		root.Error(msg, fields...)
	}
}

/* ------------------------------------------------------------------ *
|  7. Hot‑swap log‑level                                              |
* -------------------------------------------------------------------*/

func UpdateLevel(lvl string) error {
	if !active {
		return fmt.Errorf("logger not initialized")
	}
	level, err := zap.ParseAtomicLevel(lvl)
	if err != nil {
		return err
	}
	atomicLevel.SetLevel(level.Level())
	return nil
}

/* ------------------------------------------------------------------ *
|  8. Error helpers                                                   |
* -------------------------------------------------------------------*/

// NewError creates a new error with the given message
func NewError(msg string) error {
	return fmt.Errorf("%s", msg)
}

// WrapError wraps an error with additional context
func WrapError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}
