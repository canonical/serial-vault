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

var DialogBox = React.createClass({
	render: function() {

		if (this.props.message) {
			return (
				<div className="p-card warning">
					<p>{this.props.message}</p>
					<div>
						<a href="" onClick={this.props.handleYesClick} className="p-button--brand">
							{T('yes')}
						</a>
						&nbsp;
						<a href="" onClick={this.props.handleCancelClick} className="p-button--neutral">
							{T('cancel')}
						</a>
					</div>
				</div>
			);
		} else {
			return <span />;
		}
	}
});

module.exports = DialogBox;
