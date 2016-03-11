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
var Keys = require('../models/keys');
var injectIntl = require('react-intl').injectIntl;

var KeyAdd = React.createClass({
  getInitialState: function() {
    return {key: ''};
  },

  handleChangeKey: function(e) {
		this.setState({key: e.target.value});
	},

  handleSaveClick: function(e) {
    e.preventDefault();
    var self = this;

    Keys.add(this.state.key).then(function(response) {
      var data = JSON.parse(response.body);
      if ((response.statusCode >= 300) || (!data.success)) {
        self.setState({error: data.message});
      } else {
        window.location = '/keys';
      }
    });
  },

	render: function() {
    var M = this.props.intl.formatMessage;

    return (
      <div>
        <Navigation active="keys" />

        <section className="row no-border">
          <h2>{M({id: 'new-public-key'})}</h2>
          <div className="twelve-col">
            <AlertBox message={this.state.error} />

            <form>
              <fieldset>
                <ul>
                  <li>
                    <label htmlFor="key">{M({id: 'public-key'})}:</label>
                    <textarea onChange={this.handleChangeKey} defaultValue={this.state.key}
                        placeholder={M({id: 'new-public-key-description'})}>
                    </textarea>
                  </li>
                </ul>
              </fieldset>
            </form>
            <div>
							<a href='/keys' onClick={this.handleSaveClick} className="button--primary">{M({id: 'save'})}</a>
							&nbsp;
							<a href='/keys' className="button--secondary">{M({id: 'cancel'})}</a>
						</div>
          </div>
        </section>

        <Footer />
      </div>
    );
	}
});

module.exports = injectIntl(KeyAdd);
