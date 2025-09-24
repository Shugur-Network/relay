package web

import (
	"net/http"
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