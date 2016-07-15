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

jest.dontMock('../components/SigningLog');
jest.dontMock('../components/SigningLogRow');


describe('signing-log list', function() {
    it('displays the signing logs page with no logs', function() {
        var SigningLog = require('../components/SigningLog');
        var IntlProvider = require('react-intl').IntlProvider;
        var Messages = require('../components/messages').en;

        // Mock the data retrieval from the API
        var getLogs = jest.genMockFunction();
        SigningLog.WrappedComponent.prototype.__reactAutoBindMap.getLogs = getLogs;

        // Render the component
        var logsPage = TestUtils.renderIntoDocument(
            <IntlProvider locale="en" messages={Messages}>
                <SigningLog />
            </IntlProvider>
        );

        expect(TestUtils.isCompositeComponent(logsPage)).toBeTruthy();

        // Check all the expected elements are rendered
        var section = TestUtils.findRenderedDOMComponentWithTag(logsPage, 'section');
        var h2s = TestUtils.scryRenderedDOMComponentsWithTag(logsPage, 'h2');
        expect(h2s.length).toBe(1);

        // Check the getLogs was called
        expect(getLogs.mock.calls.length).toBe(1);

        // Check the 'no models' message is rendered
        expect(section.children.length).toBe(4);
        expect(section.children[3].textContent).toBe('No models signed.');
    });

    it('displays the signing logs page with some logs', function() {
        var SigningLog = require('../components/SigningLog');
        var IntlProvider = require('react-intl').IntlProvider;
        var Messages = require('../components/messages').en;

        // Shallow render the component with the translations
        const intlProvider = new IntlProvider({locale: 'en', messages: Messages}, {});
        const {intl} = intlProvider.getChildContext();
        var shallowRenderer = TestUtils.createRenderer();

        // Set up a fixture for the model data
        var logs = [
            {id: 1, 'make': 'Brand1', model: 'Name1', 'serial': 'A11', fingerprint: 'a11'},
            {id: 2, 'make': 'Brand2', model: 'Name2', 'serial': 'A22', fingerprint: 'b22'},
            {id: 3, 'make': 'Brand3', model: 'Name3', 'serial': 'A33', fingerprint: 'c22'}
        ];

        // Mock the data retrieval from the API
        var getLogs = jest.genMockFunction();
        SigningLog.WrappedComponent.prototype.__reactAutoBindMap.getLogs = getLogs;

        // Render the component
        shallowRenderer.render(
            <SigningLog.WrappedComponent intl={intl} logs={logs} />
        );
        var logsPage = shallowRenderer.getRenderOutput();
        expect(logsPage.props.children.length).toBe(3);
        var section = logsPage.props.children[1];
        expect(section.props.children.length).toBe(4);

        // Check that the logs table is rendered correctly
        var tableDiv = section.props.children[3].props.children;
        var table = tableDiv.props.children[0];
        var tbody = table.props.children[1]
        expect(tbody.props.children.length).toBe(3); // data rows
        var row1 = tbody.props.children[0];

        expect(row1.type.WrappedComponent.displayName).toBe('SigningLogRow')
        expect(row1.props.log).toBe(logs[0])

    });
});