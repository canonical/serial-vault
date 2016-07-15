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

var DialogBox = React.createClass({
	render: function() {
		var M = this.props.intl.formatMessage;

		if (this.props.message) {
			return (
				<div className="box warning">
					<p>{this.props.message}</p>
					<div>
						<a href="" onClick={this.props.handleYesClick} className="button--primary small">{M({id: "yes"})}</a>
						&nbsp;
						<a href="" onClick={this.props.handleCancelClick} className="button--secondary small">{M({id: "cancel"})}</a>
					</div>
				</div>
			);
		} else {
			return <span />;
		}
	}
});

module.exports = injectIntl(DialogBox);
