/*
 * Copyright (C) 2016-2020 Canonical Ltd
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
"use strict";

import React from "react";
import ReactTestUtils from "react-dom/test-utils";
import Adapter from "enzyme-adapter-react-16";
import { configure } from "enzyme";
import SystemUserForm from "../components/SystemUserForm";

jest.dontMock("../components/SystemUserForm");

configure({ adapter: new Adapter() });

const token = { role: 300 };

window.AppState = {
  getLocale: function () {
    return "en";
  },
};

describe("create system user", function () {
  it("displays the system user edit page for create", function () {
    // Render the component
    var systemUserPage = ReactTestUtils.renderIntoDocument(
      <SystemUserForm token={token} models={[]} />
    );
    expect(ReactTestUtils.isCompositeComponent(systemUserPage)).toBeTruthy();

    // find the header for this page
    var h2 = ReactTestUtils.findRenderedDOMComponentWithTag(
      systemUserPage,
      "h2"
    );
    expect(h2.textContent).toBe("Create System-User");

    // find all input fields for this page
    var input = ReactTestUtils.scryRenderedDOMComponentsWithTag(
      systemUserPage,
      "input"
    );
    expect(input.length).toBe(6);

    expect(input[0].getAttribute("name")).toBe("email");
    expect(input[1].getAttribute("name")).toBe("username");
    expect(input[2].getAttribute("name")).toBe("password");
    expect(input[3].getAttribute("name")).toBe("name");
    expect(input[4].getAttribute("name")).toBe("since_date_time");
    expect(input[5].getAttribute("name")).toBe("until_date_time");
  });
});

describe("create system user with serial-number", function () {
  it("displays the system user edit page for create", function () {
    // Render the component
    var systemUserPage = ReactTestUtils.renderIntoDocument(
      <SystemUserForm token={token} models={[]} />
    );
    expect(ReactTestUtils.isCompositeComponent(systemUserPage)).toBeTruthy();

    // find the header for this page
    var h2 = ReactTestUtils.findRenderedDOMComponentWithTag(
      systemUserPage,
      "h2"
    );
    expect(h2.textContent).toBe("Create System-User");

    var buttons = ReactTestUtils.scryRenderedDOMComponentsWithTag(
      systemUserPage,
      "button"
    );

    expect(buttons[0].getAttribute("title")).toBe("Add serial number");
    ReactTestUtils.Simulate.click(buttons[0]);
    
    // find all input fields for this page
    var input = ReactTestUtils.scryRenderedDOMComponentsWithTag(
      systemUserPage,
      "input"
    );
    expect(input.length).toBe(7);

    expect(input[0].getAttribute("name")).toBe("email");
    expect(input[1].getAttribute("name")).toBe("username");
    expect(input[2].getAttribute("name")).toBe("password");
    expect(input[3].getAttribute("name")).toBe("name");
    expect(input[4].getAttribute("name")).toBe("serials");
    expect(input[5].getAttribute("name")).toBe("since_date_time");
    expect(input[6].getAttribute("name")).toBe("until_date_time");
  });
});
