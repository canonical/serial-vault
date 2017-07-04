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
import ReactDOM from 'react-dom'

import App from './App.js'
import {getAuthToken}      from './components/Utils'

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

    getAuthToken( (token) => {
      console.log('----', token)
      ReactDOM.render((
        <IntlProvider locale={locale} messages={Messages[locale]}>
          <App token={token} />
        </IntlProvider>
      ), this.container);
    })
  },

  unmount: function() {
    ReactDOM.unmountComponentAtNode(this.container);
  },

  rerender: function() {
    this.unmount();
    this.render();
  }
}

window.AppState.render()
