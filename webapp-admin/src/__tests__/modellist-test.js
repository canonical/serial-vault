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
import ReactTestUtils from 'react-dom/test-utils';
import Adapter from 'enzyme-adapter-react-16';
import {shallow, configure} from 'enzyme';
import ModelList from '../components/ModelList';

configure({ adapter: new Adapter() });

jest.dontMock('../components/ModelList');
jest.dontMock('../components/KeypairList');
jest.dontMock('../components/ModelRow');
jest.dontMock('../components/Navigation');
jest.dontMock('../components/Utils');

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

const token = { role: 200 }
const tokenUser = { role: 100 }


describe('model list', function() {
  it('displays the models page with no models', function() {

    // Mock the data retrieval from the API
    var getModels = jest.fn();
    var getKeypairs = jest.fn();
    ModelList.prototype.getModels = getModels;
    ModelList.prototype.getKeypairs = getKeypairs;

    // Render the component
    var modelsPage = ReactTestUtils.renderIntoDocument(
      <ModelList token={token} models={[]} />
    );

    expect(ReactTestUtils.isCompositeComponent(modelsPage)).toBeTruthy();

    // Check all the expected elements are rendered
    var sections = ReactTestUtils.scryRenderedDOMComponentsWithTag(modelsPage, 'section');
    expect(sections.length).toBe(2)
    var h2s = ReactTestUtils.scryRenderedDOMComponentsWithTag(modelsPage, 'h2');
    expect(h2s.length).toBe(2);

    // Check the getModels was called
    // expect(getModels.mock.calls.length).toBe(1);

    // Check the 'no models' and 'no keypairs' message is rendered
    expect(sections[0].children.length).toBe(4);
    expect(sections[0].children[3].textContent).toBe('No models found.');
    expect(sections[1].children[1].textContent).toBe('No signing keys found');
  });

 it('displays the models page with some models', function() {

  // Set up a fixture for the model data
  var models = [
    {id: 1, 'brand-id': 'Brand1', model: 'Name1'},
    {id: 2, 'brand-id': 'Brand2', model: 'Name2'},
    {id: 3, 'brand-id': 'Brand3', model: 'Name3'}
  ];

  // Render the component
  var modelsPage = shallow(
      <ModelList models={models} token={token} />
  );

  expect(modelsPage.find('section')).toHaveLength(2)
  expect(modelsPage.find('ModelRow')).toHaveLength(3)
  expect(modelsPage.find('KeypairList')).toHaveLength(1)

 });

 it('displays the models page with some keypairs', function() {

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
  var getModels = jest.fn();
  var getKeypairs = jest.fn();
  ModelList.prototype.getModels = getModels;
  ModelList.prototype.getKeypairs = getKeypairs;

  // Render the component
  var modelsPage = ReactTestUtils.renderIntoDocument(
      <ModelList models={models} keypairs={keypairs} token={token} showAssert={1} />
  );

  expect(ReactTestUtils.isCompositeComponent(modelsPage)).toBeTruthy();

  // Check all the expected elements are rendered
  var sections = ReactTestUtils.scryRenderedDOMComponentsWithTag(modelsPage, 'section');
  expect(sections.length).toBe(2)
  var h2s = ReactTestUtils.scryRenderedDOMComponentsWithTag(modelsPage, 'h2');
  expect(h2s.length).toBe(2);

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

  it('displays error with no permissions', function() {

      // Render the component
      const component = shallow(
          <ModelList />
      );

      expect(component.find('div')).toHaveLength(1)
      expect(component.find('AlertBox')).toHaveLength(1)
  })

  it('displays error with insufficient permissions', function() {

      // Render the component
      const component = shallow(
          <ModelList token={tokenUser} />
      );

      expect(component.find('div')).toHaveLength(1)
      expect(component.find('AlertBox')).toHaveLength(1)
  })

});
