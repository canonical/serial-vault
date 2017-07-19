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
import Index from '../components/Index'

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

const token = { role: 200, name: 'Steven Vault' }

describe('index page', function() {
    it('displays the home page', function() {

        // Render the component
        const component = shallow(
            <Index token={token} />
        );

        expect(component.find('section')).toHaveLength(2)
        expect(component.find('AlertBox')).toHaveLength(0)
    })

    it('displays the home page with an er', function() {

        // Render the component
        const component = shallow(
            <Index token={token} error={'This is an error message'} />
        );

        expect(component.find('section')).toHaveLength(2)
        expect(component.find('AlertBox')).toHaveLength(1)
    })
})
