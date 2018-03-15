// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018-2019 Canonical Ltd
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

package pivot_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/pivot"
	"github.com/snapcore/snapd/asserts"
	check "gopkg.in/check.v1"
)

func TestPivotSuite(t *testing.T) { check.TestingT(t) }

type PivotSuite struct{}

type PivotTest struct {
	Method  string
	URL     string
	Data    []byte
	Code    int
	Type    string
	APIKey  string
	Success bool
}

var _ = check.Suite(&PivotSuite{})

const jsonType = "application/json; charset=UTF-8"

const serialAssert = `type: serial
authority-id: generic
brand-id: generic
model: generic-classic
serial: 54bdfe2e-607a-4d37-a469-68397070ec0c
device-key:
    AcbBTQRWhcGAARAAw0qcnXY26JyBvYpCigzIhWYTAnnHyjyB/MjA1lRVgYqeJUyrsUExj42q4ABC
    /kgn2yS36DP50ZeyOJDF7bHce7ELNmuknFUnSVL/r/VUYZrMmc3RyuV+7sYg042n++tqPbm0XkHS
    4Fpy3mC78a3+I8weUuhUaVaYYTEhKXX9FMPRF/qKW7Avt8v04LAaS3iNnYUk+yAyDYqyE/WedB22
    /GjiYnjk7Z5TpevuV5g4N9BiY12RuJWDOELbYR+Y8SR+js5YUH95uy5Nd2qI7uF8RPh0liSc5Tsz
    hBy+vwhBcrUTx1JN55bFqDKVtRBcjHe/ynY32Cg0inooB+yqBRbRUXK2LAZKwb4WAyo7xY923b/h
    p5KcMjmrRGd0AlBTum8FaFEm6M746ZYjSeW/X3wpJCWXnNkwe4mNkVkZJ9TgCh2MlzhtaLMkaqlM
    MldmwG1+AAE3SGCuZU6rg0BbPXp3iSqLPEIIXAV+l1vY9sxGrITQ61PN3ZqDdx0nCQxUVbha0ke9
    fFv8ZS/XsletkXME/mFvkhhGd2eaow0+B+SSaPWsaVXC/HIlzShXAZfDHI2dy1kfKmPcdUeSlqlH
    0zL2+H/h98+5vn71PvU1UXuYEg57aqDqKEvRjyU7hXN66JRzBn+j4ppjK4fo7hQJNIerEMYEKC5B
    hDkjXVlNalTrcI0AEQEAAQ==
device-key-sha3-384: cP_JQakFySqvoRbC6RiTP4ik-YXrAK1xFX_V4qN4WWqdGC0X817QpOX6SJf77E_U
timestamp: 2017-10-11T07:23:40.387695Z
sign-key-sha3-384: wrfougkz3Huq2T_KklfnufCC0HzG7bJ9wP99GV0FF-D3QH3eJtuSRlQc2JhrAoh1

AcLBUgQAAQoABgUCWd3G/AAA3SYQAJC7A18fmY4NqxLziqoVW65IPNjegi7I5NqJJhPycpedHlbA
mpZR93a2op+I+ssCj65S0YKTJJHtmeEeVFP2ns9sK3aXJIWSGW3zvVkcuVJlylpW6zFlREjM986I
9XRZR/lZ/2bMZzut9ZsYxEfzvcGiMTrqqUB0VToVqemrV6rQgmZ0k0BAoTNh4EIpxY83ruBUQqZX
X/iJOUV2dVXAai4XQ6XZRQlGTtJgVMcpNbm2pPKa71eYstChtGUkKp4klBu8hOcgJVpcIj+bSHZi
3Zzkx5jMvc6R5qOumTB26Gono8jHcYmn5BHdqZr/7x6UxDXAQm1ac1XklvICK7d7pjI4G/53uROV
aQKwxUyvVgZNPI3/ieN1L7MriNVFjTfQm7cJq8zVYL26yR0r2B9UGLxP8rG3Tr2S+ySVadqH44yh
9r/P5ZOfezOTnTeBL9JAHJeXuQNMO2VZFJWhCKZyHCAtCytGwDjCMMTOxqdkySpsM0VsBz4TizEc
InRto3Z45nadl7RhIK85zF4zycTYXveQsCTZ0zkp5Jfgm90rJ1qM0aKfQiuSOdY14YQZqQG1Fc3w
iSriN2otHAUwCXgr7vg7wiHOZKHf/LzV0/110iIHGxFRpuitZB+oHXw4EBOPPnDwpiFjBsbgspl9
4VDviwTwZm6obnuDEAxynpe6lhNz`

