package internal

import "strings"

func UrlTrancate(url string) string {
	result := url
	anchorIdx := strings.Index(result, "#")
	if anchorIdx > 0 {
		result = result[:anchorIdx]
	}
	queryIdx := strings.Index(result, "?")
	if queryIdx > 0 {
		result = result[:queryIdx]
	}
	return result
}
