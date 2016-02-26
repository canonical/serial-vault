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


jest.dontMock('../components/KeyList');
jest.dontMock('../components/Navigation');
jest.dontMock('../components/DialogBox');

describe('key list', function() {
	it('displays the keys page with no keys', function() {
		var KeyList = require('../components/KeyList');

		// Mock the data retrieval from the API
    var getKeys = jest.genMockFunction();
    KeyList.prototype.__reactAutoBindMap.getKeys = getKeys;

 	 // Render the component
 	 var keysPage = TestUtils.renderIntoDocument(
 			 <KeyList />
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

    // Check the getKeys was called
    expect(getKeys.mock.calls.length).toBe(1);

    // Check the 'no public keys' message is rendered
    expect(section.children.length).toBe(4);
    expect(section.lastChild.textContent).toBe('No public keys found.');
	});

	it('displays the keys page with some keys', function() {
		var KeyList = require('../components/KeyList');

		// Mock the data retrieval from the API
    var getKeys = jest.genMockFunction();
    KeyList.prototype.__reactAutoBindMap.getKeys = getKeys;

 	 // Render the component
 	 var keysPage = TestUtils.renderIntoDocument(
 			 <KeyList />
 	 );

	 // Set up a fixture for the keys data
   keysPage.setState({keys: [
		 'rsa-ssh abcdef0123456789 Comment',
		 'rsa-ssh 0123456789abcdef No Comment',
		 'rsa-ssh abc0123456789def Another Comment'
   ]});

 	 expect(TestUtils.isCompositeComponent(keysPage)).toBeTruthy();

	 // Check all the expected elements are rendered
	 var section = TestUtils.findRenderedDOMComponentWithTag(keysPage, 'section');
	 var h2 = TestUtils.findRenderedDOMComponentWithTag(keysPage, 'h2');
	 var nav = TestUtils.findRenderedDOMComponentWithTag(keysPage, 'nav');
   var table = TestUtils.findRenderedDOMComponentWithTag(keysPage, 'table');

	 // Check that the navigation tag is set correctly
	 expect(nav.firstChild.children.length).toBe(3);
	 expect(nav.firstChild.children[2].firstChild.className).toBe('active');
	 expect(nav.firstChild.children[2].firstChild.textContent).toBe('Public Keys');

   // Check the getModels was called
   expect(getKeys.mock.calls.length).toBe(1);

   // Check that the table is rendered correctly
   expect(table.lastChild.children.length).toBe(3); // data rows
   expect(table.lastChild.children[0].children.length).toBe(1); // cells
   expect(table.lastChild.children[0].children[0].children[2].textContent).toBe('rsa-ssh abcdef0123456789 Comment');

	});

	it('displays the delete confirmation on clicking Remove', function() {
		var KeyList = require('../components/KeyList');

		// Mock the data retrieval from the API
    var getKeys = jest.genMockFunction();
    KeyList.prototype.__reactAutoBindMap.getKeys = getKeys;

		// Mock the deletion of a key
    var handleRemoveClick = jest.genMockFunction();
    KeyList.prototype.__reactAutoBindMap.handleRemoveClick = handleRemoveClick;

 	 // Render the component
 	 var keysPage = TestUtils.renderIntoDocument(
 			 <KeyList />
 	 );

	 // Set up a fixture for the keys data
   keysPage.setState({keys: [
		 'rsa-ssh abcdef0123456789 Comment',
		 'rsa-ssh 0123456789abcdef No Comment',
		 'rsa-ssh abc0123456789def Another Comment'
   ]});

	 expect(TestUtils.isCompositeComponent(keysPage)).toBeTruthy();

	 var table = TestUtils.findRenderedDOMComponentWithTag(keysPage, 'table');

	 // Check that the table is rendered correctly
   expect(table.lastChild.children.length).toBe(3); // data rows

	 // Get the remove button for the first item and click it
	 var removeButton = table.lastChild.children[0].firstChild.firstChild.firstChild;
	 TestUtils.Simulate.click(removeButton);

	 // Should have the confirmation dialog shown
	 var dialog = TestUtils.findRenderedDOMComponentWithClass(keysPage, 'warning');
	 expect(dialog.firstChild.textContent).toBe('Confirm deletion of the public key.');
	 expect(keysPage.state.confirmDelete).toBe(0);

	 // Clicking the confirmation button should trigger the delete call
	 var yesButton = dialog.children[1].firstChild;
	 TestUtils.Simulate.click(yesButton);
	 expect(handleRemoveClick.mock.calls.length).toBe(1);

 });

});
