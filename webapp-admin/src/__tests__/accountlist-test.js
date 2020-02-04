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

import React from 'react'
import Adapter from 'enzyme-adapter-react-16';
import {shallow, configure} from 'enzyme';
import AccountList from '../components/AccountList'

configure({ adapter: new Adapter() });

jest.dontMock('../components/AccountList');

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

// Test fixtures
const account = {ID: 1, AuthorityID: "canonical", Assertion: "123456abcdef"}
const ACCOUNTS = [account]
const KEYPAIRS = [{ID: 1, AuthorityID: "canonical", KeyID: "123", Assertion: "123456abcdef"}]
const KEYPAIRS_INCOMPLETE = [{ID: 1, AuthorityID: "canonical", KeyID: "123", Assertion: null}]
const MODELS = [{id: 1, "authority-id-user": "canonical", "key-id-user": "123"}]
const MODELS_NOTUSED = [{id: 1, "authority-id-user": "canonical", "key-id-user": "456"}]

const token = { role: 200 }
const tokenUser = { role: 100 }

describe('accounts list', function() {

    it('displays the account lists', function() {

        // Render the component
        const component = shallow(
            // <AccountList token={this.props.token} selectedAccount={this.state.selectedAccount} keypairs={this.state.keypairs} />
            <AccountList token={token} selectedAccount={{}} keypairs={{}} />
        );

        expect(component.find('section')).toHaveLength(2)
        // No accounts or keys
        expect(component.find('table')).toHaveLength(0)
    })

    it('displays the keypairs with no accounts', function() {

        // Render the component
        const component = shallow(
            <AccountList selectedAccount={{}} keypairs={KEYPAIRS} models={MODELS} token={token} />
        );

        expect(component.find('section')).toHaveLength(2)
        expect(component.find('table')).toHaveLength(1)
        // No assertions displayed, one warning (+ two buttons)
        expect(component.find('pre')).toHaveLength(0)
        expect(component.find('i')).toHaveLength(5)
    })

    it('displays the accounts and keys with assertions', function() {

        // Render the component
        const component = shallow(
            <AccountList selectedAccount={account} keypairs={KEYPAIRS} models={MODELS} token={token} />
        );

        expect(component.find('section')).toHaveLength(2)
        expect(component.find('table')).toHaveLength(2)

        // Two assertions displayed
        expect(component.find('p')).toHaveLength(2)
        expect(component.find('p').get(0).props.title).toEqual("123456abcdef");
        expect(component.find('p').get(1).props.title).toEqual("123456abcdef");
    })

    it('displays the account and key with a `not used` message', function() {

        // Render the component
        const component = shallow(
            <AccountList selectedAccount={ACCOUNTS} keypairs={KEYPAIRS} models={MODELS_NOTUSED} token={token} />
        );

        expect(component.find('section')).toHaveLength(2)
        expect(component.find('table')).toHaveLength(1)
        // Only account assertion displayed, no key assertion and the key is not used for signing
        expect(component.find('p')).toHaveLength(2)
        expect(component.find('p').get(0).props.children).toEqual('No assertions found')
        expect(component.find('p').get(1).props.children).toEqual('Not used for signing system-user assertions')
    })

    it('displays the account and key with a warning message', function() {

        // Render the component
        const component = shallow(
            <AccountList selectedAccount={ACCOUNTS} keypairs={KEYPAIRS_INCOMPLETE} models={MODELS} token={token} />
        );

        expect(component.find('section')).toHaveLength(2)
        expect(component.find('table')).toHaveLength(1)
        // Only account assertion displayed, no key assertion and one warning (+ two buttons)
        expect(component.find('p')).toHaveLength(3)
        expect(component.find('i')).toHaveLength(6)
    })

    it('displays the key with a warning messages', function() {
        
        // Render the component
        const component = shallow(
            <AccountList selectedAccount={{}} keypairs={KEYPAIRS_INCOMPLETE} models={MODELS} token={token} />
        );

        expect(component.find('section')).toHaveLength(2)
        expect(component.find('table')).toHaveLength(1)
        // No account assertion displayed, no key assertion and two warnings (+ two buttons)
        expect(component.find('p')).toHaveLength(3)
        expect(component.find('i')).toHaveLength(6)
    })

    it('displays error with no permissions', function() {

        // Render the component
        const component = shallow(
            <AccountList />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

    it('displays error with insufficient permissions', function() {

        // Render the component
        const component = shallow(
            <AccountList token={tokenUser} />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

})