/*
 * Copyright (C) 2017-2018 Canonical Ltd
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

import React from 'react';
import ReactDOM from 'react-dom';
import TestUtils from 'react-addons-test-utils';
import Pagination from '../components/Pagination'

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};


function generateRows() {
    var rows = []
    for (var i=0; i<150; i++) {
        rows.push({id: i, name: 'Row ' + i})
    }
    return rows;
}

describe('pagination', function() {
    it('displays the pagination component with no rows', function() {

        // Render the component
        var component = TestUtils.renderIntoDocument(
            <Pagination page={1} displayRows={[]} />
        );

        expect(TestUtils.isCompositeComponent(component)).toBeTruthy();

        // Just the download button shown
        var buttons = TestUtils.scryRenderedDOMComponentsWithTag(component, 'button');
        expect(buttons.length).toBe(1)
        expect(buttons[0].textContent).toBe('Download')
    })

    it('displays the pagination component with with page count', function() {

        // Render the component
        var component = TestUtils.renderIntoDocument(
            <Pagination page={1} displayRows={generateRows()} />
        );

        expect(TestUtils.isCompositeComponent(component)).toBeTruthy();

        // Paging buttons shown
        var buttons = TestUtils.scryRenderedDOMComponentsWithTag(component, 'button');
        expect(buttons.length).toBe(3)
        expect(buttons[0].textContent).toBe('«')
        expect(buttons[1].textContent).toBe('»')
        expect(buttons[2].textContent).toBe('Download')

        var span = TestUtils.findRenderedDOMComponentWithTag(component, 'span');
        expect(span.textContent).toContain('1 of 3')
    })

})
