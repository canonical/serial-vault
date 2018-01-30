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
import Models from '../models/models'
import {T, isUserAdmin, formatError} from './Utils';
import SubstoreForm from './SubstoreForm';


class AccountDetail extends Component {

    constructor(props) {
        super(props);

        this.state = {
            account: {},
            substores: [],
            models: [],
            error: null,
            showNew: false,
            showEdit: null,
            showDelete: null,
            substore: {},
        }

        if (this.props.id) {
            this.getAccount(this.props.id)
            this.getSubstores(this.props.id)
        }
    }

    getAccount(accountId) {
        Accounts.get(accountId).then((response) => {
            var data = JSON.parse(response.body);

            if (response.statusCode >= 300) {
                this.setState({error: formatError(data), hideForm: true, models: []});
            } else {
                this.setState({account: data.account, hideForm: false});
                this.getModels(data.account.AuthorityID)
            }
        });
    }

    getSubstores(accountId) {
        Accounts.stores(accountId).then((response) => {
            var data = JSON.parse(response.body);

            if (response.statusCode >= 300) {
                this.setState({error: formatError(data), hideForm: true});
            } else {
                this.setState({substores: data.substores, showNew: false, showEdit: null, showDelete: null, hideForm: false});
            }
        });
    }

    getModels(authorityId) {
        Models.list().then((response) => {
          var data = JSON.parse(response.body);

          if (response.statusCode >= 300) {
            this.setState({error: formatError(data)});
          } else {
            // Only show the models for this account
            var mdls = data.models.filter((m) => {
                return m['brand-id'] === authorityId
            })
            this.setState({models: mdls});
          }
        });
    }

    handleShowNew = (e) => {
        e.preventDefault();
        this.setState({showNew: true, showEdit: null, showDelete: null, substore: {}})
    }

    handleShowEdit = (e) => {
        e.preventDefault();
        var id = parseInt(e.target.getAttribute('data-key'), 10);
        if (this.state.showEdit === id) {
            this.setState({showEdit: null, substore: {}, showDelete: null, showNew: false})
        } else {
            var substores = this.state.substores.filter( (s) => {
                return s.id === id
            })

            this.setState({substore: substores[0], showEdit: id, showDelete: null, showNew: false})
        }
    }

    handleShowDelete = (e) => {
        e.preventDefault();
        var id = parseInt(e.target.getAttribute('data-key'), 10);
        if (this.state.showDelete === id) {
            this.setState({showDelete: null, showEdit: null, showNew: false})
        } else {
            this.setState({showDelete: id, showEdit: null, showNew: false})
        }
    }

    handleSubstoreChange = (field, value) => {
        var s = this.state.substore
        s[field] = value
        this.setState({substore: s})
    }

    handleSaveSubstore = (e) => {
        e.preventDefault()

        if (!isUserAdmin(this.props.token)) {
            window.location = '/accounts';
        }

        if (!this.state.substore.id) {
            Accounts.storeNew(this.props.id, this.state.substore).then((response) => {
                var data = JSON.parse(response.body);
                if (response.statusCode >= 300) {
                    this.setState({error: formatError(data)});
                } else {
                    this.getSubstores(this.props.id)
                }
            })
        } else {
            Accounts.storeUpdate(this.state.substore).then((response) => {
                var data = JSON.parse(response.body);
                if (response.statusCode >= 300) {
                    this.setState({error: formatError(data)});
                } else {
                    this.getSubstores(this.props.id)
                }
            })
        }
    }

    handleCancelSubstore = (e) => {
        e.preventDefault()
        this.setState({substore: {}, showNew: false, showEdit: null, showDelete: null})
    }

    renderSubstore(b) {
        return (
            <tr>
                <td>
                    <div>
                        <a href="" data-key={b.id} onClick={this.handleShowEdit} className="p-button--brand small" title={T('edit-model')}>
                            <i data-key={b.id} className="fa fa-pencil"></i>
                        </a>
                        &nbsp;
                        <a href="" data-key={b.id} onClick={this.handleShowDelete} className="p-button--neutral small" title={T('delete-model')}>
                            <i data-key={b.id} className="fa fa-trash"></i>
                        </a>
                    </div>
                </td>
                <td className="overflow" title={b.fromModel.model}>{b.fromModel.model}</td>
                <td className="overflow" title={b.serialnumber}>{b.serialnumber}</td>
                <td className="overflow" title={b.store}>{b.store}</td>
                <td className="overflow" title={b.toModel.model}>{b.toModel.model}</td>
            </tr>
        )
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
                        <div className="u-equal-height">
                            <h2 className="p-card__title col-5">{T('substores')}</h2>
                            &nbsp;
                            <div className="col-1">
                                <a href="" onClick={this.handleShowNew} className="p-button--brand" title={T('new-substore-device')}>
                                    <i className="fa fa-plus"></i>
                                </a>
                            </div>
                        </div>

                        <AlertBox message={this.state.error} />

                        <table className="p-card__content">
                          <thead>
                            <tr>
                                <th className="small"></th><th>{T('model')}</th><th>{T('serial-number')}</th>
                                <th>{T('substore')}</th><th>{T('substore-model')}</th>
                            </tr>
                          </thead>
                          <tbody>
                              {this.state.showNew ?
                                    <SubstoreForm substore={this.state.substore} models={this.state.models} token={this.props.token}
                                        onSave={this.handleSaveSubstore} onChange={this.handleSubstoreChange} onCancel={this.handleCancelSubstore} />
                                : ''
                              }
                              {this.state.substores.map((b) => {
                                if (b.id === this.state.showEdit) {
                                    return (<SubstoreForm substore={b} models={this.state.models} token={this.props.token}
                                                onSave={this.handleSaveSubstore} onChange={this.handleSubstoreChange} onCancel={this.handleCancelSubstore} />)
                                } else {
                                    return this.renderSubstore(b)
                                }
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
