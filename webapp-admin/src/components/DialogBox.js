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
import {T} from './Utils';

class DialogBox extends Component {

    render() {
        var classYes = 'p-button--brand'
        var classNo = 'p-button--neutral'
        if (this.props.small) {
            classYes = 'p-button--brand small'
            classNo = 'p-button--neutral small'
        }

        if (this.props.message) {
            return (
                <div className="p-card warning">
                    <p>{this.props.message}</p>
                    <div>
                        <a href="" onClick={this.props.handleCancelClick} className={classNo}>
                            {T('no')}
                        </a>
                    </div>
                    <div>
                        <a href="" onClick={this.props.handleYesClick} className={classYes}>
                            {T('yes')}
                        </a>
                    </div>
                </div>
            );
        } else {
            return <span />;
        }
    }
}

export default DialogBox;
