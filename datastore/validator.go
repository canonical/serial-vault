package datastore

import (
	"fmt"
	"regexp"
	"strings"
)

const defaultNicknamePattern = "^[a-z0-9](?:-?[a-z0-9])*$"

func validateNotEmpty(fieldName, fieldValue string) error {
	if len(strings.TrimSpace(fieldValue)) == 0 {
		return fmt.Errorf("%v must not be empty", normalize(fieldName))
	}
	return nil
}

func validateSyntax(fieldName, fieldValue string, regularExpression *regexp.Regexp) error {
	if err := validateNotEmpty(fieldName, fieldValue); err != nil {
		return err
	}

	if strings.ToLower(fieldValue) != fieldValue {
		return fmt.Errorf("%v must not contain uppercase characters", normalize(fieldName))
	}

	if !regularExpression.MatchString(fieldValue) {
		return fmt.Errorf("%v contains invalid characters, allowed %q", normalize(fieldName), regularExpression)
	}

	return nil
}

func validateAuthorityID(AuthorityID string) error {
	return validateNotEmpty("Authority ID", AuthorityID)
}

func validateCaseInsensitive(fieldName, fieldValue string) error {
	if err := validateNotEmpty(fieldName, fieldValue); err != nil {
		return err
	}

	if strings.ToLower(fieldValue) != fieldValue {
		return fmt.Errorf("%v must not contain uppercase characters", normalize(fieldName))
	}
	return nil
}

func normalize(fieldName string) string {
	theFieldName := strings.TrimSpace(fieldName)
	if len(theFieldName) == 0 {
		theFieldName = "field"
	}
	return theFieldName
}
