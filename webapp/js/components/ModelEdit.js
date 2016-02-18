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
var AlertBox = require('./AlertBox');
var Navigation = require('./Navigation');
var Footer = require('./Footer');

var ModelEdit = React.createClass({
	getInitialState: function() {
		return {title: null};
	},

	componentDidMount: function() {
		if (this.props.params.id) {
			this.setState({title: "Edit Model"});
		} else {
			this.setState({title: "New Model"});
		}
	},

	render: function() {
		return (
			<div>
				<Navigation active="models" />

				<section className="row">
					  <h2>{this.state.title}</h2>
						<AlertBox message="Not implemented yet!" />
				</section>

				<section className="row no-border">
					<a href='/models' className="button--secondary">Cancel</a>
				</section>

				<Footer />
			</div>
		)
	}
});

module.exports = ModelEdit;
