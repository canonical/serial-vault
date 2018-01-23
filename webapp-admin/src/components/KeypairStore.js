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
import Keypairs from '../models/keypairs';
import AlertBox from './AlertBox';
import {T, formatError, isUserAdmin} from './Utils'


class KeypairStore extends Component {
    constructor(props) {
        super(props)
        this.state = {
            keypairs: [],
            keypair: {}
        }

        this.getKeypairs()
    }

    handleChangeKey = (e) => {
        var id = parseInt(e.target.value, 10);

        var keys = this.state.keypairs.filter((k) => {
            return k.ID === id
        })
        if (keys.length > 0) {
            var k = {'authority-id': keys[0].AuthorityID, 'key-name': keys[0].KeyName}
            this.setState({keypair: k})
        }
    }

    handleChangeName = (e) => {
        var keypair = this.state.keypair
        keypair.KeyName = e.target.value
        this.setState({keypair: keypair})
    }

    handleChangeEmail = (e) => {
        var keypair = this.state.keypair
        keypair.email = e.target.value
        this.setState({keypair: keypair})
    }

    handleChangePassword = (e) => {
        var keypair = this.state.keypair
        keypair.password = e.target.value
        this.setState({keypair: keypair})
    }

    handleChangeOTP = (e) => {
        var keypair = this.state.keypair
        keypair.otp = e.target.value
        this.setState({keypair: keypair})
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

    handleSaveClick = (e) => {
        e.preventDefault()

        Keypairs.register(this.state.keypair).then((response) => {
            var data = JSON.parse(response.body);
            if ((response.statusCode >= 300) || (!data.success)) {
                this.setState({error: formatError(data)});
            } else {
                window.location = '/signing-keys';
            }
        });
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
                    <h2>{T('register-signing-key')}</h2>
                    <div className="col-12">
                        <AlertBox message={this.state.error} />

                        <form>
                            <fieldset>

                                <label htmlFor="keypair">{T('signing-key')}:
                                    <select value={this.state.keypair.ID} id="keypair" onChange={this.handleChangeKey}>
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

                                <label htmlFor="key-name">{T('key-name')}:
                                    <input type="text" id="key-name" onChange={this.handleChangeName} placeholder={T('key-name-description')} value={this.state.keypair['key-name']} />
                                </label>

                                <h3>{T('store-credentials')}</h3>
                                <label htmlFor="email">{T('email')}:
                                    <input type="email" id="email" onChange={this.handleChangeEmail} placeholder={T('email-description')} value={this.state.keypair.email} />
                                </label>
                                <label htmlFor="password">{T('password')}:
                                    <input type="password" id="password" onChange={this.handleChangePassword} placeholder={T('password-description')} value={this.state.keypair.password} />
                                </label>
                                <label htmlFor="otp">{T('otp')}:
                                    <input type="text" id="otp" onChange={this.handleChangeOTP} placeholder={T('otp-description')} value={this.state.keypair.otp} />
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

export default KeypairStore;
