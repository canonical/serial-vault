// -*- Mode: Go; indent-tabs-mode: t -*-

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

import React, {Component} from 'react';
import KeypairList from './KeypairList';
import KeypairStatus from './KeypairStatus';
import AlertBox from './AlertBox';
import Keypairs from '../models/keypairs';
import {T, isUserAdmin} from './Utils'

class Keypair extends Component {

  constructor(props) {

    super(props)
    this.state = {
      keypairs: this.props.keypairs || [],
      confirmDelete: null,
      message: null,
    }
  }

  componentDidMount() {
    this.refresh();
  }

  refresh() {
    this.getKeypairs();
  }

  handleRefresh = () => {
    this.refresh()
  }

  getKeypairs() {
    Keypairs.list().then((response) => {
      var data = JSON.parse(response.body);
      var message = "";
      if (!data.success) {
        message = data.message;
      }
      this.setState({keypairs: data.keypairs, message: message});
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
            <div className="u-equal-height spacer">
              <h2 className="col-3">{T('signing-keys')}</h2>
              &nbsp;
              <div className="col-1">
                <a href="/signing-keys/new" className="p-button--brand" title={T('add-new-signing-key')}>
                  <i className="fa fa-plus"></i>
                </a>
              </div>
              <div className="col-1">
                <a href="/signing-keys/generate" className="p-button--brand" title={T('generate-signing-key')}>
                  <i className="fa fa-cog"></i>
                </a>
              </div>
            </div>
            <div className="col-12">
              <KeypairStatus token={this.props.token} />
            </div>
            <div className="col-12">
              <KeypairList keypairs={this.state.keypairs} refresh={this.handleRefresh} />
            </div>
          </section>

        </div>
    );
  }
}

export default Keypair;
