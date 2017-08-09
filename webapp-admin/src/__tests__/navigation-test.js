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
import Navigation from '../components/Navigation';
import NavigationUser from '../components/NavigationUser';

jest.dontMock('../components/Navigation');
jest.dontMock('../components/Utils');

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

const token = { role: 200, name: 'Steven Vault' }
const tokenDisabled = { role: 200 }

describe('navigation', function() {
  it('displays the navigation menu with models active', function() {

    var handleYesClick = jest.genMockFunction();
    var handleNoClick = jest.genMockFunction();

    // Render the component
    var page = ReactTestUtils.renderIntoDocument(
        <Navigation active={'models'} token={token} />
    );

    expect(ReactTestUtils.isCompositeComponent(page)).toBeTruthy();

    // Check all the expected elements are rendered
    var ul = ReactTestUtils.findRenderedDOMComponentWithTag(page, 'ul');
    expect(ul.children.length).toBe(4);
  });

  it('displays the navigation menu with models active', function() {

    var handleYesClick = jest.genMockFunction();
    var handleNoClick = jest.genMockFunction();

    // Render the component
    var page = ReactTestUtils.renderIntoDocument(
        <Navigation active={'models'} token={token} />
    );

    expect(ReactTestUtils.isCompositeComponent(page)).toBeTruthy();

    // Check all the expected elements are rendered
    var ul = ReactTestUtils.findRenderedDOMComponentWithTag(page, 'ul');
    expect(ul.children.length).toBe(4);
    expect(ul.children[0].firstChild.textContent).toBe('Models');
    expect(ul.children[0].firstChild.className).toBe('');
    expect(ul.children[0].firstChild.className).toBe('');
  });

  it('displays the OpenID link when user auth is enabled', function() {

      // Render the component
      var page = ReactTestUtils.renderIntoDocument(
          <NavigationUser token={token} />
      );

      expect(ReactTestUtils.isCompositeComponent(page)).toBeTruthy();

      // Check all the expected elements are rendered
      var ul = ReactTestUtils.findRenderedDOMComponentWithTag(page, 'ul');
      expect(ul.children.length).toBe(2);
      expect(ul.children[0].firstChild.textContent).toBe(token.name);
      expect(ul.children[1].firstChild.textContent).toBe('Logout');
  })

  it('omits the OpenID link when user auth is disabled', function() {

      // Render the component
      var page = ReactTestUtils.renderIntoDocument(
          <NavigationUser token={tokenDisabled} />
      );

      expect(ReactTestUtils.isCompositeComponent(page)).toBeTruthy();

      // Check all the expected elements are rendered
      var ul = ReactTestUtils.findRenderedDOMComponentWithTag(page, 'ul');
      expect(ul.children.length).toBe(0);
  })

});
