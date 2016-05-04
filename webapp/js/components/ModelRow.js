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

var ModelRow = React.createClass({
	renderActions: function(M) {
		if (this.props.model.id !== this.props.confirmDelete) {
			return (
				<div>
					<a href={'/models/'.concat(this.props.model.id, '/edit')} className="button--primary" title={M({id: 'edit-model'})}><i className="fa fa-pencil"></i></a>
					&nbsp;
					<a href="" onClick={this.props.delete} data-key={this.props.model.id} className="button--secondary" title={M({id: 'delete-model'})}>
						<i className="fa fa-trash" data-key={this.props.model.id}></i></a>
				</div>
			);
		} else {
			return (
				<div>
					<em>{M({id: 'confirm-model-delete'})}</em><br />
					<button onClick={this.props.deleteModel} className="button--primary small">{M({id: 'yes'})}</button>&nbsp;
					<button onClick={this.props.cancelDelete} className="button--secondary small">{M({id: 'cancel'})}</button>
				</div>
			);
		}
	},

	render: function() {
		var M = this.props.intl.formatMessage;
		return (
			<tr>
			  <td>
					{this.renderActions(M)}
				</td>
				<td>{this.props.model['brand-id']}</td>
				<td>{this.props.model.model}</td>
				<td>{this.props.model.revision}</td>
				<td>{this.props.model['authority-id']}/{this.props.model['key-id']}</td>
			</tr>
		)
	}
});

module.exports = injectIntl(ModelRow);
