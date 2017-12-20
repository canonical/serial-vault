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
import DialogBox from './DialogBox';
import {T} from './Utils'

class ModelRow extends Component {
	renderActions() {
		if (this.props.model.id !== this.props.confirmDelete) {
			return (
				<div>
					<a href={'/models/'.concat(this.props.model.id, '/edit')} className="p-button--brand small" title={T('edit-model')}><i className="fa fa-pencil"></i></a>
					&nbsp;
					<a href="" onClick={this.props.delete} data-key={this.props.model.id} className="p-button--neutral small" title={T('delete-model')}>
						<i className="fa fa-trash" data-key={this.props.model.id}></i></a>
				</div>
			);
		} else {
			return (
				<DialogBox message={T('confirm-model-delete')} handleYesClick={this.props.deleteModel} handleCancelClick={this.props.cancelDelete} />
			);
		}
	}

	render() {
		var fingerprint = this.props.model['authority-id'] + '/' + this.props.model['key-id'];
		var fingerprintUser = this.props.model['authority-id-user'] + '/' + this.props.model['key-id-user'];
		return (
			<tr>
				<td>
					{this.renderActions()}
				</td>
				<td className="overflow" title={this.props.model['brand-id']}>{this.props.model['brand-id']}</td>
				<td>{this.props.model.model}</td>
				<td className="overflow" title={fingerprint} >{fingerprint}</td>
				<td className="overflow" title={fingerprintUser} >{fingerprintUser}</td>
				<td>{this.props.model['key-active'] && this.props.model['key-active-user'] ? <i className="fa fa-check"></i> :  <i className="fa fa-times"></i>}</td>
			</tr>
		)
	}
}

export default ModelRow;
