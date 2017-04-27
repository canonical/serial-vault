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
import  AlertBox from './AlertBox'
import Accounts from '../models/accounts'
import {T, parseResponse, formatError} from './Utils';

class AccountForm extends Component {

    constructor(props) {
        super(props);

        this.state = {
            assertion: null,
            error: null
        }

    }

    handleFileUpload = (e) => {
        var reader = new FileReader();
        var file = e.target.files[0];

        reader.onload = (upload) => {
            this.setState({
                assertion: upload.target.result.split(',')[1],
            });
        }

        reader.readAsDataURL(file);
    }

    handleSaveClick = (e) => {
        e.preventDefault();

        Accounts.create(this.state.assertion).then((response) => {
            var data = parseResponse(response)
            if (!data.success) {
                this.setState({error: formatError(data)});
            } else {
                window.location = '/accounts';
            }
        });
    }

    render() {
        return (
            <div>

                <section className="row no-border">
                    <h2>{T('new-account-assertion')}</h2>
                    <div className="col-12">
                        <AlertBox message={this.state.error} />

                        <form>
                            <fieldset>
                                <label htmlFor="key">{T('assertion')}:
                                    <input type="file" onChange={this.handleFileUpload} />
                                </label>
                            </fieldset>
                        </form>
                        <div>
                            <a href='/accounts' className="p-button--neutral">{T('cancel')}</a>
                            &nbsp;
                            <a href='/accounts' onClick={this.handleSaveClick} className="p-button--brand">{T('save')}</a>
                        </div>
                    </div>
                </section>
                <br />
            </div>
        );
    }

}

export default AccountForm