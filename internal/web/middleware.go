package web

import (
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Shugur-Network/relay/internal/logger"
	"go.uber.org/zap"
)

// SecurityHeaders defines the security headers to be applied to responses
type SecurityHeaders struct {
	// Content Security Policy - prevents XSS and other injection attacks
	CSP string
	// HTTP Strict Transport Security - enforces HTTPS
	HSTS string
	// X-Frame-Options - prevents clickjacking
	XFrameOptions string
	// X-Content-Type-Options - prevents MIME sniffing
	XContentTypeOptions string
	// Referrer-Policy - controls referrer information
	ReferrerPolicy string
	// Permissions-Policy - controls browser features
	PermissionsPolicy string
	// X-XSS-Protection - XSS protection (legacy but still useful)
	XXSSProtection string
}

// DefaultSecurityHeaders returns a set of secure default headers
// Note: Basic security headers (HSTS, X-Frame-Options, etc.) are handled by Caddy proxy
// This focuses on application-specific headers like CSP
func DefaultSecurityHeaders() *SecurityHeaders {
	return &SecurityHeaders{
		// Content Security Policy - restrictive but functional for a relay dashboard
		// This is application-specific and better handled at app level than proxy level
		CSP: "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " + // Allow inline scripts for dashboard functionality
			"style-src 'self' 'unsafe-inline'; " + // Allow inline styles for dashboard
			"img-src 'self' data: https:; " + // Allow images from self, data URIs, and HTTPS
			"connect-src 'self' wss: ws:; " + // Allow WebSocket connections for relay functionality
			"font-src 'self'; " +
			"object-src 'none'; " + // Disable plugins
			"base-uri 'self'; " +
			"frame-ancestors 'none'; " + // Prevent framing
			"upgrade-insecure-requests", // Upgrade HTTP to HTTPS

		// Leave other headers empty - they're handled by Caddy proxy
		// This prevents header duplication and conflicts
		HSTS:                "",
		XFrameOptions:       "",
		XContentTypeOptions: "",
		ReferrerPolicy:      "",
		PermissionsPolicy:   "",
		XXSSProtection:      "",
	}
}

// APISecurityHeaders returns security headers optimized for API endpoints
// Note: Basic security headers are handled by Caddy proxy
// This focuses on API-specific CSP and other application-level headers
func APISecurityHeaders() *SecurityHeaders {
	return &SecurityHeaders{
		// More restrictive CSP for API endpoints - no scripts or styles needed
		CSP: "default-src 'none'; " +
			"frame-ancestors 'none'; " +
			"upgrade-insecure-requests",

		// Leave other headers empty - handled by Caddy proxy
		HSTS:                "",
		XFrameOptions:       "",
		XContentTypeOptions: "",
		ReferrerPolicy:      "",
		PermissionsPolicy:   "",
		XXSSProtection:      "",
	}
}

// SecurityMiddleware wraps an http.Handler with security headers
func SecurityMiddleware(headers *SecurityHeaders) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Apply security headers before calling the next handler
			applySecurityHeaders(w, headers)
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHandlerFunc wraps an http.HandlerFunc with security headers
func SecurityHandlerFunc(headers *SecurityHeaders, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Apply security headers before calling the handler
		applySecurityHeaders(w, headers)
		handlerFunc(w, r)
	})
}

// Apply applies the security headers directly to a ResponseWriter
func (sh *SecurityHeaders) Apply(w http.ResponseWriter) {
	applySecurityHeaders(w, sh)
}

// applySecurityHeaders applies the security headers to the response
func applySecurityHeaders(w http.ResponseWriter, headers *SecurityHeaders) {
	if headers.CSP != "" {
		w.Header().Set("Content-Security-Policy", headers.CSP)
	}
	
	if headers.HSTS != "" {
		w.Header().Set("Strict-Transport-Security", headers.HSTS)
	}
	
	if headers.XFrameOptions != "" {
		w.Header().Set("X-Frame-Options", headers.XFrameOptions)
	}
	
	if headers.XContentTypeOptions != "" {
		w.Header().Set("X-Content-Type-Options", headers.XContentTypeOptions)
	}
	
	if headers.ReferrerPolicy != "" {
		w.Header().Set("Referrer-Policy", headers.ReferrerPolicy)
	}
	
	if headers.PermissionsPolicy != "" {
		w.Header().Set("Permissions-Policy", headers.PermissionsPolicy)
	}
	
	if headers.XXSSProtection != "" {
		w.Header().Set("X-XSS-Protection", headers.XXSSProtection)
	}
}

// SecureHandlerFunc is a convenience function that wraps a handler function with default security headers
func SecureHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return SecurityHandlerFunc(DefaultSecurityHeaders(), handlerFunc)
}

// SecureAPIHandlerFunc is a convenience function that wraps an API handler function with API-optimized security headers
func SecureAPIHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return SecurityHandlerFunc(APISecurityHeaders(), handlerFunc)
}

// InputValidation provides comprehensive input validation and sanitization
type InputValidation struct {
	// MaxPathLength limits URL path length to prevent buffer overflow attacks
	MaxPathLength int
	// MaxQueryLength limits query string length
	MaxQueryLength int
	// MaxHeaderLength limits individual header length
	MaxHeaderLength int
	// AllowedQueryParams whitelist of allowed query parameter names
	AllowedQueryParams map[string]bool
	// PathPatterns allowed path patterns (regex)
	PathPatterns []*regexp.Regexp
}

// DefaultInputValidation returns secure default input validation settings
func DefaultInputValidation() *InputValidation {
	// Compile safe path patterns for our endpoints
	pathPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^/$`),                                // Root path
		regexp.MustCompile(`^/api/info$`),                        // API info endpoint
		regexp.MustCompile(`^/api/stats$`),                       // API stats endpoint
		regexp.MustCompile(`^/api/metrics$`),                     // API metrics endpoint
		regexp.MustCompile(`^/api/cluster$`),                     // API cluster endpoint
		regexp.MustCompile(`^/static/[a-zA-Z0-9._-]+\.[a-zA-Z0-9]+$`), // Static files with safe chars
	}

	allowedQueryParams := map[string]bool{
		"type":   true, // For cluster API type parameter
		"format": true, // For potential future formatting options
	}

	return &InputValidation{
		MaxPathLength:      2048, // Reasonable path length limit
		MaxQueryLength:     4096, // Query string length limit
		MaxHeaderLength:    8192, // Individual header length limit
		AllowedQueryParams: allowedQueryParams,
		PathPatterns:       pathPatterns,
	}
}

// APIInputValidation returns input validation settings optimized for API endpoints
func APIInputValidation() *InputValidation {
	// More restrictive for API endpoints
	pathPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^/api/info$`),
		regexp.MustCompile(`^/api/stats$`),
		regexp.MustCompile(`^/api/metrics$`),
		regexp.MustCompile(`^/api/cluster$`),
	}

	allowedQueryParams := map[string]bool{
		"type": true, // Only type parameter allowed for APIs
	}

	return &InputValidation{
		MaxPathLength:      1024, // Shorter path limit for APIs
		MaxQueryLength:     1024, // Shorter query limit for APIs
		MaxHeaderLength:    4096, // Shorter header limit for APIs
		AllowedQueryParams: allowedQueryParams,
		PathPatterns:       pathPatterns,
	}
}

// ValidateRequest validates an HTTP request against the input validation rules
func (iv *InputValidation) ValidateRequest(r *http.Request) error {
	// Validate path length
	if len(r.URL.Path) > iv.MaxPathLength {
		return &ValidationError{
			Type:    "path_length",
			Message: "Request path too long",
			Field:   "url_path",
			Value:   r.URL.Path,
		}
	}

	// Validate query string length
	if len(r.URL.RawQuery) > iv.MaxQueryLength {
		return &ValidationError{
			Type:    "query_length",
			Message: "Query string too long",
			Field:   "query_string",
			Value:   r.URL.RawQuery,
		}
	}

	// Validate path pattern
	pathValid := false
	for _, pattern := range iv.PathPatterns {
		if pattern.MatchString(r.URL.Path) {
			pathValid = true
			break
		}
	}
	if !pathValid {
		return &ValidationError{
			Type:    "invalid_path",
			Message: "Invalid request path",
			Field:   "url_path",
			Value:   r.URL.Path,
		}
	}

	// Validate query parameters
	if len(iv.AllowedQueryParams) > 0 {
		queryValues := r.URL.Query()
		for param := range queryValues {
			if !iv.AllowedQueryParams[param] {
				return &ValidationError{
					Type:    "invalid_query_param",
					Message: "Invalid query parameter",
					Field:   param,
					Value:   queryValues.Get(param),
				}
			}
		}
	}

	// Validate header lengths
	for name, values := range r.Header {
		for _, value := range values {
			if len(value) > iv.MaxHeaderLength {
				return &ValidationError{
					Type:    "header_length",
					Message: "Header value too long",
					Field:   name,
					Value:   value,
				}
			}
		}
	}

	// Check for potential injection patterns in critical headers
	dangerousHeaders := []string{"Host", "X-Forwarded-For", "User-Agent", "Referer"}
	for _, headerName := range dangerousHeaders {
		if headerValue := r.Header.Get(headerName); headerValue != "" {
			if err := validateHeaderValue(headerName, headerValue); err != nil {
				return err
			}
		}
	}

	return nil
}

