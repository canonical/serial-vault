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
import Messages from './messages'
import jwtDecode from 'jwt-decode'
import Ajax from '../models/Ajax'
import {Role} from './Constants'


const sections = ['models', 'keypairs', 'accounts', 'signinglog']


export function sectionFromPath(path) {
  return path === '/' ? 'home' : (
    sections.find(section => (
      path.startsWith(`/${section}`)
    )) || ''
  )
}

export function sectionIdFromPath(path, section) {
  const parts = path.split('/').slice(1)
  return (parts[0] === section && parts[1]) || ''
}

export function T(message) {
    const lang = window.AppState.getLocale()
    const msg = Messages[lang][message] || message;
    return msg
}

export function parseResponse(response) {
    // Check the content-type is JSON
    if (!response.headers['content-type'].includes('json')) {
        // Non-json body
        return {
            error_code: response.statusCode,
            message: response.body
        }
    }

    // Parse the body as JSON
    return JSON.parse(response.body);
}

export function formatError (data) {
    var message = T(data.error_code);
    if (data.error_subcode) {
        message += ': ' + T(data.error_subcode);
    } else if (data.message) {
        message += ': ' + data.message;
    }
    return message;
}

var getQueryString = function ( field, url ) {
    var href = url ? url : window.location.href;
    var reg = new RegExp( '[?&]' + field + '=([^&#]*)', 'i' );
    var string = reg.exec(href);
    return string ? string[1] : null;
};

export function getAuthToken(callback) {

    if (localStorage.getItem('token')) {
        var t = JSON.parse(localStorage.getItem('token'))
        var utcTimestamp = Math.floor((new Date()).getTime() / 1000)
        if (t.exp > utcTimestamp) {
            // Use the token from local storage
            callback(t)
            return
        }
    }

    // Get a fresh token and store it in local storage
    Ajax.getAuthToken().then((resp) => {

        // If user authentication is off, set up a default token
        var data = JSON.parse(resp.body)
        if (!data.enableUserAuth) {
            callback({role: Role.Admin})
            return
        }

        var jwt = resp.headers.authorization

        if (!jwt) {
            callback({})
            return
        }
        var token = jwtDecode(jwt)

        if (!token) {
            callback({})
            return
        }

        localStorage.setItem('token', JSON.stringify(token))
        callback(token)
    })

}

export function authToken() {
    if (localStorage.getItem('token')) {
        var t = JSON.parse(localStorage.getItem('token'))
        return t
    } else {
        return {}
    }
}

export function isLoggedIn(token) {
    return isUserStandard(token)
}

export function isUserStandard(token) {
    if (!token) return false
    if (!token.role) return false

    if (token.role >= Role.Standard) {
        return true
    } else {
        return false
    }
}

export function isUserAdmin(token) {
    if (!token) return false
    if (!token.role) return false

    if (token.role >= Role.Admin) {
        return true
    } else {
        return false
    }
}

export function isUserSuperuser(token) {
    if (!token) return false
    if (!token.role) return false

    if (token.role >= Role.Superuser) {
        return true
    } else {
        return false
    }
}