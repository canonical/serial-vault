package datastore

import (
	"regexp"
	"testing"

	check "gopkg.in/check.v1"
)

func TestValidator(t *testing.T) { check.TestingT(t) }

type validatorSuite struct{}

var _ = check.Suite(&validatorSuite{})

func (vs *validatorSuite) TestValidateNotEmptyHappyPath(c *check.C) {
	fieldName := "fieldName"
	fieldValue := "fieldValue"
	err := validateNotEmpty(fieldName, fieldValue)
	c.Assert(err, check.IsNil)
}

func (vs *validatorSuite) TestValidateNotEmptyIsEmpty(c *check.C) {
	fieldName := "fieldName"
	fieldValue := ""
	err := validateNotEmpty(fieldName, fieldValue)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "fieldName must not be empty")
}

func (vs *validatorSuite) TestValidateNotEmptyTrailingSpace(c *check.C) {
	fieldName := "fieldName"
	fieldValue := " "
	err := validateNotEmpty(fieldName, fieldValue)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "fieldName must not be empty")
}

func (vs *validatorSuite) TestValidateNotEmptyFieldNameEmpty(c *check.C) {
	fieldName := ""
	fieldValue := "fieldValue"
	err := validateNotEmpty(fieldName, fieldValue)
	c.Assert(err, check.IsNil)
}

func (vs *validatorSuite) TestValidateNotEmptyFieldNameEmptyFieldValueEmpty(c *check.C) {
	fieldName := ""
	fieldValue := ""
	err := validateNotEmpty(fieldName, fieldValue)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "field must not be empty")
}

func (vs *validatorSuite) TestValidateSyntaxHappyPath(c *check.C) {
	fieldName := "fieldName"
	fieldValue := "fieldvalue"
	pattern := "[a-zA-Z0-9]"
	err := validateSyntax(fieldName, fieldValue, regexp.MustCompile(pattern))
	c.Assert(err, check.IsNil)
}

func (vs *validatorSuite) TestValidateSyntaxFieldValueEmpty(c *check.C) {
	fieldName := "fieldName"
	fieldValue := ""
	pattern := "[a-zA-Z0-9]"
	err := validateSyntax(fieldName, fieldValue, regexp.MustCompile(pattern))
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "fieldName must not be empty")
}

func (vs *validatorSuite) TestValidateSyntaxFieldValueUpperCase(c *check.C) {
	fieldName := "fieldName"
	fieldValue := "fieldValue"
	pattern := "[a-zA-Z0-9]"
	err := validateSyntax(fieldName, fieldValue, regexp.MustCompile(pattern))
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "fieldName must not contain uppercase characters")
}

func (vs *validatorSuite) TestValidateSyntaxFieldValueDontMatchPattern(c *check.C) {
	fieldName := "fieldName"
	fieldValue := "fieldval with_invalid_chars"
	pattern := defaultNicknamePattern
	err := validateSyntax(fieldName, fieldValue, regexp.MustCompile(pattern))
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Matches, "fieldName contains invalid characters.*")
}

func (vs *validatorSuite) TestAuthorityIDHappyPath(c *check.C) {
	authorityID := "JADNF9478NA84MAPD8"
	err := validateAuthorityID(authorityID)
	c.Assert(err, check.IsNil)
}

func (vs *validatorSuite) TestAuthorityIDEmpty(c *check.C) {
	authorityID := ""
	err := validateAuthorityID(authorityID)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "Authority ID must not be empty")
}

func (vs *validatorSuite) TestAuthorityIDTrailingSpace(c *check.C) {
	authorityID := " "
	err := validateAuthorityID(authorityID)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "Authority ID must not be empty")
}
