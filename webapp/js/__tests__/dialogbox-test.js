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


jest.dontMock('../components/DialogBox');


describe('dialog box', function() {
 it('displays the dialog box with a message', function() {
  var DialogBox = require('../components/DialogBox');

  var handleYesClick = jest.genMockFunction();
  var handleNoClick = jest.genMockFunction();

  // Render the component
  var page = TestUtils.renderIntoDocument(
      <DialogBox message={'The message goes here'} handleYesClick={handleYesClick} handleCancelClick={handleNoClick} />
  );

  expect(TestUtils.isCompositeComponent(page)).toBeTruthy();

  // Check all the expected elements are rendered
  var divs = TestUtils.scryRenderedDOMComponentsWithTag(page, 'div');
  expect(divs.length).toBe(2);
  var anchors = TestUtils.scryRenderedDOMComponentsWithTag(page, 'a');
  expect(anchors.length).toBe(2);

  expect(handleYesClick.mock.calls.length).toBe(0);
  expect(handleNoClick.mock.calls.length).toBe(0);

  // Click each button and check the callback
  TestUtils.Simulate.click(anchors[1]);
  expect(handleYesClick.mock.calls.length).toBe(1);

  // Click each button and check the callback
  TestUtils.Simulate.click(anchors[0]);
  expect(handleNoClick.mock.calls.length).toBe(1);
 });

 it('displays no box when there is no message', function() {
  var DialogBox = require('../components/DialogBox');

  // Render the component
  var page = TestUtils.renderIntoDocument(
      <DialogBox message={null} />
  );

  expect(TestUtils.isCompositeComponent(page)).toBeTruthy();
  var span = TestUtils.scryRenderedDOMComponentsWithTag(page, 'span');
 });

});
