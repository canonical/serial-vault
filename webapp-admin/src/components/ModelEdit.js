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
import AlertBox from './AlertBox';
import Models from '../models/models';
import Keypairs from '../models/keypairs';
import {T, isUserAdmin} from './Utils';

class ModelEdit extends Component {

    constructor(props) {

        super(props)
        this.state = {
            title: null,
            model: {},
            error: null,
            hideForm: false,
            keypairs: [],
        }
    }

    componentDidMount() {
        this.getKeypairs();

        if (this.props.id) {
            this.setTitle('edit-model');
            this.getModel(this.props.id);
        } else {
            this.setTitle('new-model');
        }
    }

    setTitle(title) {
        this.setState({title: T(title)});
    }

    getModel(modelId) {
        Models.get(modelId).then((response) => {
            var data = JSON.parse(response.body);

            if (response.statusCode >= 300) {
                this.setState({error: this.formatError(data), hideForm: true});
            } else {
                this.setState({model: data.model, hideForm: false});
            }
        });
    }

    getKeypairs() {
        var self = this;
        Keypairs.list().then(function(response) {
            var data = JSON.parse(response.body);
            var message = "";
            if (!data.success) {
                message = data.message;
            }
            self.setState({keypairs: data.keypairs, message: message});
        });
    }

    formatError(data) {
        var message = T(data.error_code);
        if (data.error_subcode) {
            message += ': ' + T(data.error_subcode);
        }
        if (data.message) {
            message += ': ' + data.message;
        }
        return message;
    }

    handleChangeBrand = (e) => {
        var model = this.state.model;
        model['brand-id'] = e.target.value;
        this.setState({model: model});
    }

    handleChangeModel = (e) => {
        var model = this.state.model;
        model.model = e.target.value;
        this.setState({model: model});
    }

    handleChangeAPIKey = (e) => {
        var model = this.state.model;
        model['api-key'] = e.target.value;
        this.setState({model: model});
    }

    handleChangePrivateKey = (e) => {
        var model = this.state.model;
        model['keypair-id'] = parseInt(e.target.value, 10);
        this.setState({model: model});
    }

    handleChangePrivateKeyUser = (e) => {
        var model = this.state.model;
        model['keypair-id-user'] = parseInt(e.target.value, 10);
        this.setState({model: model});
    }

    handleSaveClick = (e) => {
        e.preventDefault();
        var self = this;

        if (this.state.model.id) {
            // Update the existing model
            Models.update(this.state.model).then(function(response) {
                var data = JSON.parse(response.body);
                if (response.statusCode >= 300) {
                    self.setState({error: self.formatError(data)});
                } else {
                    window.location = '/models';
                }
            });
        } else {
            // Create a new model
            Models.create(this.state.model).then(function(response) {
                var data = JSON.parse(response.body);
                if (response.statusCode >= 300) {
                    self.setState({error: self.formatError(data)});
                } else {
                    window.location = '/models';
                }
            });
        }
    }

    renderError() {
        if (this.state.error) {
            return (
                <AlertBox message={this.state.error} />
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

        if (this.state.hideForm) {
            return (
                <div className="row">
                <AlertBox message={this.state.error} />
                </div>
            )
        }

        return (
            <div className="row">
                <section className="row">
                      <h2>{this.state.title}</h2>

                        <AlertBox message={this.state.error} />

                        <form>
                            <fieldset>
                                <label htmlFor="brand">{T('brand')}:
                                    <input type="text" id="brand" placeholder={T('brand-description')}
                                        value={this.state.model['brand-id']} onChange={this.handleChangeBrand} />
                                </label>
                                <label htmlFor="model">{T('model')}:
                                    <input type="text" id="model" placeholder={T('model-description')}
                                        value={this.state.model.model} onChange={this.handleChangeModel}/>
                                </label>
                                <label htmlFor="model">{T('api-key')}:
                                    <input type="text" id="api-key" placeholder={T('api-key-description')}
                                        value={this.state.model['api-key']} onChange={this.handleChangeAPIKey}/>
                                </label>
                                <label htmlFor="keypair">{T('private-key')}:
                                    <select value={this.state.model['keypair-id']} id="keypair" onChange={this.handleChangePrivateKey}>
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
                                <label htmlFor="keypair-user">{T('private-key-user')}:
                                    <select value={this.state.model['keypair-id-user']} id="keypair-user" onChange={this.handleChangePrivateKeyUser}>
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
                            </fieldset>
                        </form>

                        <div>
                            <a href='/models' className="p-button--neutral">{T('cancel')}</a>
                            &nbsp;
                            <a href='/models' onClick={this.handleSaveClick} className="p-button--brand">{T('save')}</a>
                        </div>
                </section>
                <br />
            </div>
        )
    }
}

export default ModelEdit;
