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

import React from 'react';
import Adapter from 'enzyme-adapter-react-16';
import {shallow, configure} from 'enzyme';
import ReactTestUtils from 'react-dom/test-utils';
import AccountForm from '../components/AccountForm'

configure({ adapter: new Adapter() });

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

const tokenSuperAdmin = { role: 300 }
const tokenAdmin      = { role: 200 }
const tokenUser       = { role: 100 }

describe('account form', function() {
    it('displays the new account form', function() {

        // Render the component
        const component = shallow(
            <AccountForm token={tokenSuperAdmin} />
        );

        expect(component.find('section')).toHaveLength(1)
        expect(component.find('input')).toHaveLength(1)
    })

    it('displays error with no permissions', function() {

        // Render the component
        const component = shallow(
            <AccountForm />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

    it('displays error with insufficient permissions for user', function() {

        // Render the component
        const component = shallow(
            <AccountForm token={tokenUser} />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

    it('displays error with insufficient permissions for admin', function() {
        // Render the component
        const component = shallow(
            <AccountForm token={tokenAdmin} />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })
})
