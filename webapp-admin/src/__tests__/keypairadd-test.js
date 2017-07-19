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
import {shallow, mount, render} from 'enzyme';
import ReactTestUtils from 'react-dom/test-utils';
import { createRenderer } from 'react-test-renderer/shallow';
import KeypairAdd from '../components/KeypairAdd';

jest.dontMock('../components/KeypairAdd');
jest.dontMock('../components/Navigation');
jest.dontMock('../components/AlertBox');
jest.dontMock('../components/Utils');

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

var Messages = require('../components/messages').en;

const token = { role: 200 }
const tokenUser = { role: 100 }

describe('keypair add', function() {
    it('displays the new keypair page', function() {

        // Render the component
        var keysPage = ReactTestUtils.renderIntoDocument(
             <KeypairAdd token={token} />
        );

        expect(ReactTestUtils.isCompositeComponent(keysPage)).toBeTruthy();

        // Check all the expected elements are rendered
        var section = ReactTestUtils.findRenderedDOMComponentWithTag(keysPage, 'section');
        var h2 = ReactTestUtils.findRenderedDOMComponentWithTag(keysPage, 'h2');
    });

    it('stores updates to the form', function() {

        // Mock the onChange handler
        var handleChangeKey = jest.genMockFunction();
        var handleChangeAuthorityId = jest.genMockFunction();
        KeypairAdd.prototype.handleChangeKey = handleChangeKey;
        KeypairAdd.prototype.handleChangeAuthorityId = handleChangeAuthorityId;

        // Render the component
        var keysPage = ReactTestUtils.renderIntoDocument(
             <KeypairAdd token={token} />
        );

        expect(ReactTestUtils.isCompositeComponent(keysPage)).toBeTruthy();

        // Get the text box and update it
        var textarea = ReactTestUtils.findRenderedDOMComponentWithTag(keysPage, 'textarea');
        textarea.defaultValue = 'sushi-on-toast';
        ReactTestUtils.Simulate.change(textarea);

        // Get the AuthorityID field and update it
        var inputs = ReactTestUtils.scryRenderedDOMComponentsWithTag(keysPage, 'input');
        var textAuthority = inputs[0];
        textAuthority.value = 'sushi-on-rye';
        ReactTestUtils.Simulate.change(textAuthority);
    });

    it('displays the alert box on error', function() {

        var shallowRenderer = createRenderer();

        // Render the component
        shallowRenderer.render(
            <KeypairAdd error={'Critical: run out of sushi'} token={token} />
        );
        var keysPage = shallowRenderer.getRenderOutput();

        var section = keysPage.props.children[0];
        expect(section.props.children.length).toBe(2);
        expect(section.props.children[1].props.children[0].props.message).toBe('Critical: run out of sushi');
    });

    it('displays error with no permissions', function() {

        // Render the component
        const component = shallow(
            <KeypairAdd />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

    it('displays error with insufficient permissions', function() {

        // Render the component
        const component = shallow(
            <KeypairAdd token={tokenUser} />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })
});
