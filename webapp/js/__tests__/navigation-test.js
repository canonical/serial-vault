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


jest.dontMock('../components/Navigation');


describe('navigation', function() {
 it('displays the navigation menu with home active', function() {
	 var Navigation = require('../components/Navigation');
   var IntlProvider = require('react-intl').IntlProvider;
   var Messages = require('../components/messages').en;

   var handleYesClick = jest.genMockFunction();
   var handleNoClick = jest.genMockFunction();

	 // Render the component
	 var page = TestUtils.renderIntoDocument(
     <IntlProvider locale="en" messages={Messages}>
			 <Navigation active={'home'} />
     </IntlProvider>
	 );

	 expect(TestUtils.isCompositeComponent(page)).toBeTruthy();

	 // Check all the expected elements are rendered
	 var nav = TestUtils.findRenderedDOMComponentWithTag(page, 'nav');
   var ul = TestUtils.findRenderedDOMComponentWithTag(page, 'ul');
   expect(ul.children.length).toBe(3);
   expect(ul.children[0].firstChild.textContent).toBe('Home');
   expect(ul.children[0].firstChild.className).toBe('active');
   expect(ul.children[1].firstChild.className).toBe('');
   expect(ul.children[2].firstChild.className).toBe('');
 });

 it('displays the navigation menu with models active', function() {
	 var Navigation = require('../components/Navigation');
   var IntlProvider = require('react-intl').IntlProvider;
   var Messages = require('../components/messages').en;

   var handleYesClick = jest.genMockFunction();
   var handleNoClick = jest.genMockFunction();

	 // Render the component
	 var page = TestUtils.renderIntoDocument(
     <IntlProvider locale="en" messages={Messages}>
			 <Navigation active={'models'} />
     </IntlProvider>
	 );

	 expect(TestUtils.isCompositeComponent(page)).toBeTruthy();

	 // Check all the expected elements are rendered
	 var nav = TestUtils.findRenderedDOMComponentWithTag(page, 'nav');
   var ul = TestUtils.findRenderedDOMComponentWithTag(page, 'ul');
   expect(ul.children.length).toBe(3);
   expect(ul.children[1].firstChild.textContent).toBe('Models');
   expect(ul.children[1].firstChild.className).toBe('active');
   expect(ul.children[0].firstChild.className).toBe('');
   expect(ul.children[2].firstChild.className).toBe('');
 });

 it('displays the navigation menu with keys active', function() {
	 var Navigation = require('../components/Navigation');
   var IntlProvider = require('react-intl').IntlProvider;
   var Messages = require('../components/messages').en;

   var handleYesClick = jest.genMockFunction();
   var handleNoClick = jest.genMockFunction();

	 // Render the component
	 var page = TestUtils.renderIntoDocument(
     <IntlProvider locale="en" messages={Messages}>
			 <Navigation active={'keys'} />
     </IntlProvider>
	 );

	 expect(TestUtils.isCompositeComponent(page)).toBeTruthy();

	 // Check all the expected elements are rendered
	 var nav = TestUtils.findRenderedDOMComponentWithTag(page, 'nav');
   var ul = TestUtils.findRenderedDOMComponentWithTag(page, 'ul');
   expect(ul.children.length).toBe(3);
   expect(ul.children[2].firstChild.textContent).toBe('Public Keys');
   expect(ul.children[2].firstChild.className).toBe('active');
   expect(ul.children[0].firstChild.className).toBe('');
   expect(ul.children[1].firstChild.className).toBe('');
 });
});
