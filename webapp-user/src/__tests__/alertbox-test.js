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


jest.dontMock('../components/AlertBox');


describe('alert box', function() {
 it('displays the alert box with a message', function() {
    var AlertBox = require('../components/AlertBox');

    var handleYesClick = jest.genMockFunction();
    var handleNoClick = jest.genMockFunction();

    // Render the component
    var page = TestUtils.renderIntoDocument(
      <AlertBox message={'The message goes here'} />
    );

     expect(TestUtils.isCompositeComponent(page)).toBeTruthy();

     // Check all the expected elements are rendered
     var div = TestUtils.findRenderedDOMComponentWithTag(page, 'div');
   var p = TestUtils.findRenderedDOMComponentWithTag(page, 'p');
   expect(p.textContent).toBe('The message goes here');

 });

 it('displays no box when there is no message', function() {
     var AlertBox = require('../components/AlertBox');

   // Render the component
     var page = TestUtils.renderIntoDocument(
     <AlertBox message={null} />
     );

   expect(TestUtils.isCompositeComponent(page)).toBeTruthy();
   var span = TestUtils.scryRenderedDOMComponentsWithTag(page, 'span');
 });

});
