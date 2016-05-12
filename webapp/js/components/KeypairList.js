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
var injectIntl = require('react-intl').injectIntl;
var Keypairs = require('../models/keypairs');


var KeypairList = React.createClass({

	handleDeactivate: function(e) {
		var self = this;
		Keypairs.disable(e.target.getAttribute('data-key')).then(self.props.refresh);
	},

	handleActivate: function(e) {
		var self = this;
		Keypairs.enable(e.target.getAttribute('data-key')).then(self.props.refresh);
	},

	renderRow: function(keypr, M) {
		return (
			<tr key={keypr.ID}>
				<td>{keypr.Active ? <button data-key={keypr.ID} onClick={this.handleDeactivate} className="button--secondary">{M({id:'deactivate'})}</button> : <button data-key={keypr.ID} onClick={this.handleActivate} className="button--secondary">{M({id:'activate'})}</button>}</td>
				<td>{keypr.AuthorityID}</td>
				<td>{keypr.KeyID}</td>
				<td>{keypr.Active ? <i className="fa fa-check"></i> :  <i className="fa fa-times"></i>}</td>
			</tr>
		);
	},

	renderTable: function(M) {
    var self = this;

    if (this.props.keypairs.length > 0) {
      return (
        <table>
          <thead>
            <tr>
              <th></th><th>{M({id:'authority-id'})}</th><th>{M({id:'key-id'})}</th><th>{M({id:'active'})}</th>
            </tr>
          </thead>
          <tbody>
            {this.props.keypairs.map(function(keypr) {
              return self.renderRow(keypr, M);
            })}
          </tbody>
        </table>
      );
    } else {
      return (
        <p>{M({id:'no-signing-keys-found'})}</p>
      );
    }
  },

	render: function() {
		var M = this.props.intl.formatMessage;
		return this.renderTable(M);
	}

});

module.exports = injectIntl(KeypairList);
