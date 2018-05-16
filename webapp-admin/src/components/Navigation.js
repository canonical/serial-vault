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
import {Role} from './Constants'

const linksSuperuser = ['accounts', 'signing-keys', 'models', 'signinglog', "users"];
const linksAdmin = ['accounts', 'signing-keys', 'models', 'signinglog'];
const linksStandard = ['systemuser'];


class Navigation extends Component {

    constructor(props) {
        super(props);

        this.state = {
            accountMenu: false
        }
    }

    handleAccountChange = (e) => {
        e.preventDefault()

        // Get the account
        var accountId = parseInt(e.target.getAttribute('data-key'), 10)
        var account = this.props.accounts.filter(a => {
            return a.ID === accountId
        })[0]

        this.handleAccountMenu()
        this.props.onAccountChange(account)
    }

    handleAccountMenu = (e) => {
        if (e) {
         e.preventDefault()
        }
        this.setState({accountMenu: !this.state.accountMenu})
    }

    renderAccounts(token) {
        if (!isLoggedIn(token)) {
            return <span />
        }

        if (this.props.accounts.length === 0) {
            return <span />
        }

        var name = this.props.selectedAccount.AuthorityID
        if (name.length > 10) {
            name = name.slice(0, 10) + '...'
        }

        return (
            <li className="p-navigation__link account-menu">
                <a href="/" onClick={this.handleAccountMenu} className="p-contextual-menu__toggle" aria-controls="#account-menu" aria-expanded="false" aria-haspopup="true">
                {name} <i className="fa fa-caret-down"></i>
                
                {this.state.accountMenu ?
                <span className="p-contextual-menu__dropdown" id="account-menu" aria-hidden="false" aria-label="submenu">
                    <span className="p-contextual-menu__group">
                    {this.props.accounts.map(a => {
                    return (
                        <a key={a.ID} data-key={a.ID} href="/" onClick={this.handleAccountChange} className="p-contextual-menu__link">{a.AuthorityID}</a>
                    )
                    })}
                    </span>
                </span>
                : ''
                }
                </a>
            </li>
        )

    }

    render() {

        var token = this.props.token
        var links;

        switch(token.role) {
            case Role.Admin:
                links = linksAdmin;
                break;
            case Role.Superuser:
                links = linksSuperuser;
                break;
            case Role.Standard:
                links = linksStandard
                break
            default:
                links = []
        }

        return (
          <ul className="p-navigation__links">
              {this.renderAccounts(token)}
              {links.map((l) => {
                  var active = '';
                  if (this.props.active === l) {
                      active = ' active'
                  }
                  var link = '/' + l;
                  return (
                    <li key={l} className={'p-navigation__link' + active}><a href={link}>{T(l)}</a></li>
                  )
              })}
          </ul>
        );
    }
}

export default Navigation;
