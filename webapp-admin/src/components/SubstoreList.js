/*
 * Copyright (C) 2018 Canonical Ltd
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
import {T, isUserAdmin, formatError} from './Utils';
import SubstoreForm from './SubstoreForm';
import DialogBox from './DialogBox';


class SubstoreList extends Component {

    constructor(props) {
        super(props);

        this.state = {
            error: null,
            showNew: false,
            showEdit: null,
            showDelete: null,
            substore: {},
        }
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
            var substores = this.props.substores.filter( (s) => {
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
            var substores = this.props.substores.filter( (s) => {
                return s.id === id
            })
            this.setState({substore: substores[0], showDelete: id, showEdit: null, showNew: false})
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
            Accounts.storeNew(this.props.selectedAccount.ID, this.state.substore).then((response) => {
                var data = JSON.parse(response.body);
                if (response.statusCode >= 300) {
                    this.setState({error: formatError(data)});
                } else {
                    this.setState({substore: {}, showNew: false, showEdit: null, showDelete: null, error: null})
                    this.props.onRefresh(this.props.selectedAccount)
                }
            })
        } else {
            Accounts.storeUpdate(this.state.substore).then((response) => {
                var data = JSON.parse(response.body);
                if (response.statusCode >= 300) {
                    this.setState({error: formatError(data)});
                } else {
                    this.props.onRefresh(this.props.selectedAccount)
                    this.setState({substore: {}, showNew: false, showEdit: null, showDelete: null, error: null})
                }
            })
        }
    }

    handleDeleteSubstore = (e) => {
        e.preventDefault()

        Accounts.storeDelete(this.state.substore).then((response) => {
            var data = JSON.parse(response.body);
            if (response.statusCode >= 300) {
                this.setState({error: formatError(data)});
            } else {
                this.props.onRefresh(this.props.selectedAccount)
                this.setState({substore: {}, showNew: false, showEdit: null, showDelete: null, error: null})
            }
        })
    }

    handleCancelSubstore = (e) => {
        e.preventDefault()
        this.setState({substore: {}, showNew: false, showEdit: null, showDelete: null, error: null})
        this.props.onRefresh(this.props.selectedAccount)
    }

    renderActions(b) {
        if (this.state.showDelete === b.id) {
            return (
                <DialogBox message={T('confirm-store-delete')} handleYesClick={this.handleDeleteSubstore} handleCancelClick={this.handleCancelSubstore} small />
            );
        } else {
            return (
                <div>
                    <a href="" data-key={b.id} onClick={this.handleShowEdit} className="p-button--brand small" title={T('edit-model')}>
                        <i data-key={b.id} className="fa fa-pencil" />
                    </a>
                    &nbsp;
                    <a href="" data-key={b.id} onClick={this.handleShowDelete} className="p-button--neutral small" title={T('delete-model')}>
                        <i data-key={b.id} className="fa fa-trash" />
                    </a>
                </div>
            )
        }
    }

    renderSubstore(b) {
        return (
            <tr>
                <td>
                    {this.renderActions(b)}
                </td>
                <td className="overflow" title={b.fromModel.model}>{b.fromModel.model}</td>
                <td className="overflow" title={b.serialnumber}>{b.serialnumber}</td>
                <td className="overflow" title={b.store}>{b.store}</td>
                <td className="overflow" title={b.modelname}>{b.modelname}</td>
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

        return (
            <div>
                <section className="row no-border">
                    <div>
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

                        <table>
                          <thead>
                            <tr>
                                <th></th><th>{T('model')}</th><th>{T('serial-number')}</th>
                                <th>{T('substore')}</th><th>{T('substore-model')}</th>
                            </tr>
                          </thead>
                          <tbody>
                              {this.state.showNew ?
                                    <SubstoreForm substore={this.state.substore} models={this.props.models} token={this.props.token}
                                        onSave={this.handleSaveSubstore} onChange={this.handleSubstoreChange} onCancel={this.handleCancelSubstore} />
                                : ''
                              }
                              {this.props.substores.map((b) => {
                                if (b.id === this.state.showEdit) {
                                    return (<SubstoreForm substore={this.state.substore} models={this.props.models} token={this.props.token}
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

export default SubstoreList
