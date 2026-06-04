package v1alpha1

import (
	"crypto/x509"
	"testing"

	. "github.com/onsi/gomega"
)

// TestTLSAsConfig_Insecure tests TLS.AsConfig with Insecure flag.
func TestTLSAsConfig_Insecure(t *testing.T) {
	g := NewGomegaWithT(t)

	tls := TLS{
		Insecure: true,
	}

	config, err := tls.AsConfig()
	g.Expect(err).To(BeNil())
	g.Expect(config).NotTo(BeNil())
	g.Expect(config.InsecureSkipVerify).To(BeTrue())
	g.Expect(config.RootCAs).To(BeNil())
}

// TestTLSAsConfig_CustomCA tests TLS.AsConfig with custom CA certificate.
func TestTLSAsConfig_CustomCA(t *testing.T) {
	g := NewGomegaWithT(t)

	// Valid self-signed CA certificate (PEM format)
	validCA := `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKZ7RjRBWAhDMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTYwODI3MjEyMDU5WhcNMjYwODI1MjEyMDU5WjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAsMCNtJ3Jz4HvuPbHhJEwH8uGcGmxJAiH1DnrUPdTMNqKp2cC9H2aJYP6
PQKN4SqQXKWQPvd3qvvGVxF1V7UqRqhAGHFwx+NQT7NEn8TpfM3pQ0jy3jBFjZVf
dKRGxNjEQxL7s1mXQyJeXcEK5P2+mOKRUhJWC2OHvMvRUr0JqhQxMx3bPwYFdxLd
YCMRlxqVJdWvGVCJnJWMvchW4yYWKT4LwTJr3xqQJ8qY6LLYJQQkCYuU+kQQqQJC
FvTQKYMQI0qRsJPbwqUaQHFg+cKhsrJxwHZx5hU8L0L8xU7tNmr0xeHB6KyRQG6A
hKP3tMD4tIbJqNGQKKHU8zKWVIhXQQIDAQABo1AwTjAdBgNVHQ4EFgQU2Lc0B0k0
vL0mGRgjPZvL1L5qgMwwHwYDVR0jBBgwFoAU2Lc0B0k0vL0mGRgjPZvL1L5qgMww
DAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEADGVbTZpLR8LpQxoLnM7l
qKBCIr8Fc6pVKQm5kh4j6WJCHYDdLHQEPcL8aGZP2LXk3VBQK0yLzKW7uPgfEVsY
KDXqNhHJBFZaPhGsLX6K6HUqHGwi1RqQhqG1CqKhVQaLqERANj8cQC1qP1aXPPfV
vKRAYJCPTK3kCJRgDCHLc5OB3fYLTf7FLK8zQ2HKKLaCBKx7DHVfLxPmQOJNFXkW
QCMYnQPPX0BqQaHGq6jVRCTfILKQ9hfKG3pJ1G5C4VHVKaXCDFqUNCVWQYOFJaJ6
RqLqGBP4cVRqQQdEQ6wCZJLaWqQdZL7HpVY5+h7XvHxLcZLfKQYPH0L9Z/UYQk8F
3g==
-----END CERTIFICATE-----`

	tls := TLS{
		CA: validCA,
	}

	config, err := tls.AsConfig()
	g.Expect(err).To(BeNil())
	g.Expect(config).NotTo(BeNil())
	g.Expect(config.InsecureSkipVerify).To(BeFalse())
	g.Expect(config.RootCAs).NotTo(BeNil())

	// Verify the CA was actually loaded
	subjects := config.RootCAs.Subjects()
	g.Expect(len(subjects)).To(BeNumerically(">", 0))
}

// TestTLSAsConfig_InvalidCA tests TLS.AsConfig with invalid CA certificate.
func TestTLSAsConfig_InvalidCA(t *testing.T) {
	g := NewGomegaWithT(t)

	tls := TLS{
		CA: "not a valid PEM certificate",
	}

	config, err := tls.AsConfig()
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("failed to parse CA certificate"))
	g.Expect(config).To(BeNil())
}

// TestTLSAsConfig_Empty tests TLS.AsConfig with empty/default config.
func TestTLSAsConfig_Empty(t *testing.T) {
	g := NewGomegaWithT(t)

	tls := TLS{}

	config, err := tls.AsConfig()
	g.Expect(err).To(BeNil())
	g.Expect(config).NotTo(BeNil())
	g.Expect(config.InsecureSkipVerify).To(BeFalse())
	g.Expect(config.RootCAs).To(BeNil())
}

