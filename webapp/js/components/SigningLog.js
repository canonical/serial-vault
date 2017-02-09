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
var Navigation = require('./Navigation');
var SigningLogRow = require('./SigningLogRow');
var Footer = require('./Footer');
var AlertBox = require('./AlertBox');
var SigningLogModel = require('../models/signinglog') 
import SigningLogFilter from './SigningLogFilter'
import Pagination from './Pagination'
import {T} from './Utils'

const PAGINATION_SIZE = 50;

var SigningLogList = React.createClass({
  getInitialState: function() {
    return {
        logs: this.props.logs || [],
        confirmDelete: null,
        message: null,
        makes: [],
        models: [],
        expanded: {},
        query: '',
        page: 1,
        startRow: 0,
        endRow: PAGINATION_SIZE,
    };
  },

  componentDidMount: function () {
    this.getLogs();
    this.getFilters();
  },

  getLogs: function () {
    var self = this;

    SigningLogModel.list().then(function(response) {
      var data = JSON.parse(response.body);
      var message = "";
      if (!data.success) {
        message = data.message;
      }

      self.setState({logs: data.logs, message: message});
    });
  },

  getFilters: function () {
    var self = this;

    SigningLogModel.filters().then(function(response) {
      var data = JSON.parse(response.body);
      var message = "";
      if (!data.success) {
        message = data.message;
      }

      var makes = data.filters.makes.map(function(item) {
            return {name: item, selected: false};
      });
      var models = data.filters.models.map(function(item) {
            return {name: item, selected: false};
      });

      self.setState({makes: makes, models: models, message: message});
    });
  },

  handleDelete: function(e) {
    e.preventDefault();
    this.setState({confirmDelete: parseInt(e.target.getAttribute('data-key'))});
  },

  handleDeleteLog: function(e) {
    e.preventDefault();
    var self = this;
    var logs = this.state.logs.filter(function(log) {
      return log.id === self.state.confirmDelete;
    });
    if (logs.length === 0) {
      return;
    }

    SigningLogModel.delete(logs[0]).then(function(response) {
      var data = JSON.parse(response.body);
      if ((response.statusCode >= 300) || (!data.success)) {
        self.setState({message: self.formatError(data)});
      } else {
        window.location = '/signinglog';
      }
    });
  },

  handleDeleteLogCancel: function(e) {
    e.preventDefault();
    this.setState({confirmDelete: null});
  },

  handleExpansionClick: function(value) {
    var expanded = this.state.expanded;
    expanded[value] = !expanded[value]
    this.setState({expanded: expanded})
  },

  handleItemClick: function(index, key) {
    var items = this.state[key];
    items[index].selected = !items[index].selected;
    this.setState({key: items, startRow: 0, endRow: PAGINATION_SIZE, page: 1});
  },

  handleRecordsForPage: function(startRow, endRow) {
    this.setState({startRow: startRow, endRow: endRow});
  },

  handleSearchChange: function(e) {
    this.setState({query: e.target.value});
  },

  handleDownload: function() {
    SigningLogModel.download(this.displayRows());
  },

  filterRow: function(l, makes, models) {

    // See if it passes the text search test
    if (this.state.query.length > 0) {
      if (l.serialnumber.toLowerCase().indexOf(this.state.query.toLowerCase()) < 0) return false
    }

    // If no filters are applied, then the row can be displayed
    if ((makes.length === 0) && (models.length === 0)) return true;

    // See if it passes the makes test
    if ((makes.length > 0) && (makes.indexOf(l.make) < 0)) return false

    // See if it passes the models test
    if ((models.length > 0) && (models.indexOf(l.model) < 0)) return false

    return true;
  },

  selectedFilters: function(name) {
    var items = [];
    name.map(function(n) {
      if(n.selected) {
        items.push(n.name);
      }
    });
    return items;
  },

  displayRows: function() {
    var self = this;
    var makes = this.selectedFilters(this.state.makes);
    var models = this.selectedFilters(this.state.models);

    return this.state.logs.filter(function(l) {
      // Check if the row is filtered
      return self.filterRow(l, makes, models);
    })
  },

  renderTable: function(items) {
    var self = this;

    if (this.state.logs.length > 0) {
      return (
        <div>
          <table>
            <thead>
              <tr>
                <th className="small"></th><th>{T('brand')}</th><th>{T('model')}</th><th>{T('serial-number')}</th><th>{T('revision')}</th><th>{T('fingerprint')}</th><th>{T('date')}</th>
              </tr>
            </thead>
            <tbody>
              {this.renderRows(items)}
            </tbody>
          </table>
        </div>
      );
    } else {
      return (
        <p>No models signed.</p>
      );
    }
  },

  renderRows: function(items) {
    return items.map((l) => {
      return (
        <SigningLogRow key={l.id} log={l} delete={this.handleDelete} confirmDelete={this.state.confirmDelete}
          deleteLog={this.handleDeleteLog} cancelDelete={this.handleDeleteLogCancel} />
      );
    });
  },

  render: function() {
    var displayRows = this.displayRows();

    return (
        <div className="inner-wrapper">
          <Navigation active="signinglog" />

          <section className="row no-border">
            <h2>{T('signinglog')}</h2>
            <div className="twelve-col">
              <p>{T('signinglog-description')}</p>
            </div>
            <div className="twelve-col">
              <AlertBox message={this.state.message} />
            </div>

            <div className="full">
              <div className="col three-col">
                <div className="box filter">
                  <div className="filter-section">
                      <h3>Filter By</h3>
                      <SigningLogFilter
                          name={T('makes')} items={this.state.makes}
                          keyName={'makes'}
                          handleItemClick={this.handleItemClick}
                          expanded={this.state.expanded.makes}
                          expansionClick={this.handleExpansionClick}
                      />
                      <SigningLogFilter
                          name={T('models')} items={this.state.models}
                          keyName={'models'}
                          handleItemClick={this.handleItemClick}
                          expanded={this.state.expanded.models}
                          expansionClick={this.handleExpansionClick}
                      />
                  </div>
                </div>
              </div>
              <div className="col nine-col last-col">
                <Pagination rows={this.state.logs.length} displayRows={displayRows}
                            page={this.state.page} searchText={T('find-serialnumber')}
                            pageChange={this.handleRecordsForPage}
                            onDownload={this.handleDownload}
                            onSearchChange={this.handleSearchChange} />
                {this.renderTable(displayRows.slice(this.state.startRow, this.state.endRow))}
              </div>
            </div>

          </section>

          <Footer />
        </div>
    );

  }

});

module.exports = SigningLogList;