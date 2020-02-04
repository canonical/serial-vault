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
import Adapter from 'enzyme-adapter-react-16';
import ReactTestUtils from 'react-dom/test-utils';
import { createRenderer } from 'react-test-renderer/shallow';
import SigningLog from '../components/SigningLog';
import {shallow, configure} from 'enzyme';

jest.dontMock('../components/SigningLog');
jest.dontMock('../components/Navigation');
jest.dontMock('../components/AlertBox');
jest.dontMock('../components/SigningLogFilter');
jest.dontMock('../components/SigningLogRow');
jest.dontMock('../components/Pagination');
jest.dontMock('../components/Utils');

configure({ adapter: new Adapter() });

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

const account = {ID: 1, AuthorityID: "canonical", Assertion: "123456abcdef"}
const token = { role: 200 }
const tokenUser = { role: 100 }

describe('signing-log list', function() {
    it('displays the signing logs page with no logs', function() {

        // Mock the data retrieval from the API
        var getLogs = jest.fn();
        var getFilters = jest.fn();
        SigningLog.prototype.getLogs = getLogs;
        SigningLog.prototype.getFilters = getFilters;

        // Shallow render the component
        var shallowRenderer = createRenderer();

        shallowRenderer.render(
            <SigningLog token={token} selectedAccount={{}} logs={[]} filterModels={[]} />
        );

        var logsPage = shallowRenderer.getRenderOutput();
        var section = logsPage.props.children;
        expect(section.props.children.length).toBe(4);
        var div = section.props.children[3];
        var para = div.props.children[1].props.children[1];
        expect(para.props.children).toBe('No models signed.')
    });

    it('displays the signing logs page with some logs', function() {

        // Set up a fixture for the model data
        var logs = [
            {id: 1, 'make': 'Brand1', model: 'Name1', 'serial': 'A11', fingerprint: 'a11'},
            {id: 2, 'make': 'Brand2', model: 'Name2', 'serial': 'A22', fingerprint: 'b22'},
            {id: 3, 'make': 'Brand3', model: 'Name3', 'serial': 'A33', fingerprint: 'c22'}
        ];

        // Render the component
        var component = shallow(
            <SigningLog token={token} selectedAccount={account} />
        );
        component.setState({ logs: logs });

        expect(component.find('section')).toHaveLength(1)
        expect(component.find('SigningLogRow')).toHaveLength(logs.length)
        expect(component.find('Pagination')).toHaveLength(1)
    });

    it('displays error with no permissions', function() {

        // Render the component
        const component = shallow(
            <SigningLog token={{}} selectedAccount={{}} />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
        expect(component.find('AlertBox').get(0).props.message).toBe('You do not have permissions to access this page')   
    })

    it('displays error with insufficient permissions', function() {

        // Render the component
        const component = shallow(
            <SigningLog token={tokenUser} selectedAccount={{}} />
        );

        expect(component.find('div')).toHaveLength(1)
        expect(component.find('AlertBox')).toHaveLength(1)
        expect(component.find('AlertBox').get(0).props.message).toBe('You do not have permissions to access this page')
    })
});