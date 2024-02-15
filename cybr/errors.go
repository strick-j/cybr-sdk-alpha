package cybr

// MissingSubdomainError is an error that is returned if Subdomain configuration
// value was not found.
type MissingSubdomainError struct{}

func (*MissingSubdomainError) Error() string {
	return "a CyberArk Subdomain is required, but was not found"
}

// MissingDomainError is an error that is returned if Domain configuration
// value was not found.
type MissingDomainError struct{}

func (*MissingDomainError) Error() string {
	return "a CyberArk Domain is required, but was not found"
}
