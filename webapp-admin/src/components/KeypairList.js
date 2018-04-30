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


class KeypairList extends Component {

  handleDeactivate = (e) => {
    Keypairs.disable(e.target.getAttribute('data-key')).then(this.props.refresh);
  }

  handleActivate = (e) => {
    Keypairs.enable(e.target.getAttribute('data-key')).then(this.props.refresh);
  }

  renderRow(keypr) {
    return (
      <tr key={keypr.ID}>
        <td className="small">
          <a href={'/signing-keys/'+keypr.ID} className="p-button--brand small" title={T('edit-keypair')}>
                <i className="fa fa-edit"></i>
          </a>

          {keypr.Active ? <button data-key={keypr.ID} onClick={this.handleDeactivate} className="p-button--neutral small">{T('deactivate')}</button> : <button data-key={keypr.ID} onClick={this.handleActivate} className="p-button--neutral small">{T('activate')}</button>}
        </td>
        <td className="overflow" title={keypr.AuthorityID}>{keypr.AuthorityID}</td>
        <td className="overflow" title={keypr.KeyID}>{keypr.KeyID}</td>
        <td>{keypr.Active ? <i className="fa fa-check"></i> :  <i className="fa fa-times"></i>}</td>
        <td className="overflow" title={keypr.KeyName}>{keypr.KeyName}</td>
      </tr>
    );
  }

  render() {

    if (this.props.keypairs.length > 0) {
      return (
        <table>
          <thead>
            <tr>
              <th /><th>{T('authority-id')}</th><th>{T('key-id')}</th><th className="small">{T('active')}</th>
              <th>{T('key-name')}</th>
            </tr>
          </thead>
          <tbody>
            {this.props.keypairs.map((keypr) => {
              return this.renderRow(keypr);
            })}
          </tbody>
        </table>
      );
    } else {
      return (
        <p>{T('no-signing-keys-found')}</p>
      );
    }
  }

}

export default KeypairList;
