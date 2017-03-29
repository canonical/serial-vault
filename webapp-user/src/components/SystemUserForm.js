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
import If from './If'


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
            message: '',
        }
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
        this.setState({model: parseInt(e.target.value)});
    }

    handleChangeSinceDate = (e) => {
        // Get the date from the form field
        var date = moment.utc(e.target.value)

        // Get the time from the current setting
        date.hours(this.state.since.hours())
        date.minutes(this.state.since.minutes())
        date.seconds(this.state.since.seconds())

        this.setState({since: date});
    }

    handleChangeSinceTime = (e) => {
        var time = e.target.value.split(':')
        var t = this.state.since
        t.hours(time[0])
        t.minutes(time[1])
        t.seconds(0)

        this.setState({since: t});
    }

    onSubmit = (e) => {
        e.preventDefault()

        var form = {
            email: this.state.email,
            username: this.state.username,
            password: this.state.password,
            name:     this.state.name,
            model:    this.state.model,
            since:    this.state.since.format('YYYY-MM-DDThh:mm:ss'),
        }
        if (this.validate(form)) {
            this.props.onSubmit(form)
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

        return (
            <div className="row">

                <If cond={this.state.message}>
                    <AlertBox message={this.state.message} type={'negative'} />
                </If>

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
                                <div className="col-6">
                                    <input type="date" name="since_date" title="the date the assertion is valid from (UTC)" onChange={this.handleChangeSinceDate} value={this.state.since.format('YYYY-MM-DD')} />
                                </div>
                                <div className="col-6">
                                    <input type="time" name="since_time" title="the time the assertion is valid from (UTC)" onChange={this.handleChangeSinceTime} value={this.state.since.format('hh:mm')} />
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
