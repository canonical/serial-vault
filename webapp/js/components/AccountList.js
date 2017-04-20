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
import {T} from './Utils'

class AccountList extends Component {

    constructor(props) {
        super(props);

        this.state = {
            accounts: [],
            keypairs: [],
            message: '',
        }

        this.getAccounts()
        this.getKeypairs()
    }

    getAccounts() {
        Accounts.list().then((response) => {
            var data = JSON.parse(response.body);
            console.log(data)
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
            console.log(data)
            var message = "";
            if (!data.success) {
                message = data.message;
            }
            this.setState({keypairs: data.keypairs, message: message});
        });
    }

    renderAccounts() {
        if (this.state.accounts.length > 0) {
            return (
                <table>
                <thead>
                    <tr>
                    <th className="small"></th><th>{T('authority-id')}</th><th>{T('assertion')}</th>
                    </tr>
                </thead>
                <tbody>
                    {this.state.accounts.map((acc) => {
                    return (
                        <tr>
                            <td></td>
                            <td>{acc.AuthorityID}</td>
                            <td><pre className="code">{acc.Assertion}</pre></td>
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
                    <th className="small"></th><th>{T('key-id')}</th><th>{T('assertion')}</th>
                    </tr>
                </thead>
                <tbody>
                    {this.state.keypairs.map((acc) => {
                    return (
                        <tr>
                            <td></td>
                            <td className="overflow">
                                {acc.AuthorityID}<br />{acc.KeyID}
                            </td>
                            <td>
                                {acc.Assertion ?
                                <pre className="code">{acc.Assertion}</pre>
                                : T('no-assertion')
                                }
                            </td>
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
        return (
            <div className="row">
                <section className="row">
                    <h2>{T('accounts')}</h2>

                    <div className="col-12">
                        {this.renderAccounts()}
                    </div>
                </section>

                <section className="row">
                    <h2>{T('account-keys')}</h2>

                    <div className="col-12">
                        {this.renderAccountKeys()}
                    </div>
                </section>
            </div>
        )
    }

}

export default AccountList
