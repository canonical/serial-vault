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

class Pagination extends React.Component {

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

    var pages = parseInt(length / this.state.maxRecords);
    if (length % this.state.maxRecords > 0) {
        pages += 1;
    }

    return pages;
  }

  renderPaging() {
    var pages = this.calculatePages();
    if (pages > 1) {
        return (
            <div className="three-col last-col right">
                <button className="button--secondary small" href="" onClick={this.pageDown}>&laquo;</button>
                &nbsp;{this.state.page} of {pages}&nbsp;
                <button className="button--secondary small" href="" onClick={this.pageUp}>&raquo;</button>
            </div>
        );
    } else {
        return <div className="three-col last-col right"></div>;
    }
  }

  render() {
    return (
        <div className="nine-col pagination">
            <div>
                <div className="six-col">
                    <div>
                        <input type="search" placeholder={this.props.searchText} onChange={this.props.onSearchChange}
                                value={this.props.query} />
                    </div>
                </div>
                {this.renderPaging()}
            </div>
            <div>
                <button className="button--secondary" onClick={this.props.onDownload}>{T('download')}</button>
            </div>
        </div>
    );
  }
}

export default Pagination
