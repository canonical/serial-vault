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
import Accounts from '../models/accounts'
import Keypairs from '../models/keypairs'
import Models from '../models/models'
import AlertBox from './AlertBox'
import {T, isUserAdmin} from './Utils'

class AccountList extends Component {

    constructor(props) {
        super(props);

        this.state = {
            accounts: props.accounts || [],
            keypairs: props.keypairs || [],
            models: props.models || [],
            message: '',
        }
    }

    componentDidMount() {
        this.getModels()
        this.getAccounts()
        this.getKeypairs()
    }

    getAccounts() {
        Accounts.list().then((response) => {
            var data = JSON.parse(response.body);
            var message = "";
            if (!data.success) {
                message = data.message;
            }
            this.setState({accounts: data.accounts, message: message});
        });
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

    getModels() {
        Models.list().then((response) => {
            var data = JSON.parse(response.body);
            var message = "";
            if (!data.success) {
                message = data.message;
            }
            this.setState({models: data.models, message: message});
        });
    }

    // Indicates whether the key has everything uploaded for it
    renderKeyStatus(acc) {
        // Check if the key is used for signing system-users on any models
        if (!this.state.models.find(m => (m['authority-id-user'] === acc.AuthorityID) & (m['key-id-user'] === acc.KeyID))) {
            return <p>{T('not-used-signing')}</p>
        }

        // Check that we have an account assertion
        var messages = []
        if (!this.state.accounts.find(a => a.AuthorityID === acc.AuthorityID)) {
            messages.push(T('no-assertion'))
        }

        // Check if we have an account key assertion
        if ((!acc.Assertion) || (acc.Assertion.length === 0)) {
            messages.push(T('no-assertion-key'))
        }
        
        if (messages.length === 0) {
            return (
                <div>
                    <p title={acc.Assertion}><i className="fa fa-check information positive"></i> {T('complete')}</p>
                </div>
            )
        }

        return (
            <div>
                {messages.map((m, index, array) => {
                    return (
                        <p key={index}><i className="fa fa-exclamation-triangle warning"></i> {m}</p>
                    )
                })}
            </div>
        )
    }

    renderAccounts() {
        if (this.state.accounts.length > 0) {
            return (
                <table>
                <thead>
                    <tr>
                        {isUserAdmin(this.props.token) ? <th className="small"></th> : ''}
                        <th>{T('account')}</th><th>{T('assertion-status')}</th><th className="small">{T('reseller')}</th>
                    </tr>
                </thead>
                <tbody>
                    {this.state.accounts.map((acc) => {
                    return (
                        <tr key={acc.ID}>
                            {isUserAdmin(this.props.token) ? 
                                <td>
                                    <div>
                                    <a href={'/accounts/account/' + acc.ID} className="p-button--brand small" title={T('edit-account')}><i className="fa fa-pencil"></i></a>
                                    </div>
                                </td>
                                : ''
                            }
                            <td><a href={'/accounts/view/' + acc.ID}>{acc.AuthorityID}</a></td>
                            <td>
                                <p title={acc.Assertion}><i className="fa fa-check information positive"></i> {T('complete')}</p>
                            </td>
                            <td>{acc.ResellerAPI ? <i className="fa fa-check"></i> :  <i className="fa fa-times"></i>}</td>
                        </tr>
                    );
                    })}
                </tbody>
                </table>
            );
        } else {
            return (
                <p>{T('no-assertions')}</p>
            );
        }
    }

    renderAccountKeys() {
        if (this.state.keypairs.length > 0) {
            return (
                <table>
                <thead>
                    <tr>
                        <th>{T('key-id')}</th><th>{T('assertion-status')}</th><th className="small">{T('reseller')}</th>
                    </tr>
                </thead>
                <tbody>
                    {this.state.keypairs.map((acc) => {
                    return (
                        <tr key={acc.ID}>
                            <td className="overflow">
                                {acc.AuthorityID}<br />{acc.KeyID}
                            </td>
                            <td>
                                {this.renderKeyStatus(acc)}
                            </td>
                            <td>{acc.ResellerAPI ? <i className="fa fa-check"></i> :  <i className="fa fa-times"></i>}</td>
                        </tr>
                    );
                    })}
                </tbody>
                </table>
            );
        } else {
            return (
                <p>{T('no-assertions')}</p>
            );
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
            <div className="row" ref="root">
                <section className="row">

                    <div className="u-equal-height">
                        <h2 className="col-5">{T('accounts')}</h2>
                        &nbsp;
                        <div className="col-1">
                            <a href="/accounts/new" className="p-button--brand" title={T('new-account-assertion')}>
                                <i className="fa fa-plus"></i>
                            </a>
                        </div>
                        <div className="col-1">
                            <a href="/accounts/upload" className="p-button--brand" title={T('upload-account-assertion')}>
                                <i className="fa fa-upload"></i>
                            </a>
                        </div>
                    </div>

                    <div className="col-12">
                        {this.renderAccounts()}
                    </div>
                </section>

                <section className="row">
                    <div className="u-equal-height">
                        <h2 className="col-5">{T('account-keys')}</h2>
                        &nbsp;
                        <div className="col-1">
                            <a href="/accounts/key-assertion" className="p-button--brand" title={T('new-account-assertion')}>
                                <i className="fa fa-plus"></i>
                            </a>
                        </div>
                    </div>

                    <div className="col-12">
                        {this.renderAccountKeys()}
                    </div>
                </section>
            </div>
        )
    }

}

export default AccountList
