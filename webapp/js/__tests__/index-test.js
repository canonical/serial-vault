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
//'use strict'

import React from 'react';
import ReactDOM from 'react-dom';
import TestUtils from 'react-addons-test-utils';

var en = require('../components/messages').en;

jest.dontMock('../components/Index');
jest.dontMock('../components/Navigation');
jest.dontMock('../components/messages');
jest.dontMock('react-intl')
jest.dontMock('intl-messageformat')
jest.dontMock('intl-relativeformat')

// import {addLocaleData} from 'react-intl';
// addLocaleData([...en]);


describe('index', function() {
 it('displays the index page elements', function() {
	var Index = require('../components/Index');
	// var IntlProvider = ReactIntl.IntlProvider;
	// var Messages = require('../components/messages').en;

	 // Render the component
	 /*var indexPage = TestUtils.renderIntoDocument(
     <IntlProvider locale="en" messages={Messages}>
			 <Index />
     </IntlProvider>
	 );*/

// 	 const component = createComponentWithIntl(<AppMain/>);
// 	 console.log(component)
//   let tree = component.toJSON();

//   expect(tree).toMatchSnapshot();

	 /*const indexPage = TestUtils.renderIntoDocument(
		<createIntlComponent>
			<Index />
		</createIntlComponent>
	 )
	 console.log(indexPage)

	 expect(TestUtils.isCompositeComponent(indexPage)).toBeTruthy();

	 // Check all the expected elements are rendered
	 var section = TestUtils.findRenderedDOMComponentWithTag(indexPage, 'section');
	 var h2 = TestUtils.findRenderedDOMComponentWithTag(indexPage, 'h2');
	 var nav = TestUtils.findRenderedDOMComponentWithTag(indexPage, 'nav');

	 // Check that the navigation tag is set correctly
	 expect(nav.firstChild.children.length).toBe(3);
	 expect(nav.firstChild.children[0].firstChild.className).toBe('active');
	 expect(nav.firstChild.children[0].firstChild.textContent).toBe('Home');*/
 });
});
