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


jest.dontMock('../components/KeyAdd');
jest.dontMock('../components/Navigation');
jest.dontMock('../components/AlertBox');

describe('key add', function() {
	it('displays the new public key page', function() {
		var IntlProvider = require('react-intl').IntlProvider;
		var Messages = require('../components/messages').en;
		var KeyAdd = require('../components/KeyAdd');

		// Render the component
		var keysPage = TestUtils.renderIntoDocument(
			<IntlProvider locale="en" messages={Messages}>
			 <KeyAdd />
			</IntlProvider>
		);

		expect(TestUtils.isCompositeComponent(keysPage)).toBeTruthy();

		// Check all the expected elements are rendered
		var section = TestUtils.findRenderedDOMComponentWithTag(keysPage, 'section');
		var h2 = TestUtils.findRenderedDOMComponentWithTag(keysPage, 'h2');
		var nav = TestUtils.findRenderedDOMComponentWithTag(keysPage, 'nav');

		// Check that the navigation tag is set correctly
		expect(nav.firstChild.children.length).toBe(3);
		expect(nav.firstChild.children[2].firstChild.className).toBe('active');
		expect(nav.firstChild.children[2].firstChild.textContent).toBe('Public Keys');

	});

	it('stores updates to the form', function() {
		var IntlProvider = require('react-intl').IntlProvider;
		var Messages = require('../components/messages').en;
		var KeyAdd = require('../components/KeyAdd');

		// Mock the onChange handler
    var handleChangeKey = jest.genMockFunction();
    KeyAdd.WrappedComponent.prototype.__reactAutoBindMap.handleChangeKey = handleChangeKey;

		// Render the component
		var keysPage = TestUtils.renderIntoDocument(
			<IntlProvider locale="en" messages={Messages}>
			 <KeyAdd />
			</IntlProvider>
		);

		expect(TestUtils.isCompositeComponent(keysPage)).toBeTruthy();

		// Get the text box and update it
		var textarea = TestUtils.findRenderedDOMComponentWithTag(keysPage, 'textarea');
		textarea.defaultValue = 'sushi-on-toast';
		TestUtils.Simulate.change(textarea);
		expect(handleChangeKey.mock.calls.length).toBe(1);

	});

	it('displays the alert box on error', function() {
		var IntlProvider = require('react-intl').IntlProvider;
		var Messages = require('../components/messages').en;
		var KeyAdd = require('../components/KeyAdd');

		const intlProvider = new IntlProvider({locale: 'en', messages: Messages}, {});
    const {intl} = intlProvider.getChildContext();
		var shallowRenderer = TestUtils.createRenderer();

		// Render the component
		shallowRenderer.render(
			<KeyAdd.WrappedComponent intl={intl} error={'Critical: run out of sushi'} />
		);
		var keysPage = shallowRenderer.getRenderOutput();

		expect(keysPage.props.children.length).toBe(3);
		var section = keysPage.props.children[1];

		expect(section.props.children.length).toBe(2);
		expect(section.props.children[1].props.children[0].props.message).toBe('Critical: run out of sushi');
	});

});
