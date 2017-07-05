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

jest.dontMock('../components/KeypairAdd');
jest.dontMock('../components/Navigation');
jest.dontMock('../components/AlertBox');
jest.dontMock('../components/Utils');

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

const token = { role: 200 }
const tokenUser = { role: 100 }

describe('keypair add', function() {
    it('displays the new keypair page', function() {
        var Messages = require('../components/messages').en;
        var KeypairAdd = require('../components/KeypairAdd');

        // Render the component
        var keysPage = TestUtils.renderIntoDocument(
             <KeypairAdd token={token} />
        );

        expect(TestUtils.isCompositeComponent(keysPage)).toBeTruthy();

        // Check all the expected elements are rendered
        var section = TestUtils.findRenderedDOMComponentWithTag(keysPage, 'section');
        var h2 = TestUtils.findRenderedDOMComponentWithTag(keysPage, 'h2');
    });

    it('stores updates to the form', function() {
        var Messages = require('../components/messages').en;
        var KeypairAdd = require('../components/KeypairAdd');

        // Mock the onChange handler
        var handleChangeKey = jest.genMockFunction();
        var handleChangeAuthorityId = jest.genMockFunction();
        KeypairAdd.prototype.__reactAutoBindMap.handleChangeKey = handleChangeKey;
        KeypairAdd.prototype.__reactAutoBindMap.handleChangeAuthorityId = handleChangeAuthorityId;

        // Render the component
        var keysPage = TestUtils.renderIntoDocument(
             <KeypairAdd token={token} />
        );

        expect(TestUtils.isCompositeComponent(keysPage)).toBeTruthy();

        // Get the text box and update it
        var textarea = TestUtils.findRenderedDOMComponentWithTag(keysPage, 'textarea');
        textarea.defaultValue = 'sushi-on-toast';
        TestUtils.Simulate.change(textarea);
        expect(handleChangeKey.mock.calls.length).toBe(1);

        // Get the AuthorityID field and update it
        var inputs = TestUtils.scryRenderedDOMComponentsWithTag(keysPage, 'input');
        var textAuthority = inputs[0];
        textAuthority.value = 'sushi-on-rye';
        TestUtils.Simulate.change(textAuthority);
        expect(handleChangeAuthorityId.mock.calls.length).toBe(1);
    });

    it('displays the alert box on error', function() {
        var KeypairAdd = require('../components/KeypairAdd');

        var shallowRenderer = TestUtils.createRenderer();

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
        var KeypairAdd = require('../components/KeypairAdd');

        // Render the component
        const component = shallow(
            <KeypairAdd />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })

    it('displays error with insufficient permissions', function() {
        var KeypairAdd = require('../components/KeypairAdd');

        // Render the component
        const component = shallow(
            <KeypairAdd token={tokenUser} />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
    })
});
