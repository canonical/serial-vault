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


var Navigation = React.createClass({
    render: function() {

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
					<li className={activeHome}><a href="/">{T('home')}</a></li>
					<li className={activeModels}><a href="/models">{T('models')}</a></li>
					<li className={activeAccounts}><a href="/accounts">{T('accounts')}</a></li>
					<li className={activeSigningLog}><a href="/signinglog">{T('signinglog')}</a></li>
				</ul>
      );
    }
});

module.exports = Navigation;
