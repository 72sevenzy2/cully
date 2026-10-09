package website

import (
	"errors"
)

// global types for any arbitrary value type ment for server.go and linkedin.go


// centralised errors, one-time initialisation avoids errors.New(..) overhead on every error case.
var InvalidLinkedinURLError = errors.New("invalid LinkedIn photo URL")
var PhotoUnavailableErr = errors.New("photo unavailable")
var PhotoTooLargeErr = errors.New("photo too large")