// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
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
import UserRow from './UserRow';
import AlertBox from './AlertBox';
import Users from '../models/users';
import {T, isUserSuperuser} from './Utils'

class UserList extends Component {

  constructor(props) {
    super(props)
    this.state = {
      users: this.props.users || [],
      confirmDelete: null,
      message: null,
    }
  }

  componentDidMount() {
    this.refresh();
  }

  refresh() {
    this.getUsers();
  }

  handleRefresh = () => {
    this.refresh()
  }

  getUsers() {
    Users.list().then((response) => {
      var data = JSON.parse(response.body);
      var message = "";
      if (!data.success) {
        message = data.message;
      }
      this.setState({users: data.users, message: message});
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

  handleDelete = (e) => {
    e.preventDefault();
    this.setState({confirmDelete: parseInt(e.target.getAttribute('data-key'), 10)});
  }

  handleDeleteUser = (e) => {
    e.preventDefault();
    var users = this.state.users.filter((user) => {
      return user.ID === this.state.confirmDelete;
    });
    if (users.length === 0) {
      return;
    }

    Users.delete(users[0]).then((response) => {
      var data = JSON.parse(response.body);
      if ((response.statusCode >= 300) || (!data.success)) {
        this.setState({message: this.formatError(data)});
      } else {
        window.location = '/users';
      }
    });
  }

  handleDeleteUserCancel = (e) => {
    e.preventDefault();
    this.setState({confirmDelete: null});
  }

  renderTable() {

    if (this.state.users.length > 0) {
      return (
        <table>
          <thead>
            <tr>
              <th></th><th>{T('username')}</th><th>{T('name')}</th><th>{T('email')}</th><th>{T('role')}</th>
            </tr>
          </thead>
          <tbody>
            {this.state.users.map((user) => {
              return (
                <UserRow key={user.ID} user={user} delete={this.handleDelete} confirmDelete={this.state.confirmDelete}
                  deleteUser={this.handleDeleteUser} cancelDelete={this.handleDeleteUserCancel} />
              );
            })}
          </tbody>
        </table>
      );
    } else {
      return (
        <p>No users found.</p>
      );
    }
  }

  render() {
    if (!isUserSuperuser(this.props.token)) {
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
              <h2 className="col-3">{T('users')}</h2>
              &nbsp;
              <div className="col-1"><a href="/users/new" className="p-button--brand" title={T('add-new-user')}>
                <i className="fa fa-plus"></i>
              </a></div>
            </div>
            <div className="col-12">
              <p>{T('users_available')}:</p>
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

export default UserList;
