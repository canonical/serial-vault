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
var Keypairs = require('../models/keypairs');
var Vault = require('../models/vault')
import {T} from './Utils'


var KeypairList = React.createClass({

  componentDidMount: function() {
		this.getVersion();
	},

	getVersion: function() {
		Vault.version().then(function(response) {
		  this.updateCsrfToken(response);
		});
	},

	updateCsrfToken: function(response) {
		document.getElementsByTagName("meta")["gorilla.csrf.Token"].setAttribute("content", response.headers['x-csrf-token']);
	},

	handleDeactivate: function(e) {
		var self = this;
		Keypairs.disable(e.target.getAttribute('data-key')).then(self.props.refresh);
	},

	handleActivate: function(e) {
		var self = this;
		Keypairs.enable(e.target.getAttribute('data-key')).then(self.props.refresh);
	},

	renderRow: function(keypr) {
		return (
			<tr key={keypr.ID}>
				<td>{keypr.Active ? <button data-key={keypr.ID} onClick={this.handleDeactivate} className="button--secondary">{T('deactivate')}</button> : <button data-key={keypr.ID} onClick={this.handleActivate} className="button--secondary">{T('activate')}</button>}</td>
				<td>{keypr.AuthorityID}</td>
				<td className="overflow" title={keypr.KeyID}>{keypr.KeyID}</td>
				<td>{keypr.Active ? <i className="fa fa-check"></i> :  <i className="fa fa-times"></i>}</td>
			</tr>
		);
	},

	render: function() {

    var self = this;

    if (this.props.keypairs.length > 0) {
      return (
        <table>
          <thead>
            <tr>
              <th></th><th>{T('authority-id')}</th><th>{T('key-id')}</th><th>{T('active')}</th>
            </tr>
          </thead>
          <tbody>
            {this.props.keypairs.map(function(keypr) {
              return self.renderRow(keypr);
            })}
          </tbody>
        </table>
      );
    } else {
      return (
        <p>{T('no-signing-keys-found')}</p>
      );
    }
	}

});

module.exports = KeypairList;
