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
import React, { Component } from 'react'
import {T} from './Utils'

class Pagination extends Component {

  constructor(props) {
    super(props)

    this.state = {
      page: 1,
      query: null,
      maxRecords: 50,
    }
  }

  pageUp = () => {
    var pages = this.calculatePages();
    var page = this.state.page + 1;
    if (page > pages) {
        page = pages;
    }
    this.setState({page: page});
    this.signalPageChange(page);
  }

  pageDown = () => {
    var page = this.state.page - 1;
    if (page <= 0) {
        page = 1;
    }
    this.setState({page: page});
    this.signalPageChange(page);
  }

  signalPageChange(page) {
    // Signal the rows that the owner should display
    var startRow = ((page - 1) * this.state.maxRecords);

    this.props.pageChange(startRow, startRow + this.state.maxRecords);
  }

  calculatePages() {
    // Use the filtered row count when we a query has been entered
    var length = this.props.displayRows.length;

    var pages = parseInt(length / this.state.maxRecords, 10);
    if (length % this.state.maxRecords > 0) {
        pages += 1;
    }

    return pages;
  }

  renderPaging() {
    var pages = this.calculatePages();
    if (pages > 1) {
        return (
            <div className="u-float--right">
                <button className="p-button--neutral" href="" onClick={this.pageDown}>&laquo;</button>
                <span>&nbsp;{this.state.page} of {pages}&nbsp;</span>
                <button className="p-button--neutral" href="" onClick={this.pageUp}>&raquo;</button>
            </div>
        );
    } else {
        return <div className="col-3"></div>;
    }
  }

  render() {
    return (
        <div className="col-12">
            <div className="row">
                <div className="col-9">
                    <input type="search" placeholder={this.props.searchText} onChange={this.props.onSearchChange}
                                value={this.props.query} />
                </div>
                {this.renderPaging()}
            </div>
            <div className="row spacer">
                <div className="col-6">
                    <button className="p-button--brand" onClick={this.props.onDownload}>{T('download')}</button>
                </div>
            </div>
        </div>
    );
  }
}

export default Pagination
