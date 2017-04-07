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

jest.dontMock('../components/ModelList');
jest.dontMock('../components/KeypairList');
jest.dontMock('../components/ModelRow');
jest.dontMock('../components/Navigation');
jest.dontMock('../components/Utils');

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};


describe('model list', function() {
  it('displays the models page with no models', function() {
    var ModelList = require('../components/ModelList');

    // Mock the data retrieval from the API
    var getModels = jest.genMockFunction();
    var getKeypairs = jest.genMockFunction();
    ModelList.prototype.__reactAutoBindMap.getModels = getModels;
    ModelList.prototype.__reactAutoBindMap.getKeypairs = getKeypairs;

    // Render the component
    var modelsPage = TestUtils.renderIntoDocument(
        <ModelList />
    );

    expect(TestUtils.isCompositeComponent(modelsPage)).toBeTruthy();

    // Check all the expected elements are rendered
    var sections = TestUtils.scryRenderedDOMComponentsWithTag(modelsPage, 'section');
    expect(sections.length).toBe(2)
    var h2s = TestUtils.scryRenderedDOMComponentsWithTag(modelsPage, 'h2');
    expect(h2s.length).toBe(2);

    // Check the getModels was called
    expect(getModels.mock.calls.length).toBe(1);

    // Check the 'no models' and 'no keypairs' message is rendered
    expect(sections[0].children.length).toBe(4);
    expect(sections[0].children[3].textContent).toBe('No models found.');
    expect(sections[1].children[1].textContent).toBe('No signing keys found');
  });

 it('displays the models page with some models', function() {
  var ModelList = require('../components/ModelList');

  // Set up a fixture for the model data
  var models = [
    {id: 1, 'brand-id': 'Brand1', model: 'Name1'},
    {id: 2, 'brand-id': 'Brand2', model: 'Name2'},
    {id: 3, 'brand-id': 'Brand3', model: 'Name3'}
  ];

  // Mock the data retrieval from the API
  var getModels = jest.genMockFunction();
  ModelList.prototype.__reactAutoBindMap.getModels = getModels;

  // Shallow render the component
  var shallowRenderer = TestUtils.createRenderer();

  // Render the component
  shallowRenderer.render(
      <ModelList models={models} />
  );
  var modelsPage = shallowRenderer.getRenderOutput();

  expect(modelsPage.props.children.length).toBe(2);
  var sections = modelsPage.props.children;
  expect(sections[0].props.children.length).toBe(4);

  // Check that the models table is rendered correctly
  var table = sections[0].props.children[3].props.children;
  var tbody = table.props.children[1]
  expect(tbody.props.children.length).toBe(3); // data rows
  var row1 = tbody.props.children[0];

  expect(row1.type.displayName).toBe('ModelRow')
  expect(row1.props.model).toBe(models[0])

  // Check the keypair section
  var keypairs = sections[1].props.children[1].props.children;
  expect(keypairs.type.displayName).toBe('KeypairList');
 });

 it('displays the models page with some keypairs', function() {
  var ModelList = require('../components/ModelList');

  // Set up a fixture for the model data
  var models = [
    {id: 1, 'brand-id': 'Brand1', model: 'Name1', 'authority-id': 'Brand1', 'key-id': 'Name1'},
    {id: 2, 'brand-id': 'Brand2', model: 'Name2', 'authority-id': 'Brand1', 'key-id': 'Name1'},
    {id: 3, 'brand-id': 'Brand3', model: 'Name3', 'authority-id': 'Brand1', 'key-id': 'Name1'}
  ];

  // Set up a fixture for the keypair data
  var keypairs = [
    {ID: 11, AuthorityID: 'Brand1', KeyID: 'Name1', Active: 1},
    {ID: 22, AuthorityID: 'Brand2', KeyID: 'Name2', Active: 2},
    {ID: 33, AuthorityID: 'Brand3', KeyID: 'Name3', Active: 3}
  ];

  // Mock the data retrieval from the API
  var getModels = jest.genMockFunction();
  var getKeypairs = jest.genMockFunction();
  ModelList.prototype.__reactAutoBindMap.getModels = getModels;
  ModelList.prototype.__reactAutoBindMap.getKeypairs = getKeypairs;

  // Render the component
  var modelsPage = TestUtils.renderIntoDocument(
      <ModelList models={models} keypairs={keypairs} />
  );

  expect(TestUtils.isCompositeComponent(modelsPage)).toBeTruthy();

  // Check all the expected elements are rendered
  var sections = TestUtils.scryRenderedDOMComponentsWithTag(modelsPage, 'section');
  expect(sections.length).toBe(2)
  var h2s = TestUtils.scryRenderedDOMComponentsWithTag(modelsPage, 'h2');
  expect(h2s.length).toBe(2);

  // Check the mocks were called
  expect(getModels.mock.calls.length).toBe(1);
  expect(getKeypairs.mock.calls.length).toBe(1);

  // Check the models table
  var sectionModelRows = sections[0].children[3].children[0].children[1].children;
  expect(sectionModelRows.length).toBe(3);
  expect(sectionModelRows[1].children.length).toBe(6);
  expect(sectionModelRows[1].children[2].textContent).toBe(models[1].model);
  expect(sectionModelRows[1].children[3].textContent).toBe(models[1]['authority-id'].concat('/', models[1]['key-id']));

  // Check the keypairs table
  var sectionKeypairRows = sections[1].children[1].children[0].children[1].children;
  expect(sectionKeypairRows.length).toBe(3);
  expect(sectionKeypairRows[0].children.length).toBe(4);
  expect(sectionKeypairRows[0].children[1].textContent).toBe(keypairs[0].AuthorityID);
  expect(sectionKeypairRows[0].children[2].textContent).toBe(keypairs[0].KeyID);

 });

});
