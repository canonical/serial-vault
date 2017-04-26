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
import React from 'react'
import {Router} from 'react-router'
import ReactDOM from 'react-dom'
import {Route} from 'react-router'
import {IndexRoute} from 'react-router'
import {browserHistory} from 'react-router'
import App        from './components/App'
import Index      from './components/Index'
import ModelList  from './components/ModelList'
import ModelEdit  from './components/ModelEdit'
import KeypairAdd from './components/KeypairAdd'
import SigningLog from './components/SigningLog'
import AccountList from './components/AccountList';

// Imports for i18n
import {IntlProvider, addLocaleData} from 'react-intl';

// Translated messages
var Messages = require('./components/messages');

window.AppState = {
  container: document.getElementById("main"),

  getLocale: function() {
    return localStorage.getItem('locale') || 'en';
  },

  setLocale: function(lang) {
    localStorage.setItem('locale', lang);
  },

  render: function() {
    var locale = this.getLocale();

    ReactDOM.render((
      <IntlProvider locale={locale} messages={Messages[locale]}>
        <Router history={browserHistory}>
          <Route path="/" component={App}>
            <IndexRoute component={Index} />
            <Route path="models" component={ModelList} />
            <Route path="models/new" component={ModelEdit} />
            <Route path="models/:id/edit" component={ModelEdit} />
            <Route path="models/keypairs/new" component={KeypairAdd} />
            <Route path="accounts" component={AccountList} />
            <Route path="signinglog" component={SigningLog} />
            <Route path="*" component={Index} />
          </Route>
        </Router>
      </IntlProvider>
    ), this.container);
  },

  unmount: function() {
    ReactDOM.unmountComponentAtNode(this.container);
  },

  rerender: function() {
    this.unmount();
    this.render();
  }
}

window.AppState.render();
