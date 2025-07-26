package actions

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// sslCertCheckAction checks SSL certificate validity and details
// Args: [host] - hostname or hostname:port to check (default port 443)
// Options:
//   - timeout: connection timeout duration (default: "5s")
//   - verify_chain: validate full certificate chain (default: true)
//   - check_expiry_days: warn if expires within N days (default: 30)
//   - allow_self_signed: accept self-signed certificates (default: false)
//   - skip_hostname_verify: skip hostname verification (default: false)
func sslCertCheckAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 1 {
		return types.MissingArgsError("ssl_cert_check", 1, len(args))
	}

	// Validate arguments are resolved
	if errorResult := validateArgsResolved("ssl_cert_check", args); errorResult != nil {
		return *errorResult
	}

	hostArg := fmt.Sprintf("%v", args[0])
	if hostArg == "" {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "EMPTY_HOST").
			WithTemplate("SSL certificate check host cannot be empty").
			WithSuggestion("Provide a valid hostname or hostname:port").
			Build("empty host provided")
	}

	// Parse host and port
	host, port := parseHostPort(hostArg)
	if host == "" {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_HOST").
			WithTemplate("Invalid host format for SSL certificate check").
			WithContext("host", hostArg).
			WithSuggestion("Use format: hostname or hostname:port").
			Build(fmt.Sprintf("invalid host format: %s", hostArg))
	}

	// Parse options
	timeout := parseTimeout(options, "5s")
	verifyChain := parseBoolOption(options, "verify_chain", true)
	checkExpiryDays := parseIntOption(options, "check_expiry_days", 30)
	allowSelfSigned := parseBoolOption(options, "allow_self_signed", false)
	skipHostnameVerify := parseBoolOption(options, "skip_hostname_verify", false)

	// Validate timeout
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_TIMEOUT").
			WithTemplate("Invalid timeout format for SSL certificate check").
			WithContext("timeout", timeout).
			WithContext("valid_examples", "5s, 1000ms, 10s").
			WithSuggestion("Use Go duration format: ns, us, ms, s, m, h").
			Build(fmt.Sprintf("invalid timeout format: %s", timeout))
	}

	// Validate expiry days
	if checkExpiryDays < 0 || checkExpiryDays > 365 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "INVALID_EXPIRY_DAYS").
			WithTemplate("SSL certificate expiry check days must be between 0 and 365").
			WithContext("check_expiry_days", checkExpiryDays).
			WithSuggestion("Use a reasonable number of days (7-90 for most cases)").
			Build(fmt.Sprintf("invalid expiry days: %d", checkExpiryDays))
	}

	// Execute SSL certificate check
	result := performSSLCheck(host, port, timeoutDuration, verifyChain, checkExpiryDays, allowSelfSigned, skipHostnameVerify)
	return result
}

