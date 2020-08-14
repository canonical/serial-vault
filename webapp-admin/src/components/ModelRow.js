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
import DialogBox from './DialogBox';
import {T} from './Utils'

class ModelRow extends Component {

    copyToClipboard = (e) => {
        e.preventDefault()
        const el = document.createElement('textarea');
        el.value = e.target.getAttribute('data-key');
        document.body.appendChild(el);
        el.select();
        document.execCommand('copy');
        document.body.removeChild(el);
    }

    renderActions() {
        if (this.props.model.id !== this.props.confirmDelete) {
            return (
                <div>
                    <a href={'/models/'.concat(this.props.model.id, '/edit')} className="p-button--brand small" title={T('edit-model')}><i className="fa fa-pencil"></i></a>
                    &nbsp;
                    <button onClick={this.props.delete} data-key={this.props.model.id} className="p-button--neutral small" title={T('delete-model')}>
                        <i className="fa fa-trash" data-key={this.props.model.id}></i>
                    </button>
                </div>
            );
        } else {
            return (
                <DialogBox message={T('confirm-model-delete')} handleYesClick={this.props.deleteModel} handleCancelClick={this.props.cancelDelete} small />
            );
        }
    }

    render() {
        var fingerprint = this.props.model['key-id'];
        var fingerprintUser = this.props.model['key-id-user'];
        var fingerprintModel;
        if (this.props.model['keypair-id-model'] > 0) {
            fingerprintModel = this.props.model['key-id-model'];
        } else {
            fingerprintModel = ""
        }
        return (
            <tr>
                <td>
                    {this.renderActions()}
                </td>
                <td className="overflow" title={this.props.model.model}>
                    <button onClick={this.copyToClipboard} data-key={this.props.model['api-key']} className="p-button--neutral small" title={T('copy-api-key')}>
                        <i className="fa fa-clipboard" data-key={this.props.model['api-key']} />
                    </button>
                    &nbsp;
                    {this.props.model.model}
                </td>
                <td className="overflow" title={fingerprint} >{fingerprint}</td>
                <td className="overflow" title={fingerprintUser} >{fingerprintUser}</td>
                <td>{this.props.model['key-active'] && this.props.model['key-active-user'] ? <i className="fa fa-check"></i> :  <i className="fa fa-times"></i>}</td>
                <td className="overflow" title={fingerprintModel} >
                    <button className="p-button--neutral small" title={T('assertion-settings')} data-key={this.props.model.id} onClick={this.props.showAssert}>
                        <i className="fa fa-sliders" aria-hidden="true" data-key={this.props.model.id} />
                    </button>                    
                </td>
            </tr>
        )
    }
}

export default ModelRow;
