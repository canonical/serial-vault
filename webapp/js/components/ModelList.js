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
var AlertBox = require('./AlertBox');
var Models = require('../models/models');
var ReactIntl = require('react-intl');
import {injectIntl} from 'react-intl';

var ModelList = React.createClass({

  getInitialState: function() {
    return {models: []};
  },

  componentDidMount: function() {
    this.getModels();
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

  renderTable: function(M) {
    var self = this;

    if (this.state.models.length > 0) {
      return (
        <table>
          <thead>
            <tr>
              <th></th><th>{M({id:'brand'})}</th><th>{M({id:'model'})}</th><th>{M({id:'revision'})}</th>
            </tr>
          </thead>
          <tbody>
            {this.state.models.map(function(mdl) {
              return (
                <ModelRow key={mdl.id} model={mdl} />
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
        <div>
          <Navigation active="models" />

          <section className="row no-border">
            <h2>{M({id:'models'})} <a href="/models/new" className="button--primary small" title="Add a new model">
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
          </section>

          <Footer />
        </div>
    );
  }
});

module.exports = injectIntl(ModelList);