// ValidationError represents an input validation error
type ValidationError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Field   string `json:"field"`
	Value   string `json:"value,omitempty"` // Omit sensitive values in logs
}

func (e *ValidationError) Error() string {
	return e.Message
}

// validateHeaderValue checks header values for injection patterns
func validateHeaderValue(name, value string) error {
	// Check for null bytes and control characters
	if !utf8.ValidString(value) {
		return &ValidationError{
			Type:    "invalid_encoding",
			Message: "Invalid character encoding in header",
			Field:   name,
		}
	}

	// Check for null bytes and dangerous control characters
	if strings.ContainsAny(value, "\x00\r\n") {
		return &ValidationError{
			Type:    "header_injection",
			Message: "Potential header injection detected",
			Field:   name,
		}
	}

	// Additional validation for specific headers
	switch name {
	case "Host":
		// Basic hostname validation - should not contain spaces or special chars
		if strings.ContainsAny(value, " \t<>\"'") {
			return &ValidationError{
				Type:    "invalid_host",
				Message: "Invalid characters in Host header",
				Field:   name,
			}
		}
	case "User-Agent":
		// Reasonable length limit for User-Agent
		if len(value) > 1024 {
			return &ValidationError{
				Type:    "user_agent_length",
				Message: "User-Agent header too long",
				Field:   name,
			}
		}
	}

	return nil
}

// SanitizeQueryParam sanitizes a query parameter value
func SanitizeQueryParam(param string) string {
	// URL decode the parameter
	decoded, err := url.QueryUnescape(param)
	if err != nil {
		// If decoding fails, return empty string (reject the input)
		return ""
	}

	// Remove null bytes and control characters except space and tab
	sanitized := strings.Map(func(r rune) rune {
		// Allow printable ASCII characters and space/tab
		if r >= 32 && r <= 126 || r == '\t' {
			return r
		}
		return -1 // Remove the character
	}, decoded)

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	// Limit length
	if len(sanitized) > 256 {
		sanitized = sanitized[:256]
	}

	return sanitized
}

// SanitizePath sanitizes and validates a file path for static file serving
func SanitizePath(path string) (string, error) {
	// URL decode
	decoded, err := url.QueryUnescape(path)
	if err != nil {
		return "", &ValidationError{
			Type:    "path_decode_error",
			Message: "Failed to decode path",
			Field:   "path",
			Value:   path,
		}
	}

	// Clean the path
	cleaned := filepath.Clean(decoded)

	// Check for path traversal attempts
	if strings.Contains(cleaned, "..") || strings.HasPrefix(cleaned, "/") {
		return "", &ValidationError{
			Type:    "path_traversal",
			Message: "Path traversal attempt detected",
			Field:   "path",
			Value:   path,
		}
	}

	// Only allow safe characters in filenames
	safePattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !safePattern.MatchString(filepath.Base(cleaned)) {
		return "", &ValidationError{
			Type:    "unsafe_filename",
			Message: "Filename contains unsafe characters",
			Field:   "path",
			Value:   path,
		}
	}

	return cleaned, nil
}

// ValidationMiddleware wraps an http.Handler with input validation
func ValidationMiddleware(validation *InputValidation) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := validation.ValidateRequest(r); err != nil {
				if validationErr, ok := err.(*ValidationError); ok {
					logger.Warn("Input validation failed",
						zap.String("type", validationErr.Type),
						zap.String("field", validationErr.Field),
						zap.String("client_ip", r.RemoteAddr),
						zap.String("path", r.URL.Path),
						zap.String("user_agent", r.Header.Get("User-Agent")),
					)
				}
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ValidatedHandlerFunc wraps an http.HandlerFunc with input validation
func ValidatedHandlerFunc(validation *InputValidation, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := validation.ValidateRequest(r); err != nil {
			if validationErr, ok := err.(*ValidationError); ok {
				logger.Warn("Input validation failed",
					zap.String("type", validationErr.Type),
					zap.String("field", validationErr.Field),
					zap.String("client_ip", r.RemoteAddr),
					zap.String("path", r.URL.Path),
					zap.String("user_agent", r.Header.Get("User-Agent")),
				)
			}
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		handlerFunc(w, r)
	})
}

// SecureValidatedHandlerFunc combines security headers with input validation for regular handlers
func SecureValidatedHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return SecurityHandlerFunc(DefaultSecurityHeaders(), 
		ValidatedHandlerFunc(DefaultInputValidation(), handlerFunc))
}

// SecureValidatedAPIHandlerFunc combines security headers with input validation for API handlers
func SecureValidatedAPIHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return SecurityHandlerFunc(APISecurityHeaders(), 
		ValidatedHandlerFunc(APIInputValidation(), handlerFunc))
}