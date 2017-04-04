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
import React from 'react'
import {shallow} from 'enzyme';
import SystemUserForm from '../components/SystemUserForm'

const MODELS = [
    {"id":1,"brand-id":"mXFtdcPlBkiZ","model":"acme-pi3","type":"device","keypair-id":4,"authority-id":"mXFtdcPlBkiZ","key-id":"81ZcPubBKI0UsePBVL","key-active":true},
    {"id":2,"brand-id":"developer1","model":"router","type":"device","keypair-id":2,"authority-id":"developer1","key-id":"vREVhx0tSfYC6KkFHmLW","key-active":true},
]

describe('system-user form', function() {
    it('displays the system-user form', function() {

        // Render the component
        const component = shallow(
            <SystemUserForm models={[]} />
        );

        expect(component.find('form')).toHaveLength(1)
        expect(component.find('fieldset')).toHaveLength(1)
        expect(component.find('input')).toHaveLength(5)
        expect(component.find('select')).toHaveLength(1)
        expect(component.find('button')).toHaveLength(1)
    })

    it('displays the system-user form with models', function() {

        // Render the component
        const component = shallow(
            <SystemUserForm models={MODELS} />
        );

        expect(component.find('form')).toHaveLength(1)

       let select = component.find('select')
       expect(select).toHaveLength(1)

       let options = select.find('option')
       expect(options).toHaveLength(MODELS.length + 1)

       expect(options.nodes[0].props.value).toEqual(0)
       expect(options.nodes[1].props.value).toEqual(MODELS[0].id)
       expect(options.nodes[2].props.value).toEqual(MODELS[1].id)
       expect(options.nodes[2].props.children).toEqual([MODELS[1]['brand-id'], ' ', MODELS[1].model])
    })

})
