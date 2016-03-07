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

var ModelEdit = React.createClass({
	getInitialState: function() {
		return {title: null, model: {}, error: null};
	},

	componentDidMount: function() {
		if (this.props.params.id) {
			this.setState({title: "Edit Model"});
			this.getModel(this.props.params.id);
		} else {
			this.setState({title: "New Model"});
		}
	},

	getModel: function(modelId) {
		var self = this;
		Models.get(modelId).then(function(response) {
			var data = JSON.parse(response.body);
			console.log(data);
			self.setState({model: data.model});
		});
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

	handleChangeRevision: function(e) {
		var model = this.state.model;
		model.revision = parseInt(e.target.value);
		this.setState({model: model});
	},

	handleChangePrivateKey: function(e) {
		var self = this;
		var model = this.state.model;

		// Get the file
		var reader = new FileReader();
		var file = e.target.files[0];

		reader.onload = function(upload) {
			// Get the base64 data from the URI
			var data = upload.target.result.split(',')[1];
			console.log(data);
			model['signing-key'] = data;
			self.setState({model: model});
		}

		// Read the file as store as data URL
		reader.readAsDataURL(file);
	},

	handleSaveClick: function(e) {
		e.preventDefault();
		var self = this;

		if (this.state.model.id) {
			// Update the existing model
			Models.update(this.state.model).then(function(response) {
				var data = JSON.parse(response.body);
				if (response.statusCode >= 300) {
					self.setState({error: data.message});
				} else {
					window.location = '/models';
				}
			});
		} else {
			// Create a new model
			Models.create(this.state.model).then(function(response) {
				var data = JSON.parse(response.body);
				if (response.statusCode >= 300) {
					self.setState({error: data.message});
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

	renderPrivateKey: function() {
		if (!this.state.model.id) {
			return (
				<li>
					<label htmlFor="privateKey">Private Key for Signing:</label>
					<input type="file" id="privateKey" placeholder="The signing-key that will be used to sign the device identity"
						onChange={this.handleChangePrivateKey}/>
				</li>
			);
		}
	},

	render: function() {
		return (
			<div>
				<Navigation active="models" />

				<section className="row">
					  <h2>{this.state.title}</h2>

						<AlertBox message={this.state.error} />

						<form>
							<fieldset>
								<ul>
									<li>
										<label htmlFor="brand">Brand:</label>
										<input type="text" id="brand" placeholder="The name of the device brand"
											value={this.state.model['brand-id']} onChange={this.handleChangeBrand} />
									</li>
									<li>
										<label htmlFor="model">Model:</label>
										<input type="text" id="model" placeholder="The name of the device model"
											value={this.state.model.model} onChange={this.handleChangeModel}/>
									</li>
									<li>
										<label htmlFor="revision">Revision:</label>
										<input type="number" id="revision" placeholder="The revision of the device"
											value={this.state.model.revision} onChange={this.handleChangeRevision}/>
									</li>
									{this.renderPrivateKey()}
								</ul>
							</fieldset>
						</form>

						<div>
							<a href='/models' onClick={this.handleSaveClick} className="button--primary">Save</a>
							&nbsp;
							<a href='/models' className="button--secondary">Cancel</a>
						</div>
				</section>

				<Footer />
			</div>
		)
	}
});

module.exports = ModelEdit;
