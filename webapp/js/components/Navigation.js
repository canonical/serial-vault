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
var injectIntl = require('react-intl').injectIntl;

var Navigation = React.createClass({
    render: function() {
      var M = this.props.intl.formatMessage;

			var activeHome = '';
			var activeModels = '';
			if (this.props.active === 'home') {
				activeHome = 'active';
			}
			if (this.props.active === 'models') {
				activeModels = 'active';
			}

      return (

        <nav id="navigation" role="navigation" className="nav-secondary clearfix open">
          <ul className="second-level-nav">
            <li><a className={activeHome} href="/">{M({id:'home'})}</a></li>
            <li><a className={activeModels} href="/models">{M({id:'models'})}</a></li>
          </ul>
        </nav>
      );
    }
});

module.exports = injectIntl(Navigation);
