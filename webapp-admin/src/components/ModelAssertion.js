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

import React, {Component} from 'react';
import AlertBox from './AlertBox';
import Keypairs from '../models/keypairs';
import Models from '../models/models';
import {T, formatError, isUserAdmin} from './Utils';

class ModelAssertion extends Component {

    constructor(props) {

        super(props)

        var assertion = {}
        if ((this.props.model) && (this.props.model.assertion)) {
            assertion = this.props.model.assertion
        }
        if (!assertion.model_id) {
            assertion.model_id = this.props.model.id
        }

        this.state = {
            error: null,
            keypairs: [],
            assertion: assertion,
        }

        this.getKeypairs();
    }

    getKeypairs() {
        Keypairs.list().then((response) => {
            var data = JSON.parse(response.body);
            var message = "";
            if (!data.success) {
                message = data.message;
            }

            var keypairs = data.keypairs.filter( (k) => {
                return k.AuthorityID === this.props.model['brand-id']
            })

            this.setState({keypairs: keypairs, message: message});
        });
    }

    handleChangePrivateKeyModel = (e) => {
        var assertion = this.state.assertion;
        assertion['keypair_id'] = parseInt(e.target.value, 10);
        this.setState({assertion: assertion});
    }

    handleChangeSeries = (e) => {
        var assertion = this.state.assertion;
        assertion['series'] = parseInt(e.target.value, 10);
        this.setState({assertion: assertion});
    }

    handleChangeArchitecture = (e) => {
        var assertion = this.state.assertion;
        assertion['architecture'] = e.target.value;
        this.setState({assertion: assertion});
    }

    handleChangeRevision = (e) => {
        var assertion = this.state.assertion;
        assertion['revision'] = parseInt(e.target.value, 10);
        this.setState({assertion: assertion});
    }

    handleChangeGadget = (e) => {
        var assertion = this.state.assertion;
        assertion['gadget'] = e.target.value;
        this.setState({assertion: assertion});
    }

    handleChangeKernel = (e) => {
        var assertion = this.state.assertion;
        assertion['kernel'] = e.target.value;
        this.setState({assertion: assertion});
    }

    handleChangeStore = (e) => {
        var assertion = this.state.assertion;
        assertion['store'] = e.target.value;
        this.setState({assertion: assertion});
    }

    handleChangeSnaps = (e) => {
        var assertion = this.state.assertion;
        assertion['required_snaps'] = e.target.value;
        this.setState({assertion: assertion});
    }

    handleSave = (e) => {
        e.preventDefault()
        if (!isUserAdmin(this.props.token)) {
            window.location = '/models';
        }

        Models.assertion(this.state.assertion).then((response) => {
            var data = JSON.parse(response.body);
            console.log(data)
            if (response.statusCode >= 300) {
                this.setState({error: formatError(data)});
            } else {
                window.location = '/models';
            }
        })
    }

    render() {

      var ma = this.state.assertion

      return (
        <tr>
            <td colSpan="7">
                <h5>{T('assertion-settings')}</h5>
                <AlertBox message={this.state.error} />
                <form>
                    <fieldset>
                        <label htmlFor="keypair-model">{T('private-key-model')}:
                            <select value={ma.keypair_id} id="keypair-model" onChange={this.handleChangePrivateKeyModel}>
                                <option></option>
                                {this.state.keypairs.map(function(kpr) {
                                    if (kpr.Active) {
                                        return <option key={kpr.ID} value={kpr.ID}>{kpr.AuthorityID}/{kpr.KeyID}</option>;
                                    } else {
                                        return <option key={kpr.ID} value={kpr.ID}>{kpr.AuthorityID}/{kpr.KeyID} ({T('inactive')})</option>;
                                    }
                                })}
                            </select>
                        </label>
                        <label htmlFor="series">{T('series')}:
                            <input type="number" id="series" placeholder={T('series-description')} min="16"
                                value={ma['series']} onChange={this.handleChangeSeries} />
                        </label>
                        <label htmlFor="architecture">{T('architecture')}:
                            <input type="text" id="architecture" placeholder={T('architecture-description')}
                                value={ma['architecture']} onChange={this.handleChangeArchitecture} />
                        </label>
                        <label htmlFor="revision">{T('revision')}:
                            <input type="number" id="revision" placeholder={T('revision-description')}
                                value={ma['revision']} onChange={this.handleChangeRevision} />
                        </label>
                        <label htmlFor="gadget">{T('gadget')}:
                            <input type="text" id="gadget" placeholder={T('gadget-description')}
                                value={ma['gadget']} onChange={this.handleChangeGadget} />
                        </label>
                        <label htmlFor="kernel">{T('kernel')}:
                            <input type="text" id="kernel" placeholder={T('kernel-description')}
                                value={ma['kernel']} onChange={this.handleChangeKernel} />
                        </label>
                        <label htmlFor="store">{T('store')}:
                            <input type="text" id="store" placeholder={T('store-description')}
                                value={ma['store']} onChange={this.handleChangeStore} />
                        </label>
                        <label htmlFor="required-snaps">{T('required-snaps')}:
                            <textarea onChange={this.handleChangeSnaps} defaultValue={ma['required_snaps']} name="required-snaps"
                                placeholder={T('required-snaps-description')} />
                        </label>
                    </fieldset>
                    {isUserAdmin(this.props.token) ?
                      <span>
                        <button className="p-button--neutral" onClick={this.props.cancel} data-key={ma.model_id}>{T('cancel')}</button>
                        <button className="p-button--brand" onClick={this.handleSave} data-key={ma.model_id}>{T('save')}</button>
                      </span>
                     : <button className="p-button--neutral" onClick={this.props.cancel} data-key={ma.model_id}>{T('close')}</button>
                    }
                </form>
            </td>
        </tr>
      )
    }
}

export default ModelAssertion;
