package website

import (
	"errors"
)

// global types for any arbitrary value type ment for server.go and linkedin.go

const (
	MaxRequestsPerIPLimit = 20
)

// centralised errors, one-time initialisation avoids errors.New(..) overhead on every error case.
var InvalidLinkedinPhotoErr = errors.New("invalid LinkedIn photo URL")
var PhotoUnavailableErr = errors.New("photo unavailable")
var PhotoTooLargeErr = errors.New("photo too large")
var InvalidTestimonialConsentErr = errors.New("Please agree to publication before submitting.")
var InvalidNameLengthErr = errors.New("Enter a name up to 80 characters.")
var InvalidTestimonialLengthErr = errors.New("Write a testimonial between 20 and 1,000 characters.")
var InvalidRolesLengthErr = errors.New("Keep role and workplace within 100 characters each.")
var InvalidLinkedinURLErr = errors.New("Enter a LinkedIn profile URL starting with https://www.linkedin.com/in/.")