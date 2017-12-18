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
import Keypairs from '../models/keypairs';
import {T} from './Utils'
import { setTimeout } from 'timers';


class KeypairStatus extends Component {

  constructor(props) {
    super(props)
    this.state = {
        keypairs: [],
    }

    this.getKeypairs()
  }

  poll = () => {
      // Polls every 30s
      setTimeout(this.getKeypairs.bind(this), 30000);
  }

  getKeypairs() {
    Keypairs.status().then((response) => {
        var data = JSON.parse(response.body);
        console.log(data)
        var message = "";
        if (!data.success) {
            message = data.message;
        }
        this.setState({keypairs: data.status, message: message});
    })
    .done( ()=> {
        this.poll()
    })
  }

  renderRow(keypr) {
    return (
      <tr key={keypr.id}>
        <td className="overflow" title={keypr['authority-id']}>{keypr['authority-id']}</td>
        <td className="overflow" title={keypr['key-name']}>{keypr['key-name']}</td>
        <td className="overflow" title={keypr.status}>{T(keypr.status)}</td>
      </tr>
    );
  }

  render() {

    if (this.state.keypairs.length > 0) {
      return (
          <div className="p-card--highlighted spacer">
            <table>
            <tbody>
                {this.state.keypairs.map((keypr) => {
                return this.renderRow(keypr);
                })}
            </tbody>
            </table>
          </div>
      );
    } else {
        return <span />;
    }
  }

}

export default KeypairStatus;
