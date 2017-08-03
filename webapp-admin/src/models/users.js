/*
 * Copyright (C) 2017 Canonical Ltd
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
import Ajax from './Ajax';

var Users = {
	url: 'users',

	list: function () {
		return Ajax.get(this.url);
	},

	get: function(userId) {
		return Ajax.get(this.url + '/' + userId);
	},

	getotheraccounts: function(userId) {
		return Ajax.get(this.url + '/' + userId + '/otheraccounts');
	},

	update:  function(user) {
		return Ajax.put(this.url + '/' + user.ID, user);
	},

	delete:  function(user) {
		return Ajax.delete(this.url + '/' + user.ID, {});
	},

	create:  function(user) {
		return Ajax.post(this.url, user);
	}

}

export default Users;
