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

import React, {Component} from 'react';
import {T, isLoggedIn} from './Utils'


class Navigation extends Component {

    renderUser(token) {
        if (isLoggedIn(token)) {
            // The name is undefined if user authentication is off
            if (token.name) {
                return (
                    <li className="p-navigation__link"><a href="https://login.ubuntu.com/" className="p-link--external">{token.name}</a></li>
                )
            }
        } else {
            return (
            <li className="p-navigation__link"><a href="/login" className="p-link--external">{T('login')}</a></li>
            )
        }
    }

    renderUserLogout(token) {
        if (isLoggedIn(token)) {
            // The name is undefined if user authentication is off
            if (token.name) {
                return (
                    <li className="p-navigation__link"><a href="/logout">{T('logout')}</a></li>
                )
            }
        }
    }

    render() {

        var token = this.props.token

        return (
          <ul className="p-navigation__links u-float-right">
              {this.renderUser(token)}
              {this.renderUserLogout(token)}
          </ul>
        );
    }
}

export default Navigation;
