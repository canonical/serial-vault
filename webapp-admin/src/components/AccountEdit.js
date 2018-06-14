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
import {T, parseResponse, formatError, isUserAdmin} from './Utils';

class AccountEdit extends Component {

    constructor(props) {
        super(props);

        this.state = {
            account: {},
            error: null,
        }

        if (this.props.id) {
            this.getAccount(this.props.id)
        }
    }

    getAccount(accountId) {
        Accounts.get(accountId).then((response) => {
            var data = JSON.parse(response.body);

            if (response.statusCode >= 300) {
                this.setState({error: this.formatError(data), hideForm: true});
            } else {
                this.setState({account: data.account, hideForm: false});
            }
        });
    }

    handleChangeAccount = (e) => {
        var a = this.state.account
        a.AuthorityID = e.target.value
        this.setState({account: a})
    }

    handleChangeReseller = (e) => {
        var a = this.state.account;
        a.ResellerAPI = e.target.checked;
        this.setState({account: a});
    }

    handleSaveClick = (e) => {
        e.preventDefault();

        if (this.state.account.ID) {
            Accounts.update(this.state.account).then((response) => {
                var data = parseResponse(response)
                if (!data.success) {
                    this.setState({error: formatError(data)});
                } else {
                    window.location = '/accounts';
                }
            });
        } else {
            Accounts.create(this.state.account).then((response) => {
                var data = parseResponse(response)
                if (!data.success) {
                    this.setState({error: formatError(data)});
                } else {
                    window.location = '/accounts';
                }
            });
        }
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
                    <h2>{T('store-account-assertion')}</h2>
                    <div className="col-12">
                        <AlertBox message={this.state.error} />

                        <form>
                            <fieldset>
                                <label htmlFor="account">{T('account')}:
                                    <input type="text" id="account" onChange={this.handleChangeAccount} value={this.state.account.AuthorityID} placeholder={T('account-description')} />
                                </label>
                                <label htmlFor="reseller">{T('reseller-features')}
                                    <input className="visible" type="checkbox" id="reseller" onChange={this.handleChangeReseller} checked={this.state.account.ResellerAPI} />
                                </label>
                            </fieldset>
                        </form>
                        <div>
                            <a href='/accounts' className="p-button--neutral">{T('cancel')}</a>
                            &nbsp;
                            <a href='/accounts' onClick={this.handleSaveClick} className="p-button--brand">{T('save')}</a>
                        </div>
                        {this.props.id? 
                            <label htmlFor="assertion">{T('assertion')}:
                                <pre>{this.state.account.Assertion}</pre>
                            </label>
                        :   ''
                        }
                    </div>
                </section>
                <br />
            </div>
        );
    }

}

export default AccountEdit