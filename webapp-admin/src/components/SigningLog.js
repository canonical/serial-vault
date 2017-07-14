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

import React, {Component} from 'react';
import SigningLogRow from './SigningLogRow';
import AlertBox from './AlertBox';
import SigningLogModel from '../models/signinglog' 
import SigningLogFilter from './SigningLogFilter'
import Pagination from './Pagination'
import {T, isUserAdmin} from './Utils'

const PAGINATION_SIZE = 50;

class SigningLog extends Component {
  constructor(props) {

    super(props)
    this.state = {
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
  }

  componentDidMount() {
    this.getLogs();
    this.getFilters();
  }

  getLogs() {
    var self = this;

    SigningLogModel.list().then(function(response) {
      var data = JSON.parse(response.body);
      var message = "";
      if (!data.success) {
        message = data.message;
      }

      self.setState({logs: data.logs, message: message});
    });
  }

  getFilters() {
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
  }

  handleDelete = (e) => {
    e.preventDefault();
    this.setState({confirmDelete: parseInt(e.target.getAttribute('data-key'), 10)});
  }

  handleDeleteLog = (e) => {
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
  }

  handleDeleteLogCancel = (e) => {
    e.preventDefault();
    this.setState({confirmDelete: null});
  }

  handleExpansionClick = (value) => {
    var expanded = this.state.expanded;
    expanded[value] = !expanded[value]
    this.setState({expanded: expanded})
  }

  handleItemClick = (index, key) => {
    var items = this.state[key];
    items[index].selected = !items[index].selected;
    this.setState({key: items, startRow: 0, endRow: PAGINATION_SIZE, page: 1});
  }

  handleRecordsForPage = (startRow, endRow) => {
    this.setState({startRow: startRow, endRow: endRow});
  }

  handleSearchChange = (e) => {
    this.setState({query: e.target.value});
  }

  handleDownload = () => {
    SigningLogModel.download(this.displayRows());
  }

  filterRow(l, makes, models) {

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
  }

  selectedFilters(name) {
    var items = [];
    name.map(function(n) {
      if(n.selected) {
        items.push(n.name);
      }
      return n.name
    });
    return items;
  }

  displayRows() {
    var self = this;
    var makes = this.selectedFilters(this.state.makes);
    var models = this.selectedFilters(this.state.models);

    return this.state.logs.filter(function(l) {
      // Check if the row is filtered
      return self.filterRow(l, makes, models);
    })
  }

  renderTable(items) {

    if (this.state.logs.length > 0) {
      return (
        <div>
          <table>
            <thead>
              <tr>
                <th>{T('brand')}</th><th>{T('model')}</th><th>{T('serial-number')}</th><th>{T('revision')}</th><th>{T('fingerprint')}</th><th>{T('date')}</th>
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
  }

  renderRows(items) {
    return items.map((l) => {
      return (
        <SigningLogRow key={l.id} log={l} delete={this.handleDelete} confirmDelete={this.state.confirmDelete}
          deleteLog={this.handleDeleteLog} cancelDelete={this.handleDeleteLogCancel} />
      );
    });
  }

  render() {

    if (!isUserAdmin(this.props.token)) {
      return (
        <div className="row">
          <AlertBox message={T('error-no-permissions')} />
        </div>
      )
    }

    var displayRows = this.displayRows();

    return (
        <div className="row">

          <section className="row">
            <h2>{T('signinglog')}</h2>
            <div className="col-12">
              <p>{T('signinglog-description')}</p>
            </div>
            <div className="col-12">
              <AlertBox message={this.state.message} />
            </div>

            <div className="row">
              <div className="col-3">
                <div className="p-card filter">
                  <div className="filter-section">
                      <h4>Filter By</h4>
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
              <div className="col-9">
                <Pagination rows={this.state.logs.length} displayRows={displayRows}
                            page={this.state.page} searchText={T('find-serialnumber')}
                            pageChange={this.handleRecordsForPage}
                            onDownload={this.handleDownload}
                            onSearchChange={this.handleSearchChange} />
                {this.renderTable(displayRows.slice(this.state.startRow, this.state.endRow))}
              </div>
            </div>

          </section>

        </div>
    );

  }

}

export default SigningLog;