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

import React, {Component} from 'react'
import Keypairs from '../models/keypairs'
import AlertBox from './AlertBox'
import {T, isUserAdmin} from './Utils';

class KeypairGenerate extends Component {

    constructor(props) {
        super(props);

        this.state = {
            authorityId: null, 
            keyName: null, 
            error: this.props.error,
        };
    }

    handleChangeAuthorityId = (e) => {
        this.setState({authorityId: e.target.value});
    }

    handleChangeKeyName = (e) => {
        this.setState({keyName: e.target.value});
    }

    handleSaveClick = (e) => {
        var self = this;
        e.preventDefault();

        Keypairs.generate(this.state.authorityId, this.state.keyName).then(function(response) {
            var data = JSON.parse(response.body);
            if ((response.statusCode >= 300) || (!data.success)) {
        self.setState({error: self.formatError(data)});
      } else {
        window.location = '/models';
      }
        });
    }

    formatError = (data) => {
        var message = T(data.error_code);
        if (data.error_subcode) {
            message += ': ' + T(data.error_subcode);
        } else if (data.message) {
            message += ': ' + data.message;
        }
        return message;
    }

    render() {

        if (!isUserAdmin(this.props.token)) {
            return (
                <div className="row">
                <AlertBox message={T('error-no-permissions')} />
                </div>
            )
        }

        return (
            <div>
                <section className="row no-border">
                    <h2>{T('generate-signing-key')}</h2>
                    <div className="col-12">
                        <AlertBox message={this.state.error} />

                        <form>
                            <fieldset>
                                <label htmlFor="authority-id">{T('authority-id')}:
                                    <input type="text" id="authority-id" onChange={this.handleChangeAuthorityId} placeholder={T('authority-id-description')} />
                                </label>
                                <label htmlFor="key-name">{T('key-name')}:
                                    <input type="text" id="key-name" onChange={this.handleChangeKeyName} placeholder={T('key-name-description')} />
                                </label>

                            </fieldset>
                        </form>
                        <div>
                            <a href='/models' className="p-button--neutral">{T('cancel')}</a>
                            &nbsp;
                            <a href='/models' onClick={this.handleSaveClick} className="p-button--brand">{T('generate')}</a>
                        </div>
                    </div>
                </section>
                <br />
            </div>
        );
    }
}

export default KeypairGenerate;
