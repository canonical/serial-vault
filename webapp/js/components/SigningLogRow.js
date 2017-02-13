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
var moment = require('moment');
var DialogBox = require('./DialogBox');
import {T} from './Utils';


var SigningLogRow = React.createClass({
	renderActions: function(msg) {

		if (this.props.log.id !== this.props.confirmDelete) {
			return (
				<div>
					<a href="" onClick={this.props.delete} data-key={this.props.log.id} className="button--secondary small" title={msg.delete}>
						<i className="fa fa-trash" data-key={this.props.log.id}></i></a>
				</div>
			);
		} else {
			return (
				<DialogBox message={msg.confirmDelete} handleYesClick={this.props.deleteLog} handleCancelClick={this.props.cancelDelete} />
			);
		}
	},

	render: function() {
		const msg = {
			delete: T('delete-log'),
			confirmDelete: T('confirm-log-delete'),
		}

		return (
			<tr>
				<td>
					{this.renderActions(msg)}
				</td>
				<td className="wrap">{this.props.log.make}</td>
				<td className="wrap">{this.props.log.model}</td>
				<td className="wrap">{this.props.log.serialnumber}</td>
				<td>{this.props.log.revision}</td>
				<td className="overflow" title={this.props.log.fingerprint}>{this.props.log.fingerprint}</td>
				<td className="wrap">{moment(this.props.log.created).format("YYYY-MM-DD HH:mm")}</td>
			</tr>
		)
	}
});

module.exports = SigningLogRow;