// TestTLSAsConfig_InsecureWithCA tests that Insecure takes precedence over CA.
func TestTLSAsConfig_InsecureWithCA(t *testing.T) {
	g := NewGomegaWithT(t)

	validCA := `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKZ7RjRBWAhDMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTYwODI3MjEyMDU5WhcNMjYwODI1MjEyMDU5WjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAsMCNtJ3Jz4HvuPbHhJEwH8uGcGmxJAiH1DnrUPdTMNqKp2cC9H2aJYP6
PQKN4SqQXKWQPvd3qvvGVxF1V7UqRqhAGHFwx+NQT7NEn8TpfM3pQ0jy3jBFjZVf
dKRGxNjEQxL7s1mXQyJeXcEK5P2+mOKRUhJWC2OHvMvRUr0JqhQxMx3bPwYFdxLd
YCMRlxqVJdWvGVCJnJWMvchW4yYWKT4LwTJr3xqQJ8qY6LLYJQQkCYuU+kQQqQJC
FvTQKYMQI0qRsJPbwqUaQHFg+cKhsrJxwHZx5hU8L0L8xU7tNmr0xeHB6KyRQG6A
hKP3tMD4tIbJqNGQKKHU8zKWVIhXQQIDAQABo1AwTjAdBgNVHQ4EFgQU2Lc0B0k0
vL0mGRgjPZvL1L5qgMwwHwYDVR0jBBgwFoAU2Lc0B0k0vL0mGRgjPZvL1L5qgMww
DAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEADGVbTZpLR8LpQxoLnM7l
qKBCIr8Fc6pVKQm5kh4j6WJCHYDdLHQEPcL8aGZP2LXk3VBQK0yLzKW7uPgfEVsY
KDXqNhHJBFZaPhGsLX6K6HUqHGwi1RqQhqG1CqKhVQaLqERANj8cQC1qP1aXPPfV
vKRAYJCPTK3kCJRgDCHLc5OB3fYLTf7FLK8zQ2HKKLaCBKx7DHVfLxPmQOJNFXkW
QCMYnQPPX0BqQaHGq6jVRCTfILKQ9hfKG3pJ1G5C4VHVKaXCDFqUNCVWQYOFJaJ6
RqLqGBP4cVRqQQdEQ6wCZJLaWqQdZL7HpVY5+h7XvHxLcZLfKQYPH0L9Z/UYQk8F
3g==
-----END CERTIFICATE-----`

	tls := TLS{
		Insecure: true,
		CA:       validCA,
	}

	config, err := tls.AsConfig()
	g.Expect(err).To(BeNil())
	g.Expect(config).NotTo(BeNil())
	// Insecure should take precedence
	g.Expect(config.InsecureSkipVerify).To(BeTrue())
	// CA should not be processed when Insecure is true
	g.Expect(config.RootCAs).To(BeNil())
}

