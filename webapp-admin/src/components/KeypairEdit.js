/*
 * Copyright (C) 2018 Canonical Ltd
 * License granted by Canonical Limited
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

class KeypairEdit extends Component {
    constructor(props) {
        super(props);

        this.state = {
            keypair: {},
            error: null,
        };

        this.getKeypair(this.props.id)
    }

    getKeypair(keyId) {
        Keypairs.get(keyId).then((response) => {
            var data = JSON.parse(response.body);

            if (response.statusCode >= 300) {
                this.setState({error: this.formatError(data), hideForm: true});
            } else {
                this.setState({keypair: data.keypair, hideForm: false});
            }
        });
    }

    handleChangeKeyName = (e) => {
        var k = this.state.keypair
        k.KeyName = e.target.value
        this.setState({keypair: k});
    }

    handleSaveClick = (e) => {
        e.preventDefault();

        Keypairs.update(this.state.keypair).then((response) => {
            var data = JSON.parse(response.body);
            if ((response.statusCode >= 300) || (!data.success)) {
                this.setState({error: this.formatError(data)});
            } else {
                window.location = '/signing-keys';
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

        if (this.state.hideForm) {
            return (
                <div className="row">
                <AlertBox message={this.state.error} />
                </div>
            )
        }

        return (
            <div className="row">
                <section className="row">
                      <h2>{T('edit-signing-key')}</h2>

                        <AlertBox message={this.state.error} />
                        <form>
                            <fieldset>
                                <label htmlFor="name">{T('key-name')}:
                                    <input type="text" id="name" onChange={this.handleChangeKeyName} value={this.state.keypair.KeyName} placeholder={T('key-name-description')} />
                                </label>

                                <label htmlFor="authority-id">{T('authority-id')}:
                                    <input type="text" id="authority-id" placeholder={T('authority-id-description')}
                                        value={this.state.keypair.AuthorityID} disabled />
                                </label>
                            </fieldset>
                        </form>
                        <div>
                            <a href='/signing-keys' className="p-button--neutral">{T('cancel')}</a>
                            &nbsp;
                            <a href='/signing-keys' onClick={this.handleSaveClick} className="p-button--brand">{T('save')}</a>
                        </div>
                </section>
            </div>
        )

    }

}

export default KeypairEdit;
