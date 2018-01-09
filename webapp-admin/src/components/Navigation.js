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
import {T} from './Utils'
import {Role} from './Constants'

const linksSuperuser = ['signing-keys', 'models', 'accounts', 'systemuser', 'signinglog', "users"];
const linksAdmin = ['signing-keys', 'models', 'accounts', 'systemuser', 'signinglog'];
const linksStandard = ['systemuser'];


class Navigation extends Component {

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
