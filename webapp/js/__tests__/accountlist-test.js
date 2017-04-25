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
import {shallow, mount, render} from 'enzyme';

import AccountList from '../components/AccountList'

jest.dontMock('../components/AccountList');

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

// Test fixtures
const ACCOUNTS = [{ID: 1, AuthorityID: "canonical", Assertion: "123456abcdef"}]
const KEYPAIRS = [{ID: 1, AuthorityID: "canonical", KeyID: "123", Assertion: "123456abcdef"}]
const KEYPAIRS_INCOMPLETE = [{ID: 1, AuthorityID: "canonical", KeyID: "123", Assertion: null}]
const MODELS = [{id: 1, "authority-id-user": "canonical", "key-id-user": "123"}]
const MODELS_NOTUSED = [{id: 1, "authority-id-user": "canonical", "key-id-user": "456"}]

describe('accounts list', function() {
    it('displays the account lists', function() {

        // Render the component
        const component = shallow(
            <AccountList />
        );

        expect(component.find('section')).toHaveLength(2)
        // No accounts or keys
        expect(component.find('table')).toHaveLength(0)
    })

    it('displays the accounts and keys with assertions', function() {

        // Render the component
        const component = shallow(
            <AccountList accounts={ACCOUNTS} keypairs={KEYPAIRS} models={MODELS} />
        );

        expect(component.find('section')).toHaveLength(2)
        expect(component.find('table')).toHaveLength(2)
        // Two assertions displayed
        expect(component.find('pre')).toHaveLength(2)
    })

    it('displays the account and key with a `not used` message', function() {

        // Render the component
        const component = shallow(
            <AccountList accounts={ACCOUNTS} keypairs={KEYPAIRS} models={MODELS_NOTUSED} />
        );

        expect(component.find('section')).toHaveLength(2)
        expect(component.find('table')).toHaveLength(2)
        // Only account assertion displayed, no key assertion and the key is not used for signing
        expect(component.find('pre')).toHaveLength(1)
        expect(component.contains(<p>Not used for signing system-user assertions</p>)).toEqual(true)
    })

    it('displays the account and key with a warning message', function() {

        // Render the component
        const component = shallow(
            <AccountList accounts={ACCOUNTS} keypairs={KEYPAIRS_INCOMPLETE} models={MODELS} />
        );

        expect(component.find('section')).toHaveLength(2)
        expect(component.find('table')).toHaveLength(2)
        // Only account assertion displayed, no key assertion and one warning
        expect(component.find('pre')).toHaveLength(1)
        expect(component.find('i')).toHaveLength(2)
    })

    it('displays the key with a warning messages', function() {

        // Render the component
        const component = shallow(
            <AccountList keypairs={KEYPAIRS_INCOMPLETE} models={MODELS} />
        );

        expect(component.find('section')).toHaveLength(2)
        expect(component.find('table')).toHaveLength(1)
        // No account assertion displayed, no key assertion and two warnings
        expect(component.find('pre')).toHaveLength(0)
        expect(component.find('i')).toHaveLength(3)
    })

})