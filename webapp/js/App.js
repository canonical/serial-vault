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
import React, { Component } from 'react';
import Header from './components/Header'
import Footer from './components/Footer'
import If from './components/If'
import Index from './components/Index'
import ModelList from './components/ModelList'
import ModelEdit from './components/ModelEdit'
import AccountList from './components/AccountList'
import SigningLog from './components/SigningLog'
import AlertBox from './components/AlertBox'
import {sectionFromPath, sectionIdFromPath} from './components/Utils'
import createHistory from 'history/createBrowserHistory'
//import './sass/App.css';


const history = createHistory()

class App extends Component {
  constructor(props) {
    super(props)
    this.state = {
      location: history.location,
      token: props.token || {},
    }

    history.listen(this.handleNavigation.bind(this))
  }

  handleNavigation(location) {
    this.setState({ location: location })
    window.scrollTo(0, 0)
  }

  renderModels() {
    const id = sectionIdFromPath(location.pathname, 'models')
    console.log('id', id)

    switch(id) {
      case 'new':
        return <ModelEdit token={this.props.token} />
      case '':
        return <ModelList token={this.props.token} />
      default:
        return <ModelEdit token={this.props.token} id={id} />
    }

  }

  render() {

    var currentSection = sectionFromPath(location.pathname);
    console.log('currentSection', currentSection)

    return (
      <div className="App">
          <Header token={this.props.token} />
  
          {currentSection==='home'? <Index token={this.props.token} /> : ''}

          {currentSection==='models'? this.renderModels() : ''}

          {currentSection==='accounts'? <AccountList token={this.props.token} /> : ''}
          {currentSection==='signinglog'? <SigningLog token={this.props.token} /> : ''}

          <Footer />
      </div>
    )
  }
}

export default App;