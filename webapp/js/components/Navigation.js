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

var React = require('react');
import {T} from './Utils';
import {getAuthToken, isUserAdmin, isLoggedIn} from './Utils'


var Navigation = React.createClass({
    getInitialState: function() {
        return {token: {}}
    },

    componentDidMount: function() {
        getAuthToken(this.setAuthToken)
    },

    setAuthToken: function(token) {
        this.setState({token: token})
    },

    renderLink: function(token, active, link, label) {
        if (isUserAdmin(token)) {
            return <li className={active}><a href={link}>{T(label)}</a></li>
        }
    },

    renderUser: function(token) {
        if (isLoggedIn(token)) {
            return (
            <li className="p-navigation__link">
                <a href="https://login.ubuntu.com/" className="p-link--external">{token.name}</a>
            </li>
            )
        } else {
            return (
            <li className="p-navigation__link"><a href="/login" className="p-link--external">Login</a></li>
            )
        }
    },

    render: function() {

        var token = this.state.token
        console.log('Nav', token)

        var activeHome = 'p-navigation__link';
        var activeModels = 'p-navigation__link';
        var activeKeys = 'p-navigation__link';
        var activeSigningLog = 'p-navigation__link';
        var activeAccounts = 'p-navigation__link';
        if (this.props.active === 'home') {
            activeHome = 'p-navigation__link active';
        }
        if (this.props.active === 'models') {
            activeModels = 'p-navigation__link active';
        }
        if (this.props.active === 'keys') {
            activeKeys = 'p-navigation__link active';
        }
        if (this.props.active === 'accounts') {
            activeAccounts = 'p-navigation__link active';
        }
        if (this.props.active === 'signinglog') {
            activeSigningLog = 'p-navigation__link active';
        }

        return (
          <ul className="p-navigation__links">
              {this.renderLink(token, activeHome, '/', 'home')}
              {this.renderLink(token, activeModels, '/models', 'models')}
              {this.renderLink(token, activeAccounts, '/accounts', 'accounts')}
              {this.renderLink(token, activeSigningLog, '/signinglog', 'signinglog')}
              {this.renderUser(token)}
          </ul>
        );
    }
});

module.exports = Navigation;
