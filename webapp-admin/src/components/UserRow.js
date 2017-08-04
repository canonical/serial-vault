/*
 * Copyright (C) 2017 Canonical Ltd
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
import {T, roleAsString} from './Utils'

class UserRow extends Component {
	renderActions() {
		if (this.props.user.ID !== this.props.confirmDelete) {
			return (
				<div>
					<a href={'/users/'.concat(this.props.user.ID, '/edit')} className="p-button--brand" title={T('edit-user')}><i className="fa fa-pencil"></i></a>
					&nbsp;
					<a href="" onClick={this.props.delete} data-key={this.props.user.ID} className="p-button--neutral" title={T('delete-user')}>
						<i className="fa fa-trash" data-key={this.props.user.ID}></i></a>
				</div>
			);
		} else {
			return (
				<DialogBox message={T('confirm-user-delete')} handleYesClick={this.props.deleteUser} handleCancelClick={this.props.cancelDelete} />
			);
		}
	}

	render() {
		return (
			<tr>
				<td>
					{this.renderActions()}
				</td>
				<td className="overflow" title={this.props.user.Username}>{this.props.user.Username}</td>
				<td className="overflow" title={this.props.user.Name}>{this.props.user.Name}</td>
				<td className="overflow" title={this.props.user.Email}>{this.props.user.Email}</td>
				<td className="overflow" title={roleAsString(this.props.user.Role)}>{roleAsString(this.props.user.Role)}</td>
			</tr>
		)
	}
}

export default UserRow;
