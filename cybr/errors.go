package cybr

// MissingSubDomainError is an error that is returned if Sub Domain configuration
// value was not found.
type MissingSubDomainError struct{}

func (*MissingSubDomainError) Error() string {
	return "a CyberArk Sub Domain is required, but was not found"
}
