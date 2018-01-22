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
import Ajax from './Ajax'

var Account = {
    url: 'accounts',

    list: function () {
        return Ajax.get(this.url);
    },

    get: function (id) {
        return Ajax.get(this.url + '/' + id);
    },

    create:  function(account) {
        return Ajax.post(this.url, account);
    },

    update:  function(account) {
        return Ajax.put(this.url + '/' + account.ID, account);
    },

    upload:  function(assertion) {
        return Ajax.post(this.url + '/upload', {assertion: assertion});
    }
}

export default Account
