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

    // Render the component
    var modelsPage = ReactTestUtils.renderIntoDocument(
      <ModelList token={token} models={[]} />
    );

    expect(ReactTestUtils.isCompositeComponent(modelsPage)).toBeTruthy();

    // Check all the expected elements are rendered
    var sections = ReactTestUtils.scryRenderedDOMComponentsWithTag(modelsPage, 'section');
    expect(sections.length).toBe(1)
    var h2s = ReactTestUtils.scryRenderedDOMComponentsWithTag(modelsPage, 'h2');
    expect(h2s.length).toBe(1);

    // Check the 'no models' message is rendered
    expect(sections[0].children.length).toBe(4);
    expect(sections[0].children[3].textContent).toBe('No models found.');
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

    expect(modelsPage.find('section')).toHaveLength(1)
    expect(modelsPage.find('ModelRow')).toHaveLength(3)
 });

 it('displays the models page with some keypairs', function() {

    // Set up a fixture for the model data
    var models = [
      {id: 1, 'brand-id': 'Brand1', model: 'Name1', 'authority-id': 'Brand1', 'key-id': 'Name1'},
      {id: 2, 'brand-id': 'Brand2', model: 'Name2', 'authority-id': 'Brand1', 'key-id': 'Name1'},
      {id: 3, 'brand-id': 'Brand3', model: 'Name3', 'authority-id': 'Brand1', 'key-id': 'Name1'}
    ];

    // Mock the data retrieval from the API
    var getModels = jest.fn();
    var getKeypairs = jest.fn();
    ModelList.prototype.getModels = getModels;
    ModelList.prototype.getKeypairs = getKeypairs;

    // Render the component
    var modelsPage = shallow(
        <ModelList models={models} token={token} showAssert={1} />
    );

    expect(ReactTestUtils.isCompositeComponent(modelsPage)).toBeTruthy();

    // Check all the expected elements are rendered
    var sections = modelsPage.find('section');
    expect(sections.length).toBe(1)
    var h2s = modelsPage.find('h2');
    expect(h2s.length).toBe(1);

    var sectionModelRows = modelsPage.find('ModelRow')
    expect(sectionModelRows.length).toBe(3);
    expect(sectionModelRows.get(0).props.model).toEqual(models[0])
    expect(sectionModelRows.get(1).props.model).toEqual(models[1])
    expect(sectionModelRows.get(2).props.model).toEqual(models[2])
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
