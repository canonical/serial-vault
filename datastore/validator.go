package datastore

import (
	"fmt"
	"regexp"
	"strings"
)

const defaultNicknamePattern = "^[a-z0-9](?:-?[a-z0-9])*$"

func validateNotEmpty(fieldName, fieldValue string) error {
	if len(fieldValue) == 0 {
		return fmt.Errorf("%v must not be empty", fieldName)
	}
	return nil
}

func validateSyntax(fieldName, fieldValue string, regularExpression *regexp.Regexp) error {
	if err := validateNotEmpty(fieldName, fieldValue); err != nil {
		return err
	}

	if strings.ToLower(fieldValue) != fieldValue {
		return fmt.Errorf("%v must not contain uppercase characters", fieldName)
	}

	if !regularExpression.MatchString(fieldValue) {
		return fmt.Errorf("%v contains invalid characters, allowed %q", fieldName, regularExpression)
	}

	return nil
}
