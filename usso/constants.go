// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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

package usso

const jwtSecret = "TODO-ReplaceWithASecretFromTheConfigurationFile"

// ClaimsKey is the context key for the JWT claims
var ClaimsKey struct{}

// UserClaims holds the JWT custom claims for a user
const (
	ClaimsIdentity         = "identity"
	ClaimsUsername         = "username"
	ClaimsEmail            = "email"
	ClaimsName             = "name"
	ClaimsRole             = "role"
	StandardClaimExpiresAt = "exp"
)

// JWTCookie is the name of the cookie used to store the JWT
const JWTCookie = "X-Auth-Token"
