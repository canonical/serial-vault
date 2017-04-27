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
import  AlertBox from './AlertBox'
import Accounts from '../models/accounts'
import Keypairs from '../models/keypairs'
import {T, parseResponse, formatError} from './Utils';

class AccountKeyForm extends Component {

    constructor(props) {
        super(props);

        this.state = {
            keypairs: props.keypairs || [],
            keypairId: 0,
            assertion: null,
            error: null
        }

        this.getKeypairs()
    }

    getKeypairs() {
        Keypairs.list().then((response) => {
            var data = JSON.parse(response.body);
            var message = "";
            if (!data.success) {
                message = data.message;
            }
            this.setState({keypairs: data.keypairs, message: message});
        });
    }

    handleChangeKeypair = (e) => {
        e.preventDefault();
        this.setState({keypairId: parseInt(e.target.value)})
    }

    handleFileUpload = (e) => {
        var reader = new FileReader();
        var file = e.target.files[0];

        reader.onload = (upload) => {
            this.setState({
                assertion: upload.target.result.split(',')[1],
            });
        }

        reader.readAsDataURL(file);
    }

    handleSaveClick = (e) => {
        e.preventDefault();

        Keypairs.assertion(this.state.keypairId, this.state.assertion).then((response) => {
            var data = parseResponse(response)
            if (!data.success) {
                this.setState({error: formatError(data)});
            } else {
                window.location = '/accounts';
            }
        });
    }

    render() {
        return (
            <div>

                <section className="row no-border">
                    <h2>{T('new-account-key-assertion')}</h2>
                    <div className="col-12">
                        <AlertBox message={this.state.error} />

                        <form>
                            <fieldset>
                                <label htmlFor="keypair">{T('private-key')}:
                                    <select value={this.state.keypairId} id="keypair" onChange={this.handleChangeKeypair}>
                                        <option></option>
                                        {this.state.keypairs.map((kpr) => {
                                            if (kpr.Active) {
                                                return <option key={kpr.ID} value={kpr.ID}>{kpr.AuthorityID}/{kpr.KeyID}</option>;
                                            } else {
                                                return <option key={kpr.ID} value={kpr.ID}>{kpr.AuthorityID}/{kpr.KeyID} ({T('inactive')})</option>;
                                            }
                                        })}
                                    </select>
                                </label>
                                <label htmlFor="key">{T('assertion')}:
                                    <input type="file" onChange={this.handleFileUpload} />
                                </label>
                            </fieldset>
                        </form>
                        <div>
                            <a href='/accounts' className="p-button--neutral">{T('cancel')}</a>
                            &nbsp;
                            <a href='/accounts' onClick={this.handleSaveClick} className="p-button--brand">{T('save')}</a>
                        </div>
                    </div>
                </section>
                <br />
            </div>
        );
    }

}

export default AccountKeyForm