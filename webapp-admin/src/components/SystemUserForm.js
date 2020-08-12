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
import moment from 'moment'
import AlertBox from './AlertBox'
import Assertion from '../models/assertions'
import {T, isUserStandard} from './Utils'

import DatePicker from 'react-datepicker'
import "react-datepicker/dist/react-datepicker.css";

class SystemUserForm extends Component {

    constructor(props) {
        super(props);

        this.state = {
            email: '',
            username: '',
            password: '',
            name: '',
            model: 0,
            since: moment.utc(),
            until: moment.utc().add(1, 'year'),
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

    handleChangeSinceDate = (date) => {
        // Get the time from the current setting
        let d = moment(date)
        d.hours(this.state.since.hours())
        d.minutes(this.state.since.minutes())
        d.seconds(this.state.since.seconds())

        this.setState({since: d});
    }

    handleChangeSinceTime = (e) => {
        var time = e.target.value.split(':')
        var t = this.state.since
        t.hours(time[0])
        t.minutes(time[1])
        t.seconds(0)

        this.setState({since: t});
    }

    handleChangeUntilDate = (date) => {
        // Get the time from the current setting
        let d = moment(date)
        d.hours(this.state.until.hours())
        d.minutes(this.state.until.minutes())
        d.seconds(this.state.until.seconds())

        this.setState({until: d});
    }

    handleChangeUntilTime = (e) => {
        var time = e.target.value.split(':')
        var t = this.state.until
        t.hours(time[0])
        t.minutes(time[1])
        t.seconds(0)

        this.setState({until: t});
    }

    onSubmit = (e) => {
        e.preventDefault()

        var form = {
            email: this.state.email,
            username: this.state.username,
            password: this.state.password,
            name:     this.state.name,
            model:    this.state.model,
            since:    this.state.since.format('YYYY-MM-DDTHH:mm:ss') + 'Z',
            until:    this.state.until.format('YYYY-MM-DDTHH:mm:ss') + 'Z',
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
                                    <DatePicker selected={this.state.since.toDate()} onChange={this.handleChangeSinceDate} />
                                </div>
                                <div className="col-3">
                                    <input type="time" name="since_time" title="the time the assertion is valid from (UTC)" onChange={this.handleChangeSinceTime} value={this.state.since.format('HH:mm')} />
                                </div>
                            </div>
                        </label>
                        <label htmlFor="until">Until (UTC):
                            <div className="row">
                                <div className="col-3">
                                    <DatePicker selected={this.state.until.toDate()} onChange={this.handleChangeUntilDate} />
                                </div>
                                <div className="col-3">
                                    <input type="time" name="until_time" title="the time the assertion is valid until (UTC)" onChange={this.handleChangeUntilTime} value={this.state.until.format('HH:mm')} />
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
