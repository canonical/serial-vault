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
var request =require('then-request');
import axios from 'axios'

var  BASE_URL = '/v1/'

if (location.port === '3000') {
	// We're in dev mode so use the localhost:8082 for the backend
	BASE_URL = 'http://localhost:8082/v1/'
}

const config = {baseURL: BASE_URL,
				xsrfHeaderName: 'X-CSRF-Token',
				xsrfCookieName: 'XSRF-TOKEN',
			}

var Ajax = {

	getToken: function() {
		return axios.get('token', config)
	},

	get: function(url, qs) {
		return axios.get(url, config)
	},

	post: function(url, data) {
		// Get an updated CSRF token before a POST
		return this.getToken().then((response) => {
			// Set the CSRF token in the header
			return axios.post(BASE_URL + url, data,{
				headers: {
					'X-CSRF-Token': response.headers['x-csrf-token'],
				},
			});
		})

	},

	put: function(url, data, token) {
			return request('PUT', BASE_URL + url, {
					headers: {
						'X-CSRF-Token': token
					},
					json: data
			});
	},

	delete: function(url, data, token) {
			return request('DELETE', BASE_URL + url, {
					headers: {
						'X-CSRF-Token': token
					},
					json: data
			});
	}
}

module.exports = Ajax;