const serialAssertNonReseller = `type: serial
authority-id: vendor
brand-id: vendor
model: generic-classic
serial: 54bdfe2e-607a-4d37-a469-68397070ec0c
device-key:
    AcbBTQRWhcGAARAAw0qcnXY26JyBvYpCigzIhWYTAnnHyjyB/MjA1lRVgYqeJUyrsUExj42q4ABC
    /kgn2yS36DP50ZeyOJDF7bHce7ELNmuknFUnSVL/r/VUYZrMmc3RyuV+7sYg042n++tqPbm0XkHS
    4Fpy3mC78a3+I8weUuhUaVaYYTEhKXX9FMPRF/qKW7Avt8v04LAaS3iNnYUk+yAyDYqyE/WedB22
    /GjiYnjk7Z5TpevuV5g4N9BiY12RuJWDOELbYR+Y8SR+js5YUH95uy5Nd2qI7uF8RPh0liSc5Tsz
    hBy+vwhBcrUTx1JN55bFqDKVtRBcjHe/ynY32Cg0inooB+yqBRbRUXK2LAZKwb4WAyo7xY923b/h
    p5KcMjmrRGd0AlBTum8FaFEm6M746ZYjSeW/X3wpJCWXnNkwe4mNkVkZJ9TgCh2MlzhtaLMkaqlM
    MldmwG1+AAE3SGCuZU6rg0BbPXp3iSqLPEIIXAV+l1vY9sxGrITQ61PN3ZqDdx0nCQxUVbha0ke9
    fFv8ZS/XsletkXME/mFvkhhGd2eaow0+B+SSaPWsaVXC/HIlzShXAZfDHI2dy1kfKmPcdUeSlqlH
    0zL2+H/h98+5vn71PvU1UXuYEg57aqDqKEvRjyU7hXN66JRzBn+j4ppjK4fo7hQJNIerEMYEKC5B
    hDkjXVlNalTrcI0AEQEAAQ==
device-key-sha3-384: cP_JQakFySqvoRbC6RiTP4ik-YXrAK1xFX_V4qN4WWqdGC0X817QpOX6SJf77E_U
timestamp: 2017-10-11T07:23:40.387695Z
sign-key-sha3-384: wrfougkz3Huq2T_KklfnufCC0HzG7bJ9wP99GV0FF-D3QH3eJtuSRlQc2JhrAoh1

AcLBUgQAAQoABgUCWd3G/AAA3SYQAJC7A18fmY4NqxLziqoVW65IPNjegi7I5NqJJhPycpedHlbA
mpZR93a2op+I+ssCj65S0YKTJJHtmeEeVFP2ns9sK3aXJIWSGW3zvVkcuVJlylpW6zFlREjM986I
9XRZR/lZ/2bMZzut9ZsYxEfzvcGiMTrqqUB0VToVqemrV6rQgmZ0k0BAoTNh4EIpxY83ruBUQqZX
X/iJOUV2dVXAai4XQ6XZRQlGTtJgVMcpNbm2pPKa71eYstChtGUkKp4klBu8hOcgJVpcIj+bSHZi
3Zzkx5jMvc6R5qOumTB26Gono8jHcYmn5BHdqZr/7x6UxDXAQm1ac1XklvICK7d7pjI4G/53uROV
aQKwxUyvVgZNPI3/ieN1L7MriNVFjTfQm7cJq8zVYL26yR0r2B9UGLxP8rG3Tr2S+ySVadqH44yh
9r/P5ZOfezOTnTeBL9JAHJeXuQNMO2VZFJWhCKZyHCAtCytGwDjCMMTOxqdkySpsM0VsBz4TizEc
InRto3Z45nadl7RhIK85zF4zycTYXveQsCTZ0zkp5Jfgm90rJ1qM0aKfQiuSOdY14YQZqQG1Fc3w
iSriN2otHAUwCXgr7vg7wiHOZKHf/LzV0/110iIHGxFRpuitZB+oHXw4EBOPPnDwpiFjBsbgspl9
4VDviwTwZm6obnuDEAxynpe6lhNz`

const serialAssertInvalidBrand = `type: serial
authority-id: invalid
brand-id: invalid
model: generic-classic
serial: 54bdfe2e-607a-4d37-a469-68397070ec0c
device-key:
    AcbBTQRWhcGAARAAw0qcnXY26JyBvYpCigzIhWYTAnnHyjyB/MjA1lRVgYqeJUyrsUExj42q4ABC
    /kgn2yS36DP50ZeyOJDF7bHce7ELNmuknFUnSVL/r/VUYZrMmc3RyuV+7sYg042n++tqPbm0XkHS
    4Fpy3mC78a3+I8weUuhUaVaYYTEhKXX9FMPRF/qKW7Avt8v04LAaS3iNnYUk+yAyDYqyE/WedB22
    /GjiYnjk7Z5TpevuV5g4N9BiY12RuJWDOELbYR+Y8SR+js5YUH95uy5Nd2qI7uF8RPh0liSc5Tsz
    hBy+vwhBcrUTx1JN55bFqDKVtRBcjHe/ynY32Cg0inooB+yqBRbRUXK2LAZKwb4WAyo7xY923b/h
    p5KcMjmrRGd0AlBTum8FaFEm6M746ZYjSeW/X3wpJCWXnNkwe4mNkVkZJ9TgCh2MlzhtaLMkaqlM
    MldmwG1+AAE3SGCuZU6rg0BbPXp3iSqLPEIIXAV+l1vY9sxGrITQ61PN3ZqDdx0nCQxUVbha0ke9
    fFv8ZS/XsletkXME/mFvkhhGd2eaow0+B+SSaPWsaVXC/HIlzShXAZfDHI2dy1kfKmPcdUeSlqlH
    0zL2+H/h98+5vn71PvU1UXuYEg57aqDqKEvRjyU7hXN66JRzBn+j4ppjK4fo7hQJNIerEMYEKC5B
    hDkjXVlNalTrcI0AEQEAAQ==
device-key-sha3-384: cP_JQakFySqvoRbC6RiTP4ik-YXrAK1xFX_V4qN4WWqdGC0X817QpOX6SJf77E_U
timestamp: 2017-10-11T07:23:40.387695Z
sign-key-sha3-384: wrfougkz3Huq2T_KklfnufCC0HzG7bJ9wP99GV0FF-D3QH3eJtuSRlQc2JhrAoh1

AcLBUgQAAQoABgUCWd3G/AAA3SYQAJC7A18fmY4NqxLziqoVW65IPNjegi7I5NqJJhPycpedHlbA
mpZR93a2op+I+ssCj65S0YKTJJHtmeEeVFP2ns9sK3aXJIWSGW3zvVkcuVJlylpW6zFlREjM986I
9XRZR/lZ/2bMZzut9ZsYxEfzvcGiMTrqqUB0VToVqemrV6rQgmZ0k0BAoTNh4EIpxY83ruBUQqZX
X/iJOUV2dVXAai4XQ6XZRQlGTtJgVMcpNbm2pPKa71eYstChtGUkKp4klBu8hOcgJVpcIj+bSHZi
3Zzkx5jMvc6R5qOumTB26Gono8jHcYmn5BHdqZr/7x6UxDXAQm1ac1XklvICK7d7pjI4G/53uROV
aQKwxUyvVgZNPI3/ieN1L7MriNVFjTfQm7cJq8zVYL26yR0r2B9UGLxP8rG3Tr2S+ySVadqH44yh
9r/P5ZOfezOTnTeBL9JAHJeXuQNMO2VZFJWhCKZyHCAtCytGwDjCMMTOxqdkySpsM0VsBz4TizEc
InRto3Z45nadl7RhIK85zF4zycTYXveQsCTZ0zkp5Jfgm90rJ1qM0aKfQiuSOdY14YQZqQG1Fc3w
iSriN2otHAUwCXgr7vg7wiHOZKHf/LzV0/110iIHGxFRpuitZB+oHXw4EBOPPnDwpiFjBsbgspl9
4VDviwTwZm6obnuDEAxynpe6lhNz`

const serialAssertInvalid = `type: serial
authority-id: generic
brand-id: generic
model: invalid
serial: 54bdfe2e-607a-4d37-a469-68397070ec0c
device-key:
    AcbBTQRWhcGAARAAw0qcnXY26JyBvYpCigzIhWYTAnnHyjyB/MjA1lRVgYqeJUyrsUExj42q4ABC
    /kgn2yS36DP50ZeyOJDF7bHce7ELNmuknFUnSVL/r/VUYZrMmc3RyuV+7sYg042n++tqPbm0XkHS
    4Fpy3mC78a3+I8weUuhUaVaYYTEhKXX9FMPRF/qKW7Avt8v04LAaS3iNnYUk+yAyDYqyE/WedB22
    /GjiYnjk7Z5TpevuV5g4N9BiY12RuJWDOELbYR+Y8SR+js5YUH95uy5Nd2qI7uF8RPh0liSc5Tsz
    hBy+vwhBcrUTx1JN55bFqDKVtRBcjHe/ynY32Cg0inooB+yqBRbRUXK2LAZKwb4WAyo7xY923b/h
    p5KcMjmrRGd0AlBTum8FaFEm6M746ZYjSeW/X3wpJCWXnNkwe4mNkVkZJ9TgCh2MlzhtaLMkaqlM
    MldmwG1+AAE3SGCuZU6rg0BbPXp3iSqLPEIIXAV+l1vY9sxGrITQ61PN3ZqDdx0nCQxUVbha0ke9
    fFv8ZS/XsletkXME/mFvkhhGd2eaow0+B+SSaPWsaVXC/HIlzShXAZfDHI2dy1kfKmPcdUeSlqlH
    0zL2+H/h98+5vn71PvU1UXuYEg57aqDqKEvRjyU7hXN66JRzBn+j4ppjK4fo7hQJNIerEMYEKC5B
    hDkjXVlNalTrcI0AEQEAAQ==
device-key-sha3-384: cP_JQakFySqvoRbC6RiTP4ik-YXrAK1xFX_V4qN4WWqdGC0X817QpOX6SJf77E_U
timestamp: 2017-10-11T07:23:40.387695Z
sign-key-sha3-384: wrfougkz3Huq2T_KklfnufCC0HzG7bJ9wP99GV0FF-D3QH3eJtuSRlQc2JhrAoh1

AcLBUgQAAQoABgUCWd3G/AAA3SYQAJC7A18fmY4NqxLziqoVW65IPNjegi7I5NqJJhPycpedHlbA
mpZR93a2op+I+ssCj65S0YKTJJHtmeEeVFP2ns9sK3aXJIWSGW3zvVkcuVJlylpW6zFlREjM986I
9XRZR/lZ/2bMZzut9ZsYxEfzvcGiMTrqqUB0VToVqemrV6rQgmZ0k0BAoTNh4EIpxY83ruBUQqZX
X/iJOUV2dVXAai4XQ6XZRQlGTtJgVMcpNbm2pPKa71eYstChtGUkKp4klBu8hOcgJVpcIj+bSHZi
3Zzkx5jMvc6R5qOumTB26Gono8jHcYmn5BHdqZr/7x6UxDXAQm1ac1XklvICK7d7pjI4G/53uROV
aQKwxUyvVgZNPI3/ieN1L7MriNVFjTfQm7cJq8zVYL26yR0r2B9UGLxP8rG3Tr2S+ySVadqH44yh
9r/P5ZOfezOTnTeBL9JAHJeXuQNMO2VZFJWhCKZyHCAtCytGwDjCMMTOxqdkySpsM0VsBz4TizEc
InRto3Z45nadl7RhIK85zF4zycTYXveQsCTZ0zkp5Jfgm90rJ1qM0aKfQiuSOdY14YQZqQG1Fc3w
iSriN2otHAUwCXgr7vg7wiHOZKHf/LzV0/110iIHGxFRpuitZB+oHXw4EBOPPnDwpiFjBsbgspl9
4VDviwTwZm6obnuDEAxynpe6lhNz`

func parsePivotResponse(w *httptest.ResponseRecorder) (pivot.Response, error) {
	// Check the JSON response
	result := pivot.Response{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func (s *PivotSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)
}

func (s *PivotSuite) TestPivotModelHandler(c *check.C) {

	tests := []PivotTest{
		{"POST", "/v1/pivot", nil, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivot", []byte{}, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivot", []byte("invalid"), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivot", []byte(serialAssert), 200, jsonType, "ValidAPIKey", true},
		{"POST", "/v1/pivot", []byte(serialAssert), 400, jsonType, "InvalidAPIKey", false},
		{"POST", "/v1/pivot", []byte(serialAssertInvalid), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivot", []byte(serialAssertInvalidBrand), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivot", []byte(serialAssertNonReseller), 400, jsonType, "ValidAPIKey", false},
	}

	for _, t := range tests {
		w := sendSigningRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.APIKey, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := parsePivotResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)
	}

}

func (s *PivotSuite) TestPivotModelSerialAssertionHandler(c *check.C) {

	tests := []PivotTest{
		{"POST", "/v1/pivotmodel", nil, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte{}, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte("invalid"), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte(serialAssert), 200, asserts.MediaType, "ValidAPIKey", true},
		{"POST", "/v1/pivotmodel", []byte(serialAssert), 400, jsonType, "InvalidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte(serialAssertInvalid), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte(serialAssertInvalidBrand), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte(serialAssertNonReseller), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotserial", nil, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotserial", []byte{}, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotserial", []byte("invalid"), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotserial", []byte(serialAssert), 200, asserts.MediaType, "ValidAPIKey", true},
		{"POST", "/v1/pivotserial", []byte(serialAssert), 400, jsonType, "InvalidAPIKey", false},
		{"POST", "/v1/pivotserial", []byte(serialAssertInvalid), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotserial", []byte(serialAssertInvalidBrand), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotserial", []byte(serialAssertNonReseller), 400, jsonType, "ValidAPIKey", false},
	}

	for _, t := range tests {
		w := sendSigningRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.APIKey, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		if t.Type == jsonType {
			result, err := parsePivotResponse(w)
			c.Assert(err, check.IsNil)
			c.Assert(result.Success, check.Equals, t.Success)
		}
	}

}

func sendSigningRequest(method, url string, data io.Reader, apiKey string, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)
	r.Header.Set("api-key", apiKey)

	service.SigningRouter().ServeHTTP(w, r)

	return w
}
