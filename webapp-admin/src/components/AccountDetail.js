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


class AccountDetail extends Component {

    constructor(props) {
        super(props);

        this.state = {
            account: {},
            subbrands: [{subbrand: {AuthorityID: 'developer'}, serialnumber: 'abc123', model: {"id":7,"brand-id":"YE9nAXA0anlmXFtdcPlBkiZPLJqVdK3u","model":"aaaaa","type":"device","keypair-id":1,"api-key":"Xmxdqvhm0T9arSS0SGdF6zI335dWPH7mxtJQQRoRVFlwVyODSIbAg","authority-id":"YE9nAXA0anlmXFtdcPlBkiZPLJqVdK3u","key-id":"TGfLv_PBYcnUZ81ZcPubBKI0UsePBVLPPgvzN_AuwdaEnYOR9l8aVjhcQIWweSjR","key-active":true}}],
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

    renderSubbrands() {
        
    }

    render() {
        if (!isUserAdmin(this.props.token)) {
            return (
                <div className="row">
                <AlertBox message={T('error-no-permissions')} />
                </div>
            )
        }

        var acc = this.state.account

        return (
            <div>
                <section className="row no-border">
                    <div className="p-card">
                        <h2 className="p-card__title">{T('account')}</h2>
                        <table className="p-card__content">
                          <tbody>
                            <tr>
                                <td className="col-3 label">{T('name')}:</td>
                                <td className="col-9">{acc.AuthorityID}</td>
                            </tr>
                            <tr>
                                <td className="col-3 label">{T('reseller')}:</td>
                                <td className="col-9">{acc.ResellerAPI ? <i className="fa fa-check"></i> :  <i className="fa fa-times"></i>}</td>
                            </tr>
                          </tbody>
                       </table>
                    </div>
                </section>

                <section className="row no-border">
                    <div className="p-card">
                        <h2 className="p-card__title">{T('subbrands')}</h2>
                        <table className="p-card__content">
                          <thead>
                            <tr>
                                <th className="small"></th><th>{T('subbrand')}</th><th>{T('model')}</th>
                            </tr>
                          </thead>
                          <tbody>
                              {this.state.subbrands.map((b) => {
                                return (
                                    <tr>
                                        <td></td>
                                        <td>{b.subbrand.AuthorityID}</td>
                                        <td>{b.model.model}</td>
                                        <td>{b.serialnumber}</td>
                                    </tr>
                                )
                              })}
                          </tbody>
                        </table>
                    </div>
                </section>
            </div>
        )

    }

}

export default AccountDetail
