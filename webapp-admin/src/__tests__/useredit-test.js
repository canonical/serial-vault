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
'use strict'

import React from 'react';
import ReactDOM from 'react-dom';
import ReactTestUtils from 'react-dom/test-utils';
import {shallow, mount, render} from 'enzyme';
import UserEdit from '../components/UserEdit';

jest.dontMock('../components/UserEdit');
jest.dontMock('../components/Navigation');
jest.dontMock('../components/Utils');

const token = { role: 300 }
const tokenUser = { role: 100 }

window.AppState = {getLocale: function() {return 'en'}};

describe('user edit', function() {

    it('displays the user edit page for create', function() {

        // Mock the data retrieval from the API
        var getUser = jest.genMockFunction();
        var getNonUserAccounts = jest.genMockFunction();
        var getAllAccounts = jest.genMockFunction();
        UserEdit.prototype.getUser = getUser;
        UserEdit.prototype.getNonUserAccounts = getNonUserAccounts;
        UserEdit.prototype.getAllAccounts = getAllAccounts;

        // Render the component
        var userPage = ReactTestUtils.renderIntoDocument(
            <UserEdit params={{}} token={token} />
        );

        expect(ReactTestUtils.isCompositeComponent(userPage)).toBeTruthy();

        // Check all the expected elements are rendered
        var section = ReactTestUtils.findRenderedDOMComponentWithTag(userPage, 'section');
        var h2 = ReactTestUtils.findRenderedDOMComponentWithTag(userPage, 'h2');
        expect(h2.textContent).toBe('New User');

        // Check that the form is rendered without data
        var inputs = ReactTestUtils.scryRenderedDOMComponentsWithTag(userPage, 'input');
        expect(inputs.length).toBe(3);
        expect(inputs[0].value).toBe('');
        expect(inputs[1].value).toBe('');
        expect(inputs[2].value).toBe('');

        var selects = ReactTestUtils.scryRenderedDOMComponentsWithTag(userPage, 'select');
        expect(selects.length).toBe(3);
        expect(selects[0].value).toBe('');
        expect(selects[1].value).toBe('');
        expect(selects[2].value).toBe('');
    });

    it('displays the edit page for an existing user', function() {

        // Mock the data retrieval from the API
        // Mock the data retrieval from the API
        var getUser = jest.genMockFunction();
        var getNonUserAccounts = jest.genMockFunction();
        var getAllAccounts = jest.genMockFunction();
        var handleSaveClick = jest.genMockFunction();
        UserEdit.prototype.getUser = getUser;
        UserEdit.prototype.getNonUserAccounts = getNonUserAccounts;
        UserEdit.prototype.getAllAccounts = getAllAccounts;
        UserEdit.prototype.handleSaveClick = handleSaveClick;

        // Render the component
        var userPage = ReactTestUtils.renderIntoDocument(
            <UserEdit params={{id: 1}} token={token} />
        );

        expect(ReactTestUtils.isCompositeComponent(userPage)).toBeTruthy();
    });

    it('displays error with no permissions', function() {

        // Render the component
        const component = shallow(
            <UserEdit />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

    it('displays error with insufficient permissions', function() {

        // Render the component
        const component = shallow(
            <UserEdit token={tokenUser} />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

});
