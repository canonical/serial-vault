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
import ModelEdit from '../components/ModelEdit';

jest.dontMock('../components/ModelEdit');
jest.dontMock('../components/Navigation');
jest.dontMock('../components/Utils');

const token = { role: 200 }
const tokenUser = { role: 100 }

window.AppState = {getLocale: function() {return 'en'}};

describe('model edit', function() {

    it('displays the model edit page for create', function() {

        // Mock the data retrieval from the API
        var getModel = jest.genMockFunction();
        var getKeypairs = jest.genMockFunction();
        ModelEdit.prototype.getModel = getModel;
        ModelEdit.prototype.getKeypairs = getKeypairs;

        // Render the component
        var modelPage = ReactTestUtils.renderIntoDocument(
            <ModelEdit params={{}} token={token} />
        );

        expect(ReactTestUtils.isCompositeComponent(modelPage)).toBeTruthy();

        // Check all the expected elements are rendered
        var section = ReactTestUtils.findRenderedDOMComponentWithTag(modelPage, 'section');
        var h2 = ReactTestUtils.findRenderedDOMComponentWithTag(modelPage, 'h2');
        expect(h2.textContent).toBe('New Model');

        // Check that the form is rendered without data
        var inputs = ReactTestUtils.scryRenderedDOMComponentsWithTag(modelPage, 'input');
        expect(inputs.length).toBe(3);
        expect(inputs[0].value).toBe('');
        expect(inputs[1].value).toBe('');

    });

    it('displays the model edit page for an existing model', function() {

        // Mock the data retrieval from the API
        var getModel = jest.genMockFunction();
        var getKeypairs = jest.genMockFunction();
        var handleSaveClick = jest.genMockFunction();
        ModelEdit.prototype.getModel = getModel;
        ModelEdit.prototype.getKeypairs = getKeypairs;
        ModelEdit.prototype.handleSaveClick = handleSaveClick;

        // Render the component
        var modelPage = ReactTestUtils.renderIntoDocument(
            <ModelEdit params={{id: 1}} token={token} />
        );

        expect(ReactTestUtils.isCompositeComponent(modelPage)).toBeTruthy();
    });

    it('displays error with no permissions', function() {

        // Render the component
        const component = shallow(
            <ModelEdit />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

    it('displays error with insufficient permissions', function() {

        // Render the component
        const component = shallow(
            <ModelEdit token={tokenUser} />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

});
