// -*- Mode: Go; indent-tabs-mode: t -*-

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
var Navigation = require('./Navigation');
var Footer = require('./Footer');
var ModelRow = require('./ModelRow');
var KeypairList = require('./KeypairList');
var AlertBox = require('./AlertBox');
var Models = require('../models/models');
var Keypairs = require('../models/keypairs');
var injectIntl = require('react-intl').injectIntl;

var ModelList = React.createClass({

  getInitialState: function() {
    return {models: this.props.models || [], keypairs: this.props.keypairs || [], confirmDelete: null, message: null};
  },

  componentDidMount: function() {
    this.refresh();
  },

  refresh: function() {
    this.getModels();
    this.getKeypairs();
  },

  getModels: function() {
    var self = this;
    Models.list().then(function(response) {
      var data = JSON.parse(response.body);
      var message = "";
      if (!data.success) {
        message = data.message;
      }
      self.setState({models: data.models, message: message});
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
		var message = this.props.intl.formatMessage({id: data.error_code});
		if (data.error_subcode) {
			message += ': ' + this.props.intl.formatMessage({id: data.error_subcode});
		} else if (data.message) {
			message += ': ' + data.message;
		}
		return message;
	},

  handleDelete: function(e) {
    e.preventDefault();
    this.setState({confirmDelete: parseInt(e.target.getAttribute('data-key'))});
  },

  handleDeleteModel: function(e) {
    e.preventDefault();
    var self = this;
    var models = this.state.models.filter(function(mdl) {
      return mdl.id === self.state.confirmDelete;
    });
    if (models.length === 0) {
      return;
    }

    Models.delete(models[0]).then(function(response) {
      var data = JSON.parse(response.body);
      if ((response.statusCode >= 300) || (!data.success)) {
        self.setState({message: self.formatError(data)});
      } else {
        window.location = '/models';
      }
    });
  },

  handleDeleteModelCancel: function(e) {
    e.preventDefault();
    this.setState({confirmDelete: null});
  },

  renderTable: function(M) {
    var self = this;

    if (this.state.models.length > 0) {
      return (
        <table>
          <thead>
            <tr>
              <th></th><th>{M({id:'brand'})}</th><th>{M({id:'model'})}</th><th>{M({id:'revision'})}</th><th>{M({id:'signing-key'})}</th>
            </tr>
          </thead>
          <tbody>
            {this.state.models.map(function(mdl) {
              return (
                <ModelRow key={mdl.id} model={mdl} delete={self.handleDelete} confirmDelete={self.state.confirmDelete}
                  deleteModel={self.handleDeleteModel} cancelDelete={self.handleDeleteModelCancel} />
              );
            })}
          </tbody>
        </table>
      );
    } else {
      return (
        <p>No models found.</p>
      );
    }
  },

  render: function() {
    var M = this.props.intl.formatMessage;

    return (
        <div className="inner-wrapper">
          <Navigation active="models" />

          <section className="row no-border">
            <h2>{M({id:'models'})} <a href="/models/new" className="button--primary small" title={M({id:'add-new-model'})}>
                        <i className="fa fa-plus"></i>
                      </a>
            </h2>
            <div className="twelve-col">
              <p>{M({id:'models_available'})}:</p>
            </div>
            <div className="twelve-col">
              <AlertBox message={this.state.message} />
            </div>
            <div className="twelve-col">
              {this.renderTable(M)}
            </div>

            <h2>
              {M({id:'signing-keys'})}&nbsp;
              <a href="/models/keypairs/new" className="button--primary small" title={M({id:'add-new-signing-key'})}>
                <i className="fa fa-plus"></i>
              </a>
            </h2>
            <div className="twelve-col">
              <KeypairList keypairs={this.state.keypairs} refresh={this.refresh} />
            </div>
          </section>

          <Footer />
        </div>
    );
  }
});

module.exports = injectIntl(ModelList);
