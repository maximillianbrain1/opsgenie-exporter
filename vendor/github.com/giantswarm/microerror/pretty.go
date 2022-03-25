package microerror

import (
	"errors"
	"strings"
	"unicode"
)

const (
	prefix    = "error: "
	delimiter = ": "
)

func Pretty(err error, stackTrace bool) string {
	var message strings.Builder

	// Check if it's an annotated error.
	var aErr *annotatedError
	if errors.As(err, &aErr) {
		capitalizeAnnotation := true

		if aErr.underlying.Kind != kindNil && aErr.underlying.Kind != kindUnknown {
			message.WriteString(prettifyErrorMessage(aErr.underlying.Error(), true))
			message.WriteString(delimiter)
			capitalizeAnnotation = false
		}
		message.WriteString(prettifyErrorMessage(aErr.annotation, capitalizeAnnotation))
	} else {
		// This is either an unmasked microerror, or
		// a simple 'errors.New()' error.
		pretty := prettifyErrorMessage(err.Error(), true)
		if len(pretty) < 1 {
			return ""
		}
		message.WriteString(pretty)
	}

	if stackTrace {
		// Add formatted stack trace.
		if sErr, ok := err.(*stackedError); ok {
			message.WriteString("\n")
			trace := createStackTrace(sErr)
			message.WriteString(formatStackTrace(trace))
		}
	}

	return message.String()
}

func prettifyErrorMessage(message string, capitalize bool) string {
	if len(message) < 1 {
		return message
	}

	// Remove the 'error: ' prefix if it exists
	if strings.HasPrefix(strings.ToLower(message), prefix) {
		message = message[len(prefix):]
	}
	// This suffix is usually present in microerrors
	// without annotations.
	message = strings.TrimSuffix(message, " error")

	if capitalize {
		// Capitalize the first letter.
		tmpMessage := []rune(message)
		tmpMessage[0] = unicode.ToUpper(tmpMessage[0])
		message = string(tmpMessage)
	}

	return message
}
