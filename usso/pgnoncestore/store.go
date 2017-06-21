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

package pgnoncestore

import (
	"time"

	"github.com/CanonicalLtd/serial-vault/service"

	"gopkg.in/errgo.v1"
)

// PgNonceStore is a nonce store backed by PostgreSQL
type PgNonceStore struct {
	DB *service.DB
}

// Accept implements openid.NonceStore.Accept
func (s *PgNonceStore) Accept(endpoint, nonce string) error {
	return s.accept(endpoint, nonce, time.Now())
}

// accept is the implementation of Accept. The third parameter is the
// current time, useful for testing.
func (s *PgNonceStore) accept(endpoint, nonce string, now time.Time) error {
	// From the openid specification:
	//
	// openid.response_nonce
	//
	// Value: A string 255 characters or less in length, that MUST be
	// unique to this particular successful authentication response.
	// The nonce MUST start with the current time on the server, and
	// MAY contain additional ASCII characters in the range 33-126
	// inclusive (printable non-whitespace characters), as necessary
	// to make each response unique. The date and time MUST be
	// formatted as specified in section 5.6 of [RFC3339], with the
	// following restrictions:
	//
	// + All times must be in the UTC timezone, indicated with a "Z".
	//
	// + No fractional seconds are allowed
	//
	// For example: 2005-05-15T17:11:51ZUNIQUE

	if len(nonce) < 20 {
		return errgo.Newf("%q does not contain a valid timestamp", nonce)
	}
	t, err := time.Parse(time.RFC3339, nonce[:20])
	if err != nil {
		return errgo.Notef(err, "%q does not contain a valid timestamp", nonce)
	}
	if t.Before(now.Add(service.OpenidNonceMaxAge)) {
		return errgo.Newf("%q too old", nonce)
	}

	openidNonce := service.OpenidNonce{Nonce: nonce, Endpoint: endpoint, TimeStamp: t.Unix()}
	err = s.DB.CreateOpenidNonce(openidNonce)
	return errgo.Mask(err)
}
