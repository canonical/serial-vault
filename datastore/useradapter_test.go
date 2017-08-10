// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2017 Canonical Ltd
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
	"strings"
	"testing"
)

func TestUserName(t *testing.T) {
	username := "myusername"
	err := validateUsername(username)
	if err != nil {
		t.Errorf("Not a valid username: %v", err)
	}
}

func TestUsernameEmpty(t *testing.T) {
	username := ""
	err := validateUsername(username)
	if err == nil {
		t.Error("Expected username not to be valid, but it is")
	}
	if err.Error() != "Username must not be empty" {
		t.Error("Error happening is not the one searched for")
	}
}

func TestUsernameInvalidChar(t *testing.T) {
	username := "myusername_"
	err := validateUsername(username)
	if err == nil {
		t.Error("Expected username not to be valid, but it is")
	}
	if !strings.Contains(err.Error(), "Username contains invalid characters") {
		t.Error("Error happening is not the one searched for")
	}
}

func TestUsernameUpperChar(t *testing.T) {
	username := "myUsername"
	err := validateUsername(username)
	if err == nil {
		t.Error("Expected username not to be valid, but it is")
	}
	if err.Error() != "Username must not contain uppercase characters" {
		t.Error("Error happening is not the one searched for")
	}
}

func TestUserFullName(t *testing.T) {
	name := "Federico Martín Bahamontes"
	err := validateUserFullName(name)
	if err != nil {
		t.Errorf("Not a valid name: %v", err)
	}
}

func TestUserFullNameEmpty(t *testing.T) {
	name := ""
	err := validateUserFullName(name)
	if err == nil {
		t.Error("Expected name not to be valid, but it is")
	}
	if err.Error() != "Name must not be empty" {
		t.Error("Error happening is not the one searched for")
	}
}

func TestUserEmail(t *testing.T) {
	email := "my@email.com"
	err := validateUserEmail(email)
	if err != nil {
		t.Errorf("Not a valid email: %v", err)
	}
}

func TestUserEmailUpperChar(t *testing.T) {
	email := "my@eMail.com"
	err := validateUserEmail(email)
	if err == nil {
		t.Error("Expected email not to be valid, but it is")
	}
	if err.Error() != "Email must not contain uppercase characters" {
		t.Error("Error happening is not the one searched for")
	}
}

func TestUserEmailWithoutAt(t *testing.T) {
	email := "mymail.com"
	err := validateUserEmail(email)
	if err == nil {
		t.Error("Expected email not to be valid, but it is")
	}
	if !strings.Contains(err.Error(), "Email contains invalid characters") {
		t.Error("Error happening is not the one searched for")
	}
}

func TestUserEmailWithoutDot(t *testing.T) {
	email := "my@mailcom"
	err := validateUserEmail(email)
	if err == nil {
		t.Error("Expected email not to be valid, but it is")
	}
	if !strings.Contains(err.Error(), "Email contains invalid characters") {
		t.Error("Error happening is not the one searched for")
	}
}

func TestUserEmailWithLargeDomain(t *testing.T) {
	email := "my@mail.domain"
	err := validateUserEmail(email)
	if err == nil {
		t.Error("Expected email not to be valid, but it is")
	}
	if !strings.Contains(err.Error(), "Email contains invalid characters") {
		t.Error("Error happening is not the one searched for")
	}
}
