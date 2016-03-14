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
var AlertBox = require('./AlertBox');
var DialogBox = require('./DialogBox');
var Keys = require('../models/keys');
var injectIntl = require('react-intl').injectIntl;

var KeyList = React.createClass({
  getInitialState: function() {
    return {keys: this.props.keys || []};
  },

  componentDidMount: function() {
    this.getKeys();
  },

  getKeys: function() {
    var self = this;
    Keys.list().then(function(response) {
      var data = JSON.parse(response.body);
      var message = "";
      if (!data.success) {
        message = data.message;
      }
      self.setState({keys: data.keys, error: message});
    });
  },

  handleDeleteClick: function(e) {
    e.preventDefault();
    var self = this;

    var index = parseInt(e.target.getAttribute('data-key'));
    this.setState({confirmDelete: index});
  },

  handleRemoveClick: function(e) {
    e.preventDefault();
    var self = this;

    Keys.delete(this.state.keys[this.state.confirmDelete]).then(function(response) {
      var data = JSON.parse(response.body);
      if ((response.statusCode >= 300) || (!data.success)) {
        self.setState({error: data.message, confirmDelete: null});
      } else {
        self.setState({confirmDelete: null});
        self.getKeys();
      }
    });

  },

  handleCancelClick: function(e) {
    e.preventDefault();
    this.setState({confirmDelete: null});
    this.getKeys();
  },

  renderDialogOrButton: function(index) {
    if (this.state.confirmDelete === index) {
      return (
        <DialogBox message={this.props.intl.formatMessage({id: "public-key-confirm"})}
            handleYesClick={this.handleRemoveClick} handleCancelClick={this.handleCancelClick} />
      )
    }
  },

  renderTable: function(M) {
    var self = this;

    if (this.state.keys.length > 0) {
      var index = -1;
      return (
        <table className="PublicKey">
          <thead>
            <tr>
              <th>{M({id:'public-key'})}</th>
            </tr>
          </thead>
          <tbody>
            {this.state.keys.map(function(key) {
              index += 1;
              return (
                <tr key={index}>
          				<td>
                      {self.renderDialogOrButton(index)}
                      <div className="one-col">
                        <a onClick={self.handleDeleteClick} data-key={index} href="" className="button--secondary">
                          <i className="fa fa-times" onClick={self.handleDeleteClick} data-key={index}></i>
                        </a>
                      </div>
                      <div className="one-col"></div>
                      <div className="nine-col wide">{self.state.keys[index]}</div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      );
    } else {
      return (
        <p>No public keys found.</p>
      );
    }
  },

	render: function() {
    var M = this.props.intl.formatMessage;
    return (
        <div>
          <Navigation active="keys" />

          <section className="row no-border">
            <h2>{M({id:'public-keys'})} <a href="/keys/new" className="button--primary small" title={M({id:'public-key-description'})}>
                        <i className="fa fa-plus"></i>
                      </a>
            </h2>
            <div className="twelve-col">
              <p>{M({id:'public-keys-authorized'})}:</p>
            </div>
            <div className="twelve-col">
              <AlertBox message={this.state.error} />
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

module.exports = injectIntl(KeyList);
