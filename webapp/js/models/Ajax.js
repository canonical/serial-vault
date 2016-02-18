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
'use strict'
var request =require('then-request');
var API_VERSION = '/1.0/';

var Ajax = {
	get: function(url) {
			return request('GET', API_VERSION + url, {
					headers: {}
			});
	},

	post: function(url, data) {
			return request('POST', '/' + API_VERSION + url, {
					headers: {},
					json: data
			});
	},

	put: function(url, data) {
			return request('PUT', '/' + API_VERSION + url, {
					headers: {},
					json: data
			});
	},

	delete: function(url, data) {
			return request('DELETE', '/' + API_VERSION + url, {
					headers: {},
					json: data
			});
	}
}

module.exports = Ajax;
