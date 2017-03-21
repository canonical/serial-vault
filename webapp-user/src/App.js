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
import Models from './models/models'
import Assertion from './models/assertion'
import './sass/App.css';

class App extends Component {

  constructor(props) {
    super(props)
    this.state = {
      models: []
    }

    this.getModels()
  }

  getModels() {
    Models.list().then((response) => {
      //var data = JSON.parse(response.body);
      var data = response.data;
      var message = "";
      if (!data.success) {
        message = data.message;
      }
      this.setState({models: data.models, message: message});
    })
  }

  onSubmitForm(form) {
    console.log(form)
    Assertion.create(form).then((response) => {
      var data = response.data;
      var message = "";
      if (!data.success) {
        message = data.message;
      }
      console.log(data)
      console.log(message)
    })
  }

  render() {
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
          <h3>Create System User Assertion</h3>

          <SystemUserForm onSubmit={this.onSubmitForm} models={this.state.models} />

        </div>
      </div>
    );
  }
}

export default App;