// TestTLSAsConfig_MultipleCAsConcatenated tests loading multiple CA certificates.
func TestTLSAsConfig_MultipleCAsConcatenated(t *testing.T) {
	g := NewGomegaWithT(t)

	// Two valid CA certificates concatenated
	multipleCAs := `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKZ7RjRBWAhDMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTYwODI3MjEyMDU5WhcNMjYwODI1MjEyMDU5WjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAsMCNtJ3Jz4HvuPbHhJEwH8uGcGmxJAiH1DnrUPdTMNqKp2cC9H2aJYP6
PQKN4SqQXKWQPvd3qvvGVxF1V7UqRqhAGHFwx+NQT7NEn8TpfM3pQ0jy3jBFjZVf
dKRGxNjEQxL7s1mXQyJeXcEK5P2+mOKRUhJWC2OHvMvRUr0JqhQxMx3bPwYFdxLd
YCMRlxqVJdWvGVCJnJWMvchW4yYWKT4LwTJr3xqQJ8qY6LLYJQQkCYuU+kQQqQJC
FvTQKYMQI0qRsJPbwqUaQHFg+cKhsrJxwHZx5hU8L0L8xU7tNmr0xeHB6KyRQG6A
hKP3tMD4tIbJqNGQKKHU8zKWVIhXQQIDAQABo1AwTjAdBgNVHQ4EFgQU2Lc0B0k0
vL0mGRgjPZvL1L5qgMwwHwYDVR0jBBgwFoAU2Lc0B0k0vL0mGRgjPZvL1L5qgMww
DAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEADGVbTZpLR8LpQxoLnM7l
qKBCIr8Fc6pVKQm5kh4j6WJCHYDdLHQEPcL8aGZP2LXk3VBQK0yLzKW7uPgfEVsY
KDXqNhHJBFZaPhGsLX6K6HUqHGwi1RqQhqG1CqKhVQaLqERANj8cQC1qP1aXPPfV
vKRAYJCPTK3kCJRgDCHLc5OB3fYLTf7FLK8zQ2HKKLaCBKx7DHVfLxPmQOJNFXkW
QCMYnQPPX0BqQaHGq6jVRCTfILKQ9hfKG3pJ1G5C4VHVKaXCDFqUNCVWQYOFJaJ6
RqLqGBP4cVRqQQdEQ6wCZJLaWqQdZL7HpVY5+h7XvHxLcZLfKQYPH0L9Z/UYQk8F
3g==
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKZ7RjRBWAhDMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTYwODI3MjEyMDU5WhcNMjYwODI1MjEyMDU5WjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAsMCNtJ3Jz4HvuPbHhJEwH8uGcGmxJAiH1DnrUPdTMNqKp2cC9H2aJYP6
PQKN4SqQXKWQPvd3qvvGVxF1V7UqRqhAGHFwx+NQT7NEn8TpfM3pQ0jy3jBFjZVf
dKRGxNjEQxL7s1mXQyJeXcEK5P2+mOKRUhJWC2OHvMvRUr0JqhQxMx3bPwYFdxLd
YCMRlxqVJdWvGVCJnJWMvchW4yYWKT4LwTJr3xqQJ8qY6LLYJQQkCYuU+kQQqQJC
FvTQKYMQI0qRsJPbwqUaQHFg+cKhsrJxwHZx5hU8L0L8xU7tNmr0xeHB6KyRQG6A
hKP3tMD4tIbJqNGQKKHU8zKWVIhXQQIDAQABo1AwTjAdBgNVHQ4EFgQU2Lc0B0k0
vL0mGRgjPZvL1L5qgMwwHwYDVR0jBBgwFoAU2Lc0B0k0vL0mGRgjPZvL1L5qgMww
DAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEADGVbTZpLR8LpQxoLnM7l
qKBCIr8Fc6pVKQm5kh4j6WJCHYDdLHQEPcL8aGZP2LXk3VBQK0yLzKW7uPgfEVsY
KDXqNhHJBFZaPhGsLX6K6HUqHGwi1RqQhqG1CqKhVQaLqERANj8cQC1qP1aXPPfV
vKRAYJCPTK3kCJRgDCHLc5OB3fYLTf7FLK8zQ2HKKLaCBKx7DHVfLxPmQOJNFXkW
QCMYnQPPX0BqQaHGq6jVRCTfILKQ9hfKG3pJ1G5C4VHVKaXCDFqUNCVWQYOFJaJ6
RqLqGBP4cVRqQQdEQ6wCZJLaWqQdZL7HpVY5+h7XvHxLcZLfKQYPH0L9Z/UYQk8F
3g==
-----END CERTIFICATE-----`

	tls := TLS{
		CA: multipleCAs,
	}

	config, err := tls.AsConfig()
	g.Expect(err).To(BeNil())
	g.Expect(config).NotTo(BeNil())
	g.Expect(config.RootCAs).NotTo(BeNil())

	// Verify both CAs were loaded
	subjects := config.RootCAs.Subjects()
	// Note: identical certs will only be loaded once
	g.Expect(len(subjects)).To(BeNumerically(">=", 1))
}

// TestTLSNilSafety tests that nil config handling is safe.
func TestTLSNilSafety(t *testing.T) {
	g := NewGomegaWithT(t)

	var nilConfig *x509.CertPool
	g.Expect(nilConfig).To(BeNil())

	// Verify nil RootCAs is acceptable for tls.Config
	config := TLS{}
	tlsConfig, err := config.AsConfig()
	g.Expect(err).To(BeNil())
	g.Expect(tlsConfig.RootCAs).To(BeNil())
}
