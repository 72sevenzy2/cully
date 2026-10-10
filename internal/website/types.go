package website

import "errors"

// MaxPhoto is the maximum upload or imported-photo size in bytes.
const MaxPhoto = 5 << 20

const (
	maxRequestsPerMinute  = 20
	maxProfileRequestBody = 1024
)

// Validation errors retain the messages shown by the submission form.
var (
	errInvalidLinkedInPhoto   = errors.New("invalid LinkedIn photo URL")
	errPhotoUnavailable       = errors.New("photo unavailable")
	errPhotoTooLarge          = errors.New("photo too large")
	errMissingConsent         = errors.New("Please agree to publication before submitting.")
	errInvalidNameLength      = errors.New("Enter a name up to 80 characters.")
	errInvalidQuoteLength     = errors.New("Write a testimonial between 20 and 1,000 characters.")
	errInvalidRoleLength      = errors.New("Keep role and workplace within 100 characters each.")
	errInvalidLinkedInURL     = errors.New("Enter a LinkedIn profile URL starting with https://www.linkedin.com/in/.")
	errUnsupportedPhotoFormat = errors.New("Choose a JPG, PNG, or WebP photo.")
	errInvalidPhotoDimensions = errors.New("Choose a valid photo no larger than 4096 × 4096 pixels.")
	errInvalidReviewState     = errors.New("invalid review state")
	errInvalidSubmissionID    = errors.New("invalid submission ID")
)
