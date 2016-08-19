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


jest.dontMock('../components/ModelEdit');
jest.dontMock('../components/Navigation');

describe('model edit', function() {
 it('displays the model edit page for create', function() {
	 var ModelEdit = require('../components/ModelEdit');
	 var IntlProvider = require('react-intl').IntlProvider;
	 var Messages = require('../components/messages').en;

	 // Mock the data retrieval from the API
   var getModel = jest.genMockFunction();
   var getKeypairs = jest.genMockFunction();
   ModelEdit.WrappedComponent.prototype.__reactAutoBindMap.getModel = getModel;
   ModelEdit.WrappedComponent.prototype.__reactAutoBindMap.getKeypairs = getKeypairs;

	 // Render the component
	 var modelPage = TestUtils.renderIntoDocument(
			<IntlProvider locale="en" messages={Messages}>
			 <ModelEdit params={{}} />
		 </IntlProvider>
	 );

	 expect(TestUtils.isCompositeComponent(modelPage)).toBeTruthy();

   // Check all the expected elements are rendered
	 var section = TestUtils.findRenderedDOMComponentWithTag(modelPage, 'section');
	 var h2 = TestUtils.findRenderedDOMComponentWithTag(modelPage, 'h2');
   expect(h2.textContent).toBe('New Model');
   var nav = TestUtils.findRenderedDOMComponentWithTag(modelPage, 'nav');

   // Check that the navigation tag is set correctly
	 expect(nav.firstChild.children.length).toBe(3);
	 expect(nav.firstChild.children[1].firstChild.className).toBe('active');
	 expect(nav.firstChild.children[1].firstChild.textContent).toBe('Models');

   // Check the data retrieval calls
   expect(getModel.mock.calls.length).toBe(0);
   expect(getKeypairs.mock.calls.length).toBe(1);

   // Check that the form is rendered without data
   var inputs = TestUtils.scryRenderedDOMComponentsWithTag(modelPage, 'input');
   expect(inputs.length).toBe(2);
   expect(inputs[0].value).toBe('');
   expect(inputs[1].value).toBe('');

 });

 it('displays the model edit page for an existing model', function() {
   var ModelEdit = require('../components/ModelEdit');
	 var IntlProvider = require('react-intl').IntlProvider;
	 var Messages = require('../components/messages').en;

	 // Mock the data retrieval from the API
   var getModel = jest.genMockFunction();
   var getKeypairs = jest.genMockFunction();
   var handleSaveClick = jest.genMockFunction();
   ModelEdit.WrappedComponent.prototype.__reactAutoBindMap.getModel = getModel;
   ModelEdit.WrappedComponent.prototype.__reactAutoBindMap.getKeypairs = getKeypairs;
   ModelEdit.WrappedComponent.prototype.__reactAutoBindMap.handleSaveClick = handleSaveClick;

   //var MODEL = {id: 1, 'brand-id': 'Brand1', model: 'Name1', 'authority-id': 'Brand1', 'key-id': 'Name1'};

	 // Render the component
	 var modelPage = TestUtils.renderIntoDocument(
			<IntlProvider locale="en" messages={Messages}>
			 <ModelEdit params={{id: 1}} />
		 </IntlProvider>
	 );

	 expect(TestUtils.isCompositeComponent(modelPage)).toBeTruthy();

   // Check the data retrieval calls
   //expect(getModel.mock.calls.length).toBe(1);
   expect(getKeypairs.mock.calls.length).toBe(1);

   // Get the save link
   var anchors = TestUtils.scryRenderedDOMComponentsWithTag(modelPage, 'a');
   expect(anchors.length).toBe(5);
   expect(anchors[3].textContent).toBe('Save');
   TestUtils.Simulate.click(anchors[3]);
   expect(handleSaveClick.mock.calls.length).toBe(1);
 });

});
