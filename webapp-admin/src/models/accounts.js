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

    list() {
        return Ajax.get(this.url);
    },

    get(id) {
        return Ajax.get(this.url + '/' + id);
    },

    create(account) {
        return Ajax.post(this.url, account);
    },

    update(account) {
        return Ajax.put(this.url + '/' + account.ID, account);
    },

    upload(assertion) {
        return Ajax.post(this.url + '/upload', {assertion: assertion});
    },

    stores(id) {
        return Ajax.get(this.url + '/' + id + '/stores');
    },

    storeNew(accountID, store) {
        store.accountID = parseInt(accountID, 10)
        return Ajax.post(this.url + '/stores', store);
    },

    storeUpdate(store) {
        return Ajax.put(this.url + '/stores/' + store.id, store);
    },

    storeDelete(store) {
        return Ajax.delete(this.url + '/stores/' + store.id, {});
    }
}

export default Account