// performSSLCheck executes the actual SSL certificate check
func performSSLCheck(host string, port int, timeout time.Duration, verifyChain bool, checkExpiryDays int, allowSelfSigned bool, skipHostnameVerify bool) types.ActionResult {
	address := fmt.Sprintf("%s:%d", host, port)
	
	fmt.Printf("🔒 Checking SSL certificate for %s...\n", address)

	// Configure TLS connection
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: !verifyChain || allowSelfSigned || skipHostnameVerify,
	}

	// Set up dialer with timeout
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	// Connect to the server
	conn, err := tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	if err != nil {
		// Check for specific error types
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return types.NewFailureBuilder(types.FailureCategoryResponse, "SSL_CONNECTION_TIMEOUT").
				WithTemplate("SSL connection to server timed out").
				WithContext("host", host).
				WithContext("port", port).
				WithContext("timeout", timeout.String()).
				WithSuggestion("Check network connectivity and increase timeout if needed").
				WithSuggestion("Verify the host is running an SSL/TLS service on the specified port").
				Build(fmt.Sprintf("SSL connection timeout for %s", address))
		}

		if strings.Contains(err.Error(), "certificate") {
			return types.NewFailureBuilder(types.FailureCategoryValidation, "SSL_CERTIFICATE_ERROR").
				WithTemplate("SSL certificate validation failed").
				WithContext("host", host).
				WithContext("port", port).
				WithContext("error", err.Error()).
				WithSuggestion("Check if the certificate is valid and trusted").
				WithSuggestion("Use allow_self_signed: true for self-signed certificates").
				Build(fmt.Sprintf("SSL certificate error for %s: %s", address, err.Error()))
		}

		return types.NewErrorBuilder(types.ErrorCategoryNetwork, "SSL_CONNECTION_FAILED").
			WithTemplate("Failed to establish SSL connection").
			WithContext("host", host).
			WithContext("port", port).
			WithContext("error", err.Error()).
			WithSuggestion("Verify the host is reachable and running an SSL/TLS service").
			WithSuggestion("Check firewall settings and port availability").
			Build(fmt.Sprintf("SSL connection failed for %s: %s", address, err.Error()))
	}
	defer conn.Close()

	// Get certificate chain
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return types.NewFailureBuilder(types.FailureCategoryValidation, "NO_CERTIFICATES").
			WithTemplate("No SSL certificates found in connection").
			WithContext("host", host).
			WithContext("port", port).
			WithSuggestion("Verify the server is configured to present SSL certificates").
			Build(fmt.Sprintf("no certificates found for %s", address))
	}

	// Analyze the leaf certificate (first in chain)
	cert := certs[0]
	now := time.Now()
	
	// Calculate expiry information
	daysUntilExpiry := int(cert.NotAfter.Sub(now).Hours() / 24)
	isExpired := now.After(cert.NotAfter)
	isNotYetValid := now.Before(cert.NotBefore)
	expiryWarning := daysUntilExpiry <= checkExpiryDays && daysUntilExpiry > 0

	// Check for hostname verification if not skipped
	var hostnameError string
	if !skipHostnameVerify && verifyChain {
		if err := cert.VerifyHostname(host); err != nil {
			hostnameError = err.Error()
		}
	}

	// Perform additional certificate chain verification if requested
	var chainError string
	if verifyChain && !allowSelfSigned {
		roots, err := x509.SystemCertPool()
		if err == nil {
			opts := x509.VerifyOptions{
				Roots:         roots,
				Intermediates: x509.NewCertPool(),
			}
			
			// Add intermediate certificates to the pool
			for i := 1; i < len(certs); i++ {
				opts.Intermediates.AddCert(certs[i])
			}
			
			_, err = cert.Verify(opts)
			if err != nil {
				chainError = err.Error()
			}
		}
	}

	// Determine overall validity
	valid := !isExpired && !isNotYetValid && hostnameError == "" && chainError == ""
	
	// Check if it's self-signed
	isSelfSigned := cert.Issuer.String() == cert.Subject.String()

	// Build result data
	resultData := map[string]any{
		"valid":               valid,
		"host":                host,
		"port":                port,
		"expires_at":          cert.NotAfter.Format(time.RFC3339),
		"valid_from":          cert.NotBefore.Format(time.RFC3339),
		"days_until_expiry":   daysUntilExpiry,
		"is_expired":          isExpired,
		"is_not_yet_valid":    isNotYetValid,
		"expiry_warning":      expiryWarning,
		"issuer":              cert.Issuer.String(),
		"subject":             cert.Subject.String(),
		"serial_number":       cert.SerialNumber.String(),
		"signature_algorithm": cert.SignatureAlgorithm.String(),
		"public_key_algorithm": cert.PublicKeyAlgorithm.String(),
		"chain_length":        len(certs),
		"self_signed":         isSelfSigned,
		"dns_names":           cert.DNSNames,
		"ip_addresses":        cert.IPAddresses,
		"hostname_error":      hostnameError,
		"chain_error":         chainError,
	}

	// Add key size information if available
	if cert.PublicKey != nil {
		switch key := cert.PublicKey.(type) {
		case *interface{}:
			// Handle different key types
			resultData["key_info"] = fmt.Sprintf("%T", key)
		}
	}

	// Determine status and provide appropriate feedback
	if !valid {
		var issues []string
		if isExpired {
			issues = append(issues, "certificate has expired")
		}
		if isNotYetValid {
			issues = append(issues, "certificate is not yet valid")
		}
		if hostnameError != "" {
			issues = append(issues, fmt.Sprintf("hostname verification failed: %s", hostnameError))
		}
		if chainError != "" {
			issues = append(issues, fmt.Sprintf("certificate chain verification failed: %s", chainError))
		}

		failureResult := types.NewFailureBuilder(types.FailureCategoryValidation, "SSL_CERTIFICATE_INVALID").
			WithTemplate("SSL certificate validation failed").
			WithContext("host", host).
			WithContext("port", port).
			WithContext("issues", strings.Join(issues, "; ")).
			WithContext("expires_at", cert.NotAfter.Format(time.RFC3339)).
			WithContext("days_until_expiry", daysUntilExpiry).
			WithSuggestion("Renew the SSL certificate before it expires").
			WithSuggestion("Verify certificate matches the hostname").
			Build(fmt.Sprintf("SSL certificate invalid for %s: %s", address, strings.Join(issues, "; ")))
		
		// Add the data to the failure result
		failureResult.Data = resultData
		return failureResult
	}

	// Provide warning for upcoming expiry
	if expiryWarning {
		fmt.Printf("⚠️  Certificate expires in %d days (%s)\n", daysUntilExpiry, cert.NotAfter.Format("2006-01-02"))
	}

	fmt.Printf("✅ SSL certificate valid for %s (expires: %s)\n", address, cert.NotAfter.Format("2006-01-02"))

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   resultData,
	}
}

// Helper functions
func parseHostPort(hostArg string) (string, int) {
	// Check if port is specified
	if strings.Contains(hostArg, ":") {
		parts := strings.Split(hostArg, ":")
		if len(parts) == 2 {
			host := strings.TrimSpace(parts[0])
			if port := parseInt(parts[1]); port > 0 && port <= 65535 {
				return host, port
			}
		}
		return "", 0
	}
	
	// Default to port 443 for HTTPS
	return strings.TrimSpace(hostArg), 443
}

func parseTimeout(options map[string]any, defaultValue string) string {
	if timeoutVal, exists := options["timeout"]; exists {
		return fmt.Sprintf("%v", timeoutVal)
	}
	return defaultValue
}

func parseBoolOption(options map[string]any, key string, defaultValue bool) bool {
	if val, exists := options[key]; exists {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
		if strVal, ok := val.(string); ok {
			return strings.ToLower(strVal) == "true"
		}
	}
	return defaultValue
}

func parseIntOption(options map[string]any, key string, defaultValue int) int {
	if val, exists := options[key]; exists {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		if strVal, ok := val.(string); ok {
			if parsed := parseInt(strVal); parsed >= 0 {
				return parsed
			}
		}
	}
	return defaultValue
}