/*
 * Copyright (C) 2017-2018 Canonical Ltd
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
import React, { Component } from 'react';
import Header from './components/Header'
import SystemUserForm from './components/SystemUserForm'
import If from './components/If'
import Models from './models/models'
import Assertion from './models/assertion'
import {sectionFromHash} from './models/Utils'
import createHistory from 'history/createBrowserHistory'
import './sass/App.css';

const history = createHistory()

class App extends Component {

  constructor(props) {
    super(props)
    this.state = {
      models: [],
      assertion: '',
      location: history.location,
    }

    history.listen(this.handleNavigation.bind(this))

    this.getModels()
  }

  handleNavigation(location) {
    this.setState({ location: location })
    window.scrollTo(0, 0)
  }

  getModels() {
    Models.list().then((response) => {
      var data = response.data;
      var message = "";
      if (!data.success) {
        message = data.message;
      }
      this.setState({models: data.models, message: message});
    })
  }

  onSubmitForm = (form) => {
    Assertion.create(form).then((response) => {
      var data = response.data;
      var message = "";
      if (!data.success) {
        message = data.message;
      }
      console.log(data)
      console.log(message)

      if (data.success) {
        this.setState({assertion: data.assertion})
        location = '#assertion';
      }

    })
  }

  downloadAssertion () {
    return 'data:application/octet-stream;charset=utf-8,' + encodeURIComponent(this.state.assertion)
  }

  render() {

    var currentSection = sectionFromHash(location.hash);

    return (
      <div className="App">
        <Header />
        <div className="row">
          <h2>System User Assertion</h2>
          <div className="p-card">
            <h3 className="p-card__title">Usage</h3>
            <p className="p-card__content">
              A system-user assertion allows a local system user to be created on specified devices.<br />
              Once the assertion is created, download the file to a USB drive and insert it into the device.
            </p>
          </div>

          <br />
          <If cond={currentSection==='home'}>
            <h3>Create System User Assertion</h3>

            <SystemUserForm onSubmit={this.onSubmitForm} models={this.state.models} />
          </If>

          <If cond={currentSection==='assertion'}>
            <a href="#">&laquo; Back to the form</a>
            <h3>System User Assertion</h3>

            <pre>{this.state.assertion}</pre>
            <a className="p-button--brand" href={this.downloadAssertion()} download="auto-import.assert">Download</a>
          </If>

        </div>
      </div>
    );
  }
}

export default App;
