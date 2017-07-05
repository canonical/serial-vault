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
import TestUtils from 'react-addons-test-utils';
import {shallow, mount, render} from 'enzyme';

jest.dontMock('../components/ModelEdit');
jest.dontMock('../components/Navigation');
jest.dontMock('../components/Utils');

const token = { role: 200 }
const tokenUser = { role: 100 }

describe('model edit', function() {

    it('displays the model edit page for create', function() {
        var ModelEdit = require('../components/ModelEdit');

        // Mock the data retrieval from the API
        var getModel = jest.genMockFunction();
        var getKeypairs = jest.genMockFunction();
        ModelEdit.prototype.__reactAutoBindMap.getModel = getModel;
        ModelEdit.prototype.__reactAutoBindMap.getKeypairs = getKeypairs;
        window.AppState = {getLocale: function() {return 'en'}};

        // Render the component
        var modelPage = TestUtils.renderIntoDocument(
            <ModelEdit params={{}} token={token} />
        );

        expect(TestUtils.isCompositeComponent(modelPage)).toBeTruthy();

        // Check all the expected elements are rendered
        var section = TestUtils.findRenderedDOMComponentWithTag(modelPage, 'section');
        var h2 = TestUtils.findRenderedDOMComponentWithTag(modelPage, 'h2');
        expect(h2.textContent).toBe('New Model');

        // Check the data retrieval calls
        expect(getModel.mock.calls.length).toBe(0);
        expect(getKeypairs.mock.calls.length).toBe(1);

        // Check that the form is rendered without data
        var inputs = TestUtils.scryRenderedDOMComponentsWithTag(modelPage, 'input');
        expect(inputs.length).toBe(2);
        expect(inputs[0].value).toBe('');
        expect(inputs[1].value).toBe('');

    });

    it('displays the model edit page for an existing model', function() {
        var ModelEdit = require('../components/ModelEdit');

            // Mock the data retrieval from the API
        var getModel = jest.genMockFunction();
        var getKeypairs = jest.genMockFunction();
        var handleSaveClick = jest.genMockFunction();
        ModelEdit.prototype.__reactAutoBindMap.getModel = getModel;
        ModelEdit.prototype.__reactAutoBindMap.getKeypairs = getKeypairs;
        ModelEdit.prototype.__reactAutoBindMap.handleSaveClick = handleSaveClick;
        window.AppState = {getLocale: function() {return 'en'}};

        // Render the component
        var modelPage = TestUtils.renderIntoDocument(
            <ModelEdit params={{id: 1}} token={token} />
        );

        expect(TestUtils.isCompositeComponent(modelPage)).toBeTruthy();

        // Check the data retrieval calls
        //expect(getModel.mock.calls.length).toBe(1);
        expect(getKeypairs.mock.calls.length).toBe(1);

        // Get the save link
        var anchors = TestUtils.scryRenderedDOMComponentsWithTag(modelPage, 'a');
        expect(anchors.length).toBe(2);
        expect(anchors[1].textContent).toBe('Save');
        TestUtils.Simulate.click(anchors[1]);
        expect(handleSaveClick.mock.calls.length).toBe(1);
    });

    it('displays error with no permissions', function() {
        var ModelEdit = require('../components/ModelEdit');

        // Render the component
        const component = shallow(
            <ModelEdit />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

    it('displays error with insufficient permissions', function() {
        var ModelEdit = require('../components/ModelEdit');

        // Render the component
        const component = shallow(
            <ModelEdit token={tokenUser} />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

});
