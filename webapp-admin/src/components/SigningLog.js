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
        logs: [],
        filterModels: [],
        filterString: '',
        message: null,
        expanded: {models: true},
        query: '',
        startRow: 0,
        endRow: PAGINATION_SIZE,
        totalLogs: 0,
        authorityID: this.props.selectedAccount.AuthorityID,
    };
    this.getSigningLogs(0, '', '')
    this.getSigningLogFilters()
  }

  handleExpansionClick = (value) => {
    var expanded = this.state.expanded;
    expanded[value] = !expanded[value]
    this.setState({expanded: expanded})
  }

  handleAccountChange = (authorityID) => {
    this.setState({authorityID: authorityID}, () => {
      this.getSigningLogs(0, '', '')
      this.getSigningLogFilters()  
    });
  }

  handleItemClick = (index, key) => {
    var items;
    if (key === 'models') {
      items = this.state.filterModels;
      items[index].selected = !items[index].selected;
      var filter = items.filter(i => i.selected === true).map(i => i.name).join(',')
      this.getSigningLogs(0, filter, this.state.query)
      this.setState({filterModels: items, filterString: filter});
    }
  }

  getSigningLogs = (offset, filter, serialnumber) => {
    SigningLogModel.listForAccount(this.state.authorityID, offset, filter, serialnumber).then((response) => {
      var data = JSON.parse(response.body);
      var message = null;
      if (!data.success) {
        message = data.message;
      }
      this.setState({logs: data.logs, message: message, totalLogs: data.total_count});
    });
  }

  getSigningLogFilters = () => {
    SigningLogModel.filters(this.state.authorityID).then((response) => {
      var data = JSON.parse(response.body);
      var message = "";
      if (!data.success) {
        message = data.message;
      }
      var filterModels = data.filters.models.map(function(item) {
        return {name: item, selected: false};
      });
      this.setState({filterModels: filterModels, message: message});
    });
  }

  handleSearchChange = (e) => {
    this.getSigningLogs(0, this.state.filterString, e.target.value)
    this.setState({query: e.target.value});
  }

  handleDownload = () => {
    SigningLogModel.download(this.state.authorityID, this.state.filterString, this.state.query);
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
                          name={T('models')} items={this.state.filterModels}
                          keyName={'models'}
                          handleItemClick={this.handleItemClick}
                          expanded={this.state.expanded.models}
                          expansionClick={this.handleExpansionClick}
                      />
                  </div>
                </div>
              </div>
              <div className="col-9">
                <Pagination totalLogs={this.state.totalLogs}
                            query={this.state.query}
                            searchText={T('find-serialnumber')}
                            pageChange={this.getSigningLogs}
                            filterString={this.state.filterString}
                            onDownload={this.handleDownload}
                            onSearchChange={this.handleSearchChange} />

                {this.renderTable(this.state.logs)}

              </div>
            </div>

          </section>

        </div>
    );

  }

}

export default SigningLog;