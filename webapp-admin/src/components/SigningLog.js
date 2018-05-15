/*
 * Copyright (C) 2016-2018 Canonical Ltd
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
        message: null,
        expanded: {models: true},
        query: '',
        startRow: 0,
        endRow: PAGINATION_SIZE,
    };
  }

  handleExpansionClick = (value) => {
    var expanded = this.state.expanded;
    expanded[value] = !expanded[value]
    this.setState({expanded: expanded})
  }

  handleItemClick = (index, key) => {
    this.setState({startRow: 0, endRow: PAGINATION_SIZE});
    this.props.onItemClick(index, key)
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

  filterRow(l, models) {

    // See if it passes the text search test
    if (this.state.query.length > 0) {
      if (l.serialnumber.toLowerCase().indexOf(this.state.query.toLowerCase()) < 0) return false
    }

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
    var models = this.selectedFilters(this.props.filterModels);

    return this.props.logs.filter((l) => {
      // Check if the row is filtered
      return this.filterRow(l, models);
    })
  }

  renderTable(items) {

    if (this.props.logs.length > 0) {
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
        <SigningLogRow key={l.id} log={l} />
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
                          name={T('models')} items={this.props.filterModels}
                          keyName={'models'}
                          handleItemClick={this.handleItemClick}
                          expanded={this.state.expanded.models}
                          expansionClick={this.handleExpansionClick}
                      />
                  </div>
                </div>
              </div>
              <div className="col-9">
                <Pagination rows={this.props.logs.length} displayRows={displayRows}
                            searchText={T('find-serialnumber')}
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