// -*- Mode: Go; indent-tabs-mode: t -*-

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

import React, {Component} from 'react';
import ModelRow from './ModelRow';
import AlertBox from './AlertBox';
import ModelAssertion from './ModelAssertion';
import Models from '../models/models';
import {T, isUserAdmin} from './Utils'

class ModelList extends Component {

  constructor(props) {

    super(props)
    this.state = {
      models: this.props.models || [],
      confirmDelete: null,
      message: null,
      showAssert: null,
    }
  }

  componentDidMount() {
    this.refresh();
  }

  refresh() {
    this.getModels();
  }

  handleRefresh = () => {
    this.refresh()
  }

  getModels() {
    Models.list().then((response) => {
      var data = JSON.parse(response.body);
      var message = "";
      if (!data.success) {
        message = data.message;
      }
      this.setState({models: data.models, message: message});
    });
  }

  formatError(data) {
    var message = T(data.error_code);
    if (data.error_subcode) {
      message += ': ' + T(data.error_subcode);
    } else if (data.message) {
      message += ': ' + data.message;
    }
    return message;
  }

  handleShowAssert = (e) => {
    e.preventDefault();
    var assertId = parseInt(e.target.getAttribute('data-key'), 10);
    if (this.state.showAssert === assertId) {
        this.setState({showAssert: null})
    } else {
        this.setState({showAssert: assertId})
    }
  }

  handleDelete = (e) => {
    e.preventDefault();
    this.setState({confirmDelete: parseInt(e.target.getAttribute('data-key'), 10)});
  }

  handleDeleteModel = (e) => {
    e.preventDefault();
    var models = this.state.models.filter((mdl) => {
      return mdl.id === this.state.confirmDelete;
    });
    if (models.length === 0) {
      return;
    }

    Models.delete(models[0]).then((response) => {
      var data = JSON.parse(response.body);
      if ((response.statusCode >= 300) || (!data.success)) {
        this.setState({message: this.formatError(data)});
      } else {
        window.location = '/models';
      }
    });
  }

  handleDeleteModelCancel = (e) => {
    e.preventDefault();
    this.setState({confirmDelete: null});
  }

  renderTable() {

    if (this.state.models.length > 0) {
      return (
        <table>
          <thead>
            <tr>
              <th></th><th>{T('brand')}</th><th>{T('model')}</th><th>{T('private-key-short')}</th><th>{T('private-key-user-short')}</th>
              <th className="small">{T('active')}</th>
              <th>{T('private-key-model-short')}</th>
            </tr>
          </thead>
            {this.state.models.map((mdl) => {
              return (
                <tbody>
                  <ModelRow key={mdl.id} model={mdl} delete={this.handleDelete} confirmDelete={this.state.confirmDelete}
                      deleteModel={this.handleDeleteModel} cancelDelete={this.handleDeleteModelCancel} showAssert={this.handleShowAssert} />

                  {this.state.showAssert === mdl.id ? <ModelAssertion model={mdl} token={this.props.token} cancel={this.handleShowAssert} /> : ''}
                </tbody>
              );
            })}
        </table>
      );
    } else {
      return (
        <p>No models found.</p>
      );
    }
  }

  render() {
    if (!isUserAdmin(this.props.token)) {
      return (
        <div className="row">
          <AlertBox message={T('error-no-permissions')} />
        </div>
      )
    }

    return (
        <div className="row">

          <section className="row">
            <div className="u-equal-height">
              <h2 className="col-3">{T('models')}</h2>
              &nbsp;
              <div className="col-1"><a href="/models/new" className="p-button--brand" title={T('add-new-model')}>
                <i className="fa fa-plus"></i>
              </a></div>
            </div>
            <div className="col-12">
              <p>{T('models_available')}:</p>
            </div>
            <div className="col-12">
              <AlertBox message={this.state.message} />
            </div>
            <div className="col-12">
              {this.renderTable()}
            </div>
          </section>

        </div>
    );
  }
}

export default ModelList;
