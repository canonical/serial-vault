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
var React = require('react');
var Router = require('react-router').Router;
var render = require('react-dom').render;
var Route = require('react-router').Route;
var Link = require('react-router').Link;
var browserHistory = require('react-router').browserHistory;
var Index = require('./components/Index');
var ModelList = require('./components/ModelList');
var ModelEdit = require('./components/ModelEdit');


render((
  <Router history={browserHistory}>
    <Route path="/" component={Index} />
    <Route path="/models" component={ModelList} />
    <Route path="/models/new" component={ModelEdit} />
    <Route path="/models/:id/edit" component={ModelEdit} />
    <Route path="*" component={Index} />
  </Router>
), document.getElementById('main'))
