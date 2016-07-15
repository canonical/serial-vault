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
var SigningLogRow = require('./SigningLogRow');
var Footer = require('./Footer');
var AlertBox = require('./AlertBox');
var SigningLogModel = require('../models/signinglog') 
var injectIntl = require('react-intl').injectIntl;

const PAGINATION_SIZE = 50;

var SigningLogList = React.createClass({
  getInitialState: function() {
    return {logs: this.props.logs || [], confirmDelete: null, message: null, fromID: null, showMore: true};
  },

  componentDidMount: function () {
    this.getLogs();
  },

  getLogs: function () {
    var self = this;

    SigningLogModel.list(this.state.fromID).then(function(response) {
      var data = JSON.parse(response.body);
      var message = "";
      if (!data.success) {
        message = data.message;
      }
      var showMore = data.logs.length == PAGINATION_SIZE;
      self.setState({logs: self.state.logs.concat(data.logs), message: message, showMore: showMore});
    });
  },

  getMoreLogs: function() {
    var fromID = this.state.logs[this.state.logs.length-1].id
    console.log(fromID);
    this.setState({fromID: fromID}, this.getLogs);
  },

  renderTable: function(M) {
    var self = this;

    if (this.state.logs.length > 0) {
      return (
        <div>
          <table>
            <thead>
              <tr>
                <th></th><th>{M({id:'brand'})}</th><th>{M({id:'model'})}</th><th>{M({id:'serial-number'})}</th><th>{M({id:'fingerprint'})}</th><th>{M({id:'date'})}</th>
              </tr>
            </thead>
            <tbody>
              {this.state.logs.map(function(l) {
                return (
                  <SigningLogRow key={l.id} log={l} delete={self.handleDelete} confirmDelete={self.state.confirmDelete}
                    deleteLog={self.handleDeleteLog} cancelDelete={self.handleDeleteLogCancel} />
                );
              })}
            </tbody>
          </table>
          {this.state.showMore? <button onClick={self.getMoreLogs}>{M({id:'more'})}</button> : ''}
        </div>
      );
    } else {
      return (
        <p>No models signed.</p>
      );
    }

  },

  render: function() {
    var M = this.props.intl.formatMessage;

    return (
        <div className="inner-wrapper">
          <Navigation active="signinglog" />

          <section className="row no-border">
            <h2>{M({id:'signinglog'})}</h2>
            <div className="twelve-col">
              <p>{M({id:'signinglog-description'})}</p>
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

module.exports = injectIntl(SigningLogList);