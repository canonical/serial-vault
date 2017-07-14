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

import React from 'react';
import {FormattedMessage} from 'react-intl'
import {T, isLoggedIn} from './Utils'
import AlertBox from './AlertBox'


var Index = React.createClass({

  getInitialState: function() {
    return {token: this.props.token}
  },

  renderUser: function() {
    if (isLoggedIn(this.props.token)) {
      return <div />
    } else {
      return (
        <div>
          <a href="/login" className="p-button--brand">{T('login')}</a>
        </div>
      )
    }
  },

  renderError: function() {
    if (this.props.error) {
      return (
        <AlertBox message={T('user-not-found')} />
      )
    }
  },

  render: function() {
    return (
        <div className="row">


          <section className="row">
            <h2><FormattedMessage id="title" /></h2>
            {this.renderError()}
            <div>
              <div className="p-card">
                <FormattedMessage id="description" />
              </div>
            </div>
          </section>

          <section className="row">
            {this.renderUser()}
          </section>
        </div>
    );
  }
});

export default Index;