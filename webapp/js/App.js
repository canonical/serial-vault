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
import Index from './components/Index'
import ModelList from './components/ModelList'
import ModelEdit from './components/ModelEdit'
import KeypairAdd from './components/KeypairAdd'
import AccountList from './components/AccountList'
import AccountForm from './components/AccountForm'
import AccountKeyForm from './components/AccountKeyForm'
import SigningLog from './components/SigningLog'
import SystemUserForm from './components/SystemUserForm'
import AlertBox from './components/AlertBox'
import {sectionFromPath, sectionIdFromPath} from './components/Utils'
import createHistory from 'history/createBrowserHistory'


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

    switch(id) {
      case 'new':
        return <ModelEdit token={this.props.token} />
      case '':
        return <ModelList token={this.props.token} />
      default:
        return <ModelEdit token={this.props.token} id={id} />
    }
  }

  renderAccounts() {
    const id = sectionIdFromPath(location.pathname, 'accounts')

    switch(id) {
      case 'new':
        return <AccountForm token={this.props.token} />
      case 'key-assertion':
        return <AccountKeyForm token={this.props.token} />
      default:
        return <AccountList token={this.props.token} />
    }
  }


  render() {

    var currentSection = sectionFromPath(location.pathname);

    return (
      <div className="App">
          <Header token={this.props.token} />

          <div className="spacer" />
  
          {currentSection==='home'? <Index token={this.props.token} /> : ''}
          {currentSection==='notfound'? <Index token={this.props.token} error={true} /> : ''}

          {currentSection==='models'? this.renderModels() : ''}
          {currentSection==='keypairs'? <KeypairAdd token={this.props.token} />: ''}

          {currentSection==='accounts'? this.renderAccounts() : ''}
          {currentSection==='signinglog'? <SigningLog token={this.props.token} /> : ''}
          {currentSection==='signinglog'? <SigningLog token={this.props.token} /> : ''}

          {currentSection==='systemuser'? <SystemUserForm token={this.props.token} /> : ''}

          <Footer />
      </div>
    )
  }
}

export default App;