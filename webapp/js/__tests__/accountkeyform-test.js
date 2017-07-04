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
import {shallow, mount, render} from 'enzyme';
import AccountKeyForm from '../components/AccountKeyForm'

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

const token = { role: 200 }
const tokenUser = { role: 100 }

describe('account key form', function() {
    it('displays the new account key form', function() {

        // Render the component
        const component = shallow(
            <AccountKeyForm token={token} />
        );

        expect(component.find('section')).toHaveLength(1)
        expect(component.find('input')).toHaveLength(1)
    })

    it('displays error with no permissions', function() {

        // Render the component
        const component = shallow(
            <AccountKeyForm />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

    it('displays error with insufficient permissions', function() {

        // Render the component
        const component = shallow(
            <AccountKeyForm token={tokenUser} />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })
})
