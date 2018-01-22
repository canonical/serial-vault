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


const sections = ['signing-keys', 'models', 'keypairs', 'accounts', 'signinglog', 'systemuser', 'users', 'notfound']


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

export function subSectionIdFromPath(path, section) {
    const parts = path.split('/').slice(1)
    console.log('---parts', parts)
    return (parts[0] === section && parts[2]) || ''
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

export function getAuthToken(callback) {

    // Get a fresh token and return it to the callback
    // The token will be passed to the views
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
    return isUser(Role.Standard, token)
}

export function isUserAdmin(token) {
    return isUser(Role.Admin, token)
}

export function isUserSuperuser(token) {
    return isUser(Role.Superuser, token)
}

export function roleAsString(role) {
    var str
    switch (role) {
        case Role.Standard:
            str = "Standard"	
            break;
        case Role.Admin:
            str = "Admin"
            break;
        case Role.Superuser:
            str = "Superuser"
            break
        default:
            str= "invalid role"
            break;
    }
    return str
}

function isUser(role, token) {
    if (!token) return false
    if (!token.role) return false

    return (token.role >= role)
}