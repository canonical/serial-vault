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
import moment from 'moment';


class SigningLogRow extends Component {
	render() {
		return (
			<tr>
				<td className="wrap">{this.props.log.make}</td>
				<td className="wrap">{this.props.log.model}</td>
				<td className="wrap">{this.props.log.serialnumber}</td>
				<td>{this.props.log.revision}</td>
				<td className="overflow" title={this.props.log.fingerprint}>{this.props.log.fingerprint}</td>
				<td className="wrap">{moment(this.props.log.created).format("YYYY-MM-DD HH:mm")}</td>
			</tr>
		)
	}
}

export default SigningLogRow;
