/*
 * Copyright (C) 2017-2018 Canonical Ltd
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
import React, { Component } from 'react'


class SigningLogFilter extends Component {

    render() {
        if (!this.props.expanded) {
            return (
                <div className="filter-section">
                <h5 className='collapsed' onClick={this.props.expansionClick.bind(this, this.props.keyName)}>{this.props.name}</h5>
                </div>
            );
        }

        return (
            <div className="filter-section">
            <h5 onClick={this.props.expansionClick.bind(this, this.props.keyName)} className="expanded">{this.props.name}</h5>
            <ul>
                {this.props.items.map(function(item, i) {
                    return (
                        <li key={i} onClick={this.props.handleItemClick.bind(this, i, this.props.keyName)}>
                            <span className={item.selected ? "checked overflow" : "unchecked overflow"}>{item.name}</span>
                        </li>
                    );
                }, this)}
            </ul>
            </div>
        );
    }
}

export default SigningLogFilter