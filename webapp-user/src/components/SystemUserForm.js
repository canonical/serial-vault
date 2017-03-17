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


class SystemUserForm extends Component {

    constructor(props) {
        super(props);

        this.state = {
            email: '',
            username: '',
            password: '',
            name: '',
            model: '',
            since: moment.utc(),
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
        this.setState({mode: e.target.value});
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
        this.props.onSubmit(form)
    }

    render() {

        return (
            <div className="row">
                <form>
                    <fieldset>
                        <label htmlFor="email">Email:
                            <input type="text" name="email" placeholder="email address" onChange={this.handleChangeEmail} value={this.state.email} />
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
                            <input type="text" name="model" placeholder="FIX: select the model" onChange={this.handleChangeModel} value={this.state.model} />
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
