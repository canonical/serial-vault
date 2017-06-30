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
import request from 'then-request'


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
