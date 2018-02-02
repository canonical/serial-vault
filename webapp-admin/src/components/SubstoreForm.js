/*
 * Copyright (C) 2018-2019 Canonical Ltd
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
import {T, isUserAdmin} from './Utils';


class SubstoreForm extends Component {

    handleChangeSerial = (e) => {
        e.preventDefault()
        this.props.onChange('serialnumber', e.target.value)
    }

    handleChangeSubstore = (e) => {
        e.preventDefault()
        this.props.onChange('store', e.target.value)
    }

    handleChangeFromModel = (e) => {
        e.preventDefault()
        this.props.onChange('fromModelID', parseInt(e.target.value, 10))
    }

    handleChangeToModel = (e) => {
        e.preventDefault()
        this.props.onChange('toModelID', parseInt(e.target.value, 10))
    }

    render() {

        var b = this.props.substore

        return (
            <tr>
                <td colSpan="5">
                    <form>
                        <fieldset>
                            <label htmlFor="fromModel">{T('model')}:
                                <select value={b.fromModelID} id="fromModel" onChange={this.handleChangeFromModel}>
                                    <option></option>
                                    {this.props.models.map(function(m) {
                                        return <option key={m.id} value={m.id}>{m.model}</option>;
                                    })}
                                </select>
                            </label>

                            <label htmlFor="serial">{T('serial-number')}:
                                <input type="text" id="serial" placeholder={T('serial-number-description')}
                                    value={b.serialnumber} onChange={this.handleChangeSerial} />
                            </label>

                            <label htmlFor="substore">{T('substore')}:
                                <input type="text" id="substore" placeholder={T('substore-description')}
                                    value={b.store} onChange={this.handleChangeSubstore} />
                            </label>

                            <label htmlFor="toModel">{T('substore-model')}:
                                <select value={b.toModelID} id="toModel" onChange={this.handleChangeToModel}>
                                    <option></option>
                                    {this.props.models.map(function(m) {
                                        return <option key={m.id} value={m.id}>{m.model}</option>;
                                    })}
                                </select>
                            </label>
                        </fieldset>
                        {isUserAdmin(this.props.token) ?
                          <span>
                            <button className="p-button--neutral" onClick={this.props.onCancel} data-key={b.id}>{T('cancel')}</button>
                            <button className="p-button--brand" onClick={this.props.onSave} data-key={b.id}>{T('save')}</button>
                          </span>
                          : <button className="p-button--neutral" onClick={this.props.onCancel} data-key={b.id}>{T('close')}</button>
                        }
                    </form>
                </td>
            </tr>
        )
    }
}

export default SubstoreForm
