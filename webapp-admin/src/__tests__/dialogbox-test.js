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
import DialogBox from '../components/DialogBox';

jest.dontMock('../components/DialogBox');

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};


describe('dialog box', function() {
 it('displays the dialog box with a message', function() {

  var handleYesClick = jest.genMockFunction();
  var handleNoClick = jest.genMockFunction();

  // Render the component
  var page = ReactTestUtils.renderIntoDocument(
      <DialogBox message={'The message goes here'} handleYesClick={handleYesClick} handleCancelClick={handleNoClick} />
  );

  expect(ReactTestUtils.isCompositeComponent(page)).toBeTruthy();

  // Check all the expected elements are rendered
  var divs = ReactTestUtils.scryRenderedDOMComponentsWithTag(page, 'div');
  expect(divs.length).toBe(2);
  var anchors = ReactTestUtils.scryRenderedDOMComponentsWithTag(page, 'a');
  expect(anchors.length).toBe(2);

  expect(handleYesClick.mock.calls.length).toBe(0);
  expect(handleNoClick.mock.calls.length).toBe(0);

  // Click each button and check the callback
  ReactTestUtils.Simulate.click(anchors[1]);
  expect(handleYesClick.mock.calls.length).toBe(1);

  // Click each button and check the callback
  ReactTestUtils.Simulate.click(anchors[0]);
  expect(handleNoClick.mock.calls.length).toBe(1);
 });

 it('displays no box when there is no message', function() {

  // Render the component
  var page = ReactTestUtils.renderIntoDocument(
      <DialogBox message={null} />
  );

  expect(ReactTestUtils.isCompositeComponent(page)).toBeTruthy();
  var span = ReactTestUtils.scryRenderedDOMComponentsWithTag(page, 'span');
 });

});
