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
var Router = require('react-router').Router;
var ReactDOM = require('react-dom');
var Route = require('react-router').Route;
var IndexRoute = require('react-router').IndexRoute;
var browserHistory = require('react-router').browserHistory;
var App = require('./components/App');
var Index = require('./components/Index');
var ModelList = require('./components/ModelList');
var ModelEdit = require('./components/ModelEdit');
var KeyList = require('./components/KeyList');
var KeyAdd = require('./components/KeyAdd');

// Imports for i18n
import {IntlProvider, addLocaleData} from 'react-intl';
import en from 'react-intl/lib/locale-data/en';
import zh from 'react-intl/lib/locale-data/zh';

// Translated messages
var Messages = require('./components/messages');

// Add the locales we need
addLocaleData(en);
addLocaleData(zh);

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
            <Route path="keys" component={KeyList} />
            <Route path="keys/new" component={KeyAdd} />
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
