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

describe('model edit', function() {
 it('displays the model edit page for create', function() {
	 var ModelEdit = require('../components/ModelEdit');
	 var IntlProvider = require('react-intl').IntlProvider;
	 var Messages = require('../components/messages').en;

	 // Mock the data retrieval from the API
   var getModels = jest.genMockFunction();
   var getKeypairs = jest.genMockFunction();
   ModelEdit.WrappedComponent.prototype.__reactAutoBindMap.getModels = getModels;
   ModelEdit.WrappedComponent.prototype.__reactAutoBindMap.getKeypairs = getKeypairs;

	 // Render the component
	 var modelPage = TestUtils.renderIntoDocument(
			<IntlProvider locale="en" messages={Messages}>
			 <ModelEdit params={{}} />
		 </IntlProvider>
	 );

	 expect(TestUtils.isCompositeComponent(modelPage)).toBeTruthy();

	 

 });

});
