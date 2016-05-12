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


jest.dontMock('../components/Footer');


describe('footer', function() {
 it('displays the footer', function() {
	 var Footer = require('../components/Footer');
   var IntlProvider = require('react-intl').IntlProvider;
   var Messages = require('../components/messages').en;

   var getVersion = jest.genMockFunction();
   Footer.WrappedComponent.prototype.__reactAutoBindMap.getVersion = getVersion;

	 // Render the component
	 var page = TestUtils.renderIntoDocument(
     <IntlProvider locale="en" messages={Messages}>
			 <Footer />
     </IntlProvider>
	 );

	 expect(TestUtils.isCompositeComponent(page)).toBeTruthy();

	 // Check all the expected elements are rendered
	 var footer = TestUtils.findRenderedDOMComponentWithTag(page, 'footer');
	 var div = TestUtils.findRenderedDOMComponentWithTag(page, 'div');
   expect(div.textContent).toContain('Version');

 });

});
