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
import KeypairGenerate from './components/KeypairGenerate'
import AccountList from './components/AccountList'
import AccountForm from './components/AccountForm'
import AccountKeyForm from './components/AccountKeyForm'
import Keypair from './components/Keypair'
import SigningLog from './components/SigningLog'
import SystemUserForm from './components/SystemUserForm'
import UserList from './components/UserList'
import UserEdit from './components/UserEdit'
import {sectionFromPath, sectionIdFromPath} from './components/Utils'
import createHistory from 'history/createBrowserHistory'
import './sass/App.css'

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
    const id = sectionIdFromPath(window.location.pathname, 'models')

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
    const id = sectionIdFromPath(window.location.pathname, 'accounts')

    switch(id) {
      case 'new':
        return <AccountForm token={this.props.token} />
      case 'key-assertion':
        return <AccountKeyForm token={this.props.token} />
      default:
        return <AccountList token={this.props.token} />
    }
  }

  renderUsers() {
    const id = sectionIdFromPath(window.location.pathname, 'users')

    switch(id) {
      case 'new':
        return <UserEdit token={this.props.token} />
      case '':
        return <UserList token={this.props.token} />
      default:
        return <UserEdit token={this.props.token} id={id} />
    }
  }

  renderKeypairs() {
    const id = sectionIdFromPath(window.location.pathname, 'signing-keys')

    switch(id) {
      case 'generate':
        return <KeypairGenerate token={this.props.token} />
      case 'new':
        return <KeypairAdd token={this.props.token} />
      default:
        return <Keypair token={this.props.token} />
    }
  }

  render() {

    var currentSection = sectionFromPath(window.location.pathname);

    return (
      <div className="App">
          <Header token={this.props.token} />

          <div className="spacer" />
  
          {currentSection==='home'? <Index token={this.props.token} /> : ''}
          {currentSection==='notfound'? <Index token={this.props.token} error={true} /> : ''}

          {currentSection==='signing-keys'? this.renderKeypairs(): ''}
          {currentSection==='models'? this.renderModels() : ''}

          {currentSection==='accounts'? this.renderAccounts() : ''}
          {currentSection==='signinglog'? <SigningLog token={this.props.token} /> : ''}

          {currentSection==='systemuser'? <SystemUserForm token={this.props.token} /> : ''}

          {currentSection==='users'? this.renderUsers() : ''}

          <Footer />
      </div>
    )
  }
}

export default App;