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
import React, {Component} from 'react'
import AlertBox from './AlertBox'
import Assertion from '../models/assertions'
import {T, isUserStandard} from './Utils'

function newUTCDate() {
    return new Date(new Date().toUTCString())
}

class SystemUserForm extends Component {

    constructor(props) {
        super(props);

    	const startDate = newUTCDate();
    	const endDate = new Date(newUTCDate().setFullYear(
            startDate.getFullYear() + 1
        ));
        this.state = {
            email: '',
            username: '',
            password: '',
            name: '',
            model: 0,
            since: startDate,
            until: endDate,
            message: '',
            assertion: null,
        }
    }

    submitAssertion(form) {
        Assertion.create(form).then((response) => {
            var data = JSON.parse(response.body);
            var message = null;
            if (!data.success) {
                message = data.message;
            }

            if (data.success) {
                this.setState({assertion: data.assertion, message: message})
            } else {
                this.setState({message: message})
            }
        })
    }

    downloadAssertion () {
        return 'data:application/octet-stream;charset=utf-8,' + encodeURIComponent(this.state.assertion)
    }

    onShowForm= (e) => {
        e.preventDefault()
        this.setState({assertion: null})
    }

    handleChangeEmail = (e) => {
        this.setState({email: e.target.value});
    }

    handleChangeUsername = (e) => {
        this.setState({username: e.target.value});
    }

    handleChangePassword = (e) => {
        this.setState({password: e.target.value});
    }

    handleChangeName = (e) => {
        this.setState({name: e.target.value});
    }

    handleChangeModel = (e) => {
        this.setState({model: parseInt(e.target.value, 10)});
    }

    handleChangeSinceDateTime = (e) => {
        var date = new Date(e.target.value);   
        this.setState({since: date});
    }

    handleChangeUntilDateTime = (e) => {
        var date = new Date(e.target.value);   
        this.setState({until: date});
    }

    onSubmit = (e) => {
        e.preventDefault()

        var form = {
            email: this.state.email,
            username: this.state.username,
            password: this.state.password,
            name:     this.state.name,
            model:    this.state.model,
            since:    this.state.since.toISOString(),
            until:    this.state.until.toISOString(),
        }
        if (this.validate(form)) {
            // this.props.onSubmit(form)
            this.submitAssertion(form)
        }
    }

    validate(form) {
        // Check the mandatory fields
        if ((!form.email) || (!form.username) || (!form.password) || (!form.name) || (!form.model) || (form.model === 0)) {
            this.setState({message: 'All the fields must be entered'});
            return false;
        }

        // Check the email
        if (! /^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$/.test(form.email)) {
            this.setState({message: 'The email is not valid'});
            return false
        }

        this.setState({message: ''})
        return true;
    }

    render() {

        if (!isUserStandard(this.props.token)) {
            return (
                <div className="row">
                <AlertBox message={T('error-no-permissions')} />
                </div>
            )
        }

        if (this.state.assertion) {
            return (
                <div className="row">
                    <a href="#form" onClick={this.onShowForm}>&laquo; Back to the form</a>
                    <h3>System User Assertion</h3>

                    <pre>{this.state.assertion}</pre>
                    <a className="p-button--brand" href={this.downloadAssertion()} download="auto-import.assert">Download</a>
                </div>
            )
        }

        return (
            <div className="row">

                <h2>{T('create-system-user')}</h2>

                <AlertBox message={this.state.message} type={'negative'} />

                <form>
                    <fieldset>
                        <label htmlFor="email">Email:
                            <input type="email" name="email" required placeholder="email address" onChange={this.handleChangeEmail} value={this.state.email} />
                        </label>
                        <label htmlFor="username">Username:
                            <input type="text" name="username" placeholder="system-user name" onChange={this.handleChangeUsername} value={this.state.username} />
                        </label>
                        <label htmlFor="password">Password:
                            <input type="password" name="password" placeholder="password for the system-user" onChange={this.handleChangePassword} value={this.state.password} />
                        </label>
                        <label htmlFor="name">Full Name:
                            <input type="text" name="name" placeholder="name of the user" onChange={this.handleChangeName} value={this.state.name} />
                        </label>
                        <label htmlFor="name">Model:
                            <select onChange={this.handleChangeModel} value={this.state.model.id}>
                                <option value={0}>--</option>
                            {this.props.models.map((m) => {
                                return (
                                    <option key={m.id} value={m.id}>{m['brand-id']} {m.model}</option>
                                )
                            })}
                            </select>
                        </label>
                        <label htmlFor="since">Since (UTC):
                            <div className="row">
                                <div className="col-3">
                                     <input type="datetime-local"
                                        name="since_date_time"
                                        value={this.state.since.toISOString().substr(0, 16)}
                                        onChange={this.handleChangeSinceDateTime} />
				                </div>
                            </div>
                        </label>
                        <label htmlFor="until">Until (UTC):
                            <div className="row">
                                <div className="col-3">
                                    <input type="datetime-local"
                                        name="until_date_time"
                                        value={this.state.until.toISOString().substr(0, 16)}
                                        onChange={this.handleChangeUntilDateTime} />
                                </div>
                            </div>
                        </label>
                    
                    </fieldset>
                    <button className="p-button--brand" onClick={this.onSubmit}>Create</button>
                </form>
            </div>
        )
    }
}

export default SystemUserForm
