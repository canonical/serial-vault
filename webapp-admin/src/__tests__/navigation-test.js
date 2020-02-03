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
import Navigation from '../components/Navigation';
import NavigationUser from '../components/NavigationUser';

jest.dontMock('../components/Navigation');
jest.dontMock('../components/Utils');

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

const systemUser = { role: 100, name: 'Steven Vault' }
const tokenAdmin = { role: 200, name: 'Steven Vault' }
const superUser = { role: 300, name: 'Steven Vault' }
const tokenDisabled = { role: 200 }

describe('navigation', function() {
  it('displays the navigation menu with models active for systemuser', function() {
    // Render the component
    var page = ReactTestUtils.renderIntoDocument(
        <Navigation active={'models'} token={systemUser} accounts={[]} />
    );

    expect(ReactTestUtils.isCompositeComponent(page)).toBeTruthy();

    // Check all the expected elements are rendered
    var ul = ReactTestUtils.findRenderedDOMComponentWithTag(page, 'ul');
    expect(ul.children.length).toBe(2);
    expect(ul.children[1].firstChild.textContent).toBe('System-User');
  });

  it('displays the navigation menu with models active for superuser', function() {
    // Render the component
    var page = ReactTestUtils.renderIntoDocument(
        <Navigation active={'models'} token={superUser} accounts={[]} />
    );

    expect(ReactTestUtils.isCompositeComponent(page)).toBeTruthy();

    // Check all the expected elements are rendered
    var ul = ReactTestUtils.findRenderedDOMComponentWithTag(page, 'ul');
    expect(ul.children.length).toBe(6);
    expect(ul.children[1].firstChild.textContent).toBe('Accounts');
    expect(ul.children[2].firstChild.textContent).toBe('Signing Keys');
    expect(ul.children[3].firstChild.textContent).toBe('Models');
    expect(ul.children[4].firstChild.textContent).toBe('Signing Log');
    expect(ul.children[5].firstChild.textContent).toBe('Users');
  });

  it('displays the navigation menu with models active for admin', function() {

    // Render the component
    var page = ReactTestUtils.renderIntoDocument(
        <Navigation active={'models'} token={tokenAdmin} accounts={[]} />
    );

    expect(ReactTestUtils.isCompositeComponent(page)).toBeTruthy();

    // Check all the expected elements are rendered
    var ul = ReactTestUtils.findRenderedDOMComponentWithTag(page, 'ul');
    // 4 links and account menu
    expect(ul.children.length).toBe(5);
  });

  it('displays the navigation menu with models active', function() {

    // Render the component
    var page = ReactTestUtils.renderIntoDocument(
        <Navigation active={'models'} token={tokenAdmin} accounts={[]} />
    );

    expect(ReactTestUtils.isCompositeComponent(page)).toBeTruthy();

    // Check all the expected elements are rendered
    var ul = ReactTestUtils.findRenderedDOMComponentWithTag(page, 'ul');
    expect(ul.children.length).toBe(5);
    expect(ul.children[1].firstChild.textContent).toBe('Accounts');
    expect(ul.children[1].firstChild.className).toBe('');
    expect(ul.children[1].firstChild.className).toBe('');
    expect(ul.children[2].firstChild.textContent).toBe('Signing Keys');
    expect(ul.children[3].firstChild.textContent).toBe('Models');
    expect(ul.children[4].firstChild.textContent).toBe('Signing Log');
  });

  it('displays the OpenID link when user auth is enabled', function() {

      // Render the component
      var page = ReactTestUtils.renderIntoDocument(
          <NavigationUser token={tokenAdmin} accounts={[]} />
      );

      expect(ReactTestUtils.isCompositeComponent(page)).toBeTruthy();

      // Check all the expected elements are rendered
      var ul = ReactTestUtils.findRenderedDOMComponentWithTag(page, 'ul');
      expect(ul.children.length).toBe(2);
      expect(ul.children[0].firstChild.textContent).toBe(tokenAdmin.name);
      expect(ul.children[1].firstChild.textContent).toBe('Logout');
  })

  it('omits the OpenID link when user auth is disabled', function() {

      // Render the component
      var page = ReactTestUtils.renderIntoDocument(
          <NavigationUser token={tokenDisabled} accounts={[]} />
      );

      expect(ReactTestUtils.isCompositeComponent(page)).toBeTruthy();

      // Check all the expected elements are rendered
      var ul = ReactTestUtils.findRenderedDOMComponentWithTag(page, 'ul');
      expect(ul.children.length).toBe(0);
  })

});
