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
import Adapter from 'enzyme-adapter-react-16';
import {shallow, configure} from 'enzyme';
import UserList from '../components/UserList';

jest.dontMock('../components/UserList');
jest.dontMock('../components/UserRow');
jest.dontMock('../components/Navigation');
jest.dontMock('../components/Utils');

configure({ adapter: new Adapter() });

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

const token = { role: 300 }
const tokenUser = { role: 100 }


describe('user list', function() {
  it('displays the users page with no users', function() {

    // Mock the data retrieval from the API
    var getUsers = jest.fn();
    UserList.prototype.getUsers = getUsers;

    // Render the component
    var usersPage = ReactTestUtils.renderIntoDocument(
        <UserList token={token} />
    );

    expect(ReactTestUtils.isCompositeComponent(usersPage)).toBeTruthy();

    // Check all the expected elements are rendered
    var sections = ReactTestUtils.scryRenderedDOMComponentsWithTag(usersPage, 'section');
    expect(sections.length).toBe(1)
    var h2s = ReactTestUtils.scryRenderedDOMComponentsWithTag(usersPage, 'h2');
    expect(h2s.length).toBe(1);

    // Check the getUsers was called
    // expect(getUsers.mock.calls.length).toBe(1);

    // Check the 'no users'message is rendered
    expect(sections[0].children.length).toBe(4);
    expect(sections[0].children[3].textContent).toBe('No users found.');
  });

  it('displays the users page with some users', function() {

  // Set up a fixture for the user data
  var users = [
    {ID: 1, Username: 'user1', Name: 'User One', Email: "user1@domain.dom", Role: 100},
    {ID: 2, Username: 'user2', Name: 'User Two', Email: "user2@domain.dom", Role: 200},
    {ID: 3, Username: 'user3', Name: 'User Three', Email: "user3@domain.dom", Role: 100},
  ];

  // Render the component
  var usersPage = shallow(
      <UserList users={users} token={token} />
  );

  expect(usersPage.find('section')).toHaveLength(1)
  expect(usersPage.find('UserRow')).toHaveLength(3)

  });

  it('displays error with no permissions', function() {

      // Render the component
      const component = shallow(
          <UserList />
      );

      expect(component.find('div')).toHaveLength(1)
      expect(component.find('AlertBox')).toHaveLength(1)
  })

  it('displays error with insufficient permissions', function() {

      // Render the component
      const component = shallow(
          <UserList token={tokenUser} />
      );

      expect(component.find('div')).toHaveLength(1)
      expect(component.find('AlertBox')).toHaveLength(1)
  })

});
