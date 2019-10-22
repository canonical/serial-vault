// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
 * License granted by Canonical Limited
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package datastore

import (
	"fmt"
	"regexp"
	"strings"
)

const defaultNicknamePattern = `^\S+$`

func validateNotEmpty(fieldName, fieldValue string) error {
	if len(strings.TrimSpace(fieldValue)) == 0 {
		return fmt.Errorf("%v must not be empty", normalize(fieldName))
	}
	return nil
}

func validateStringsNotEmpty(args ...string) bool {
	for _, s := range args {
		if len(strings.TrimSpace(s)) == 0 {
			return false
		}
	}
	return true
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
