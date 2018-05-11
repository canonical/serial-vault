/*
 * Copyright (C) 2018 Canonical Ltd
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
import {T} from './Utils'

class NavigationSubmenu extends Component {

    render() {
        var index = 0;
        return (
            <nav className="p-tabs">
            <ul className="p-tabs__list" role="tablist">
                {
                    this.props.items.map((item) => {
                        index += 1
                        return (
                            <li className="p-tabs__item" role="presentation">
                                <a href={'/' + item} className="p-tabs__link" tabIndex={index} role="tab"
                                    aria-controls={item} aria-selected={item===this.props.selected}>
                                    {T(item)}
                                </a>
                            </li>
                        )
                    })
                }
            </ul>
            </nav>
        )
    }
}

export default NavigationSubmenu;
