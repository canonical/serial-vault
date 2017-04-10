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
'use strict'

var React = require('react');
var AlertBox = require('./AlertBox');
var Navigation = require('./Navigation');
var Footer = require('./Footer');
var AlertBox = require('./AlertBox');
var Models = require('../models/models');
var Keypairs = require('../models/keypairs');
import {T} from './Utils';

var ModelEdit = React.createClass({
	getInitialState: function() {
		return {title: null, model: {}, error: null, keypairs: []};
	},

	componentDidMount: function() {
		this.getKeypairs();

		if (this.props.params.id) {
			this.setTitle('edit-model');
			this.getModel(this.props.params.id);
		} else {
			this.setTitle('new-model');
		}
	},

	setTitle: function(title) {
		this.setState({title: T(title)});
	},

	getModel: function(modelId) {
		var self = this;
		Models.get(modelId).then(function(response) {
			var data = JSON.parse(response.body);
			self.setState({model: data.model});
		});
	},

	getKeypairs: function() {
		var self = this;
		Keypairs.list().then(function(response) {
			var data = JSON.parse(response.body);
			var message = "";
			if (!data.success) {
				message = data.message;
			}
			self.setState({keypairs: data.keypairs, message: message});
		});
	},

	formatError: function(data) {
		var message = T(data.error_code);
		if (data.error_subcode) {
			message += ': ' + T(data.error_subcode);
		} else if (data.message) {
			message += ': ' + data.message;
		}
		return message;
	},

	handleChangeBrand: function(e) {
		var model = this.state.model;
		model['brand-id'] = e.target.value;
		this.setState({model: model});
	},

	handleChangeModel: function(e) {
		var model = this.state.model;
		model.model = e.target.value;
		this.setState({model: model});
	},

	handleChangePrivateKey: function(e) {
		var self = this;
		var model = this.state.model;
		model['keypair-id'] = parseInt(e.target.value);
		this.setState({model: model});
	},

	handleChangePrivateKeyUser: function(e) {
		var self = this;
		var model = this.state.model;
		model['keypair-id-user'] = parseInt(e.target.value);
		this.setState({model: model});
	},

	handleSaveClick: function(e) {
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
	},

	renderError: function() {
		if (this.state.error) {
			return (
				<AlertBox message={this.state.error} />
			);
		}
	},

	render: function() {
		var self = this;

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
});

module.exports = ModelEdit;
