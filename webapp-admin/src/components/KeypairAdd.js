/*
 * Copyright (C) 2016-2018 Canonical Ltd
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
import Accounts from '../models/accounts';
import AlertBox from './AlertBox'
import {T, isUserAdmin} from './Utils';

class KeypairAdd extends Component {

    constructor(props) {
        super(props);

        this.state = {
            accounts: this.props.accounts || [],
            authorityId: null,
            name: null,
            key: null,
            error: this.props.error,
        };

        this.getAccounts()
    }

    getAccounts() {
        Accounts.list().then((response) => {
            var authority = null;
            var data = JSON.parse(response.body);
            var message = "";

            if (!data.success) {
                message = data.message;
            } else {
                // Select the first account
                if (data.accounts.length > 0) {
                    authority = data.accounts[0].AuthorityID;
                }
            }
            this.setState({accounts: data.accounts, authorityId: authority, error: message});
        });
    }

    handleChangeAuthorityId = (e) => {
        this.setState({authorityId: e.target.value});
    }

    handleChangeKeyName = (e) => {
        this.setState({name: e.target.value});
    }

    handleChangeKey = (e) => {
        this.setState({key: e.target.value});
    }

    handleFileUpload = (e) => {
        var self = this;
        var reader = new FileReader();
        var file = e.target.files[0];

        reader.onload = function(upload) {
        self.setState({
            key: upload.target.result.split(',')[1],
        });
        }

        reader.readAsDataURL(file);
    }

    handleSaveClick = (e) => {
        e.preventDefault();

        Keypairs.create(this.state.authorityId, this.state.key, this.state.name).then((response) => {
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

        return (
            <div>
                <section className="row no-border">
                    <h2>{T('new-signing-key')}</h2>
                    <div className="col-12">
                        <AlertBox message={this.state.error} />

                        <form>
                            <fieldset>
                                <label htmlFor="name">{T('key-name')}:
                                    <input type="text" id="name" onChange={this.handleChangeKeyName} value={this.state.name} placeholder={T('key-name-description')} />
                                </label>
                                <label htmlFor="authority-id">{T('authority-id')}:
                                    <select value={this.state.authorityId} id="authority-id" onChange={this.handleChangeAuthorityId}>
                                        {this.state.accounts.map(function(a) {
                                            return <option key={a.AuthorityID} value={a.AuthorityID}>{a.AuthorityID}</option>;
                                        })}
                                    </select>
                                </label>
                                <label htmlFor="key">{T('signing-key')}:
                                    <textarea onChange={this.handleChangeKey} defaultValue={this.state.key} id="key"
                                            placeholder={T('new-signing-key-description')}>
                                    </textarea>
                                    <input type="file" onChange={this.handleFileUpload} />
                                </label>
                            </fieldset>
                        </form>
                        <div>
                            <a href='/signing-keys' className="p-button--neutral">{T('cancel')}</a>
                            &nbsp;
                            <a href='/signing-keys' onClick={this.handleSaveClick} className="p-button--brand">{T('save')}</a>
                        </div>
                    </div>
                </section>
                <br />
            </div>
        );
    }
}

export default KeypairAdd;
