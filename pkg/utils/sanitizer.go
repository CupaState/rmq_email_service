package utils

import "github.com/microcosm-cc/bluemonday"

// func to sanitize string
func SanitizeString(s string) string {
	ugcPolicy := bluemonday.UGCPolicy()
	return ugcPolicy.Sanitize(s)
}
