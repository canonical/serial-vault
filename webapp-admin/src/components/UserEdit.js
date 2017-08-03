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
import AlertBox from './AlertBox';
import Users from '../models/users';
import {T, isUserSuperuser} from './Utils';

class UserEdit extends Component {

    constructor(props) {
        super(props)
        this.state = {
            title: null,
            user: {},
            error: null,
            // TODO temporary move user.Accounts to userAccounts, as backend provides accounts for the user 
            // that way. In future this will be get in an independant call.
            userAccountsOptions: [],
            nonUserAccountsOptions: [],
        }
    }

    componentDidMount() {
        if (this.props.id) {
            this.setTitle('edit-user');
            this.getUser(this.props.id);
        } else {
            this.setTitle('new-user');
        }
    }

    setTitle(title) {
        this.setState({title: T(title)});
    }

    getUser(userId) {
        var self = this;
        Users.get(userId).then(function(response) {
            var data = JSON.parse(response.body);
            self.setState({
                user: data.user, 
                userAccountsOptions: self.buildSelectBoxOptions(data.user.Accounts)});
        });
    }

    buildSelectBoxOptions(accounts) {
        var options = [];
        for (var i = 0; i < accounts.length; ++i) {
            options.push(this.buildSelectBoxOption(accounts[i].AuthorityID))
        }
        return options
    }

    buildSelectBoxOption = (val) => {
        return <option key={val} value={val}>{val}</option>; 
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

    handleChangeUsername = (e) => {
        var user = this.state.user;
        user.Username = e.target.value;
        this.setState({user: user});
    }

    handleChangeName = (e) => {
        var user = this.state.user;
        user.Name = e.target.value;
        this.setState({user: user});
    }

    handleChangeEmail = (e) => {
        var user = this.state.user;
        user.Email = e.target.value;
        this.setState({user: user});
    }

    handleChangeRole = (e) => {
        var user = this.state.user;
        user.Role = parseInt(e.target.value, 10);
        this.setState({user: user});
    }

    handleAddAccountClick = (e) => {
        e.preventDefault();

        this.moveSelectBoxesOption(
            this.state.nonUserAccountsOptions, 
            this.state.userAccountsOptions, 
            this.getSelectBoxSelectedIndex(this.refs.nonUserAccounts));
        
        this.forceUpdate();
    }

    handleRemoveAccountClick = (e) => {
        e.preventDefault();

        this.moveSelectBoxesOption(
            this.state.userAccountsOptions, 
            this.state.nonUserAccountsOptions, 
            this.getSelectBoxSelectedIndex(this.refs.userAccounts));

        this.forceUpdate();
    }

    moveSelectBoxesOption = (src, dst, index) => {
        if (index == -1) {
            return
        }
        this.addOption(dst, src[index]);
        this.removeOption(src, index);
    }

    getSelectBoxSelectedIndex = (selectBox) => {
        return selectBox.selectedIndex;
    }

    addOption = (options, val) => {
        options.push(
            <option key={val} value={val}>{val}</option>
        );
    }

    removeOption = (options, index) => {
        if (index > -1) {
            options.splice(index, 1);
        }
    }

    handleSaveClick = (e) => {
        e.preventDefault();
        var self = this;

        if (this.state.user.id) {
            // Update the existing user
            Users.update(this.state.user).then(function(response) {
                var data = JSON.parse(response.body);
                if (response.statusCode >= 300) {
                    self.setState({error: self.formatError(data)});
                } else {
                    window.location = '/users';
                }
            });
        } else {
            // Create a new user
            Users.create(this.state.user).then(function(response) {
                var data = JSON.parse(response.body);
                if (response.statusCode >= 300) {
                    self.setState({error: self.formatError(data)});
                } else {
                    window.location = '/users';
                }
            });
        }
    }

    renderError() {
        if (this.state.error) {
            return (
                <AlertBox message={this.state.error} />
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
                      <h2>{this.state.title}</h2>

                        <AlertBox message={this.state.error} />

                        <form>
                            <fieldset>
                                <label htmlFor="username">{T('username')}:
                                    <input type="text" id="username" placeholder={T('user-username')}
                                        value={this.state.user.Username} onChange={this.handleChangeUsername} />
                                </label>
                                <label htmlFor="name">{T('name')}:
                                    <input type="text" id="name" placeholder={T('user-name')}
                                        value={this.state.user.Name} onChange={this.handleChangeName}/>
                                </label>
                                <label htmlFor="email">{T('email')}:
                                    <input type="text" id="email" placeholder={T('user-email')}
                                        value={this.state.user.Email} onChange={this.handleChangeEmail}/>
                                </label>
                                <label htmlFor="role">{T('role')}:
                                    <select value={this.state.user.Role} id="role" onChange={this.handleChangeRole}>
                                        <option></option>
                                        <option key="standard" value="100">Standard</option>
                                        <option key="admin" value="200">Admin</option>
                                        <option key="superuser" value="300">Superuser</option>
                                    </select>
                                </label>
                                <label htmlFor="accounts">{T('accounts')}:
                                    <select multiple id="accounts" ref="userAccounts">{this.state.userAccountsOptions}</select>
                                </label> 
                                <button onClick={this.handleAddAccountClick} className="p-button--positive">{T('↑ add ↑')}</button>
                                &nbsp;
                                <button onClick={this.handleRemoveAccountClick} className="p-button--negative">{T('↓ remove ↓')}</button>
                                <label htmlFor="otherAccounts">{T('other-accounts')}:
                                    <select multiple id="other-accounts" ref="nonUserAccounts">{this.state.nonUserAccountsOptions}</select>
                                </label> 
                            </fieldset>
                        </form>

                        <div>
                            <a href='/users' className="p-button--neutral">{T('cancel')}</a>
                            &nbsp;
                            <a href='/users' onClick={this.handleSaveClick} className="p-button--brand">{T('save')}</a>
                        </div>
                </section>
                <br />
            </div>
        )
    }
}

export default UserEdit;
