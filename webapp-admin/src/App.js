/*
 * Copyright (C) 2017-2018 Canonical Ltd
 * License granted by Canonical Limited
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
import KeypairEdit from './components/KeypairEdit'
import KeypairGenerate from './components/KeypairGenerate'
import KeypairStore from './components/KeypairStore'
import AccountList from './components/AccountList'
import AccountForm from './components/AccountForm'
import AccountEdit from './components/AccountEdit'
import AccountKeyForm from './components/AccountKeyForm'
import Keypair from './components/Keypair'
import SigningLog from './components/SigningLog'
import SubstoreList from './components/SubstoreList'
import SystemUserForm from './components/SystemUserForm'
import NavigationSubmenu from './components/NavigationSubmenu';
import UserList from './components/UserList'
import UserEdit from './components/UserEdit'
import Accounts from './models/accounts'
import Keypairs from './models/keypairs'
import Models from './models/models';
import {sectionFromPath, sectionIdFromPath, subSectionIdFromPath, isLoggedIn, getAccount, saveAccount, isUserAdmin, isUserSuperuser, formatError} from './components/Utils'
import createHistory from 'history/createBrowserHistory'
import './sass/App.css'

const history = createHistory()
const submenuModels = ['models','substores','systemuser']

class App extends Component {
  constructor(props) {
    super(props)
    this.state = {
      location: history.location,
      token: props.token || {},
      models: [],
      accounts: [],
      keypairs: [],
      substores: [],
      selectedAccount: getAccount() || {},
    }

    history.listen(this.handleNavigation.bind(this))
    this.signinglog = React.createRef();
    this.getAccounts()
  }

  handleNavigation(location) {
    this.setState({ location: location })
    window.scrollTo(0, 0)
  }

  getAccounts() {
    if (isLoggedIn(this.props.token)) {
      Accounts.list().then((response) => {
          var data = JSON.parse(response.body);
          var message = "";
          if (!data.success) {
              message = data.message;
          }

          var selectedAccount = this.state.selectedAccount;
          if ((!this.state.selectedAccount.ID) && (!getAccount().AuthorityID)) {
            // Set to the first in the account list
            if (data.accounts.length > 0) {
              selectedAccount = data.accounts[0]
              saveAccount(selectedAccount)
            }
          } else {
            // Refresh the current account details
            if (data.accounts.length > 0) {
              var accs = data.accounts.filter((a) => {
                return this.state.selectedAccount.ID === a.ID
              })
              if (accs.length > 0) {
                selectedAccount = accs[0]
                saveAccount(selectedAccount)
              }
            }
          }

          this.updateDataForRoute(selectedAccount)
          this.setState({accounts: data.accounts, selectedAccount: selectedAccount, message: message});
      });
    }
  }

  getKeypairs(authorityID) {
    Keypairs.list().then((response) => {
        var data = JSON.parse(response.body);
        var message = "";
        if (!data.success) {
            message = data.message;
        }

        var keypairs = data.keypairs.filter((k) => {
            return k.AuthorityID === authorityID;
        })

        this.setState({keypairs: keypairs, message: message});
    });
  }

  getModels(authorityID) {
    Models.list().then((response) => {
      var data = JSON.parse(response.body);
      var message = "";
      if (!data.success) {
        message = data.message;
      }

      var models = data.models.filter((m) => {
        return m['brand-id'] === authorityID;
      })

      this.setState({models: models, message: message});
    });
  }

  getSubstores(accountId) {
    Accounts.stores(accountId).then((response) => {
        var data = JSON.parse(response.body);

        if (response.statusCode >= 300) {
            this.setState({message: formatError(data), hideForm: true});
        } else {
            this.setState({substores: data.substores, message: null});
        }
    });
  }

  updateDataForRoute(selectedAccount) {
    var currentSection = sectionFromPath(window.location.pathname);
    if (currentSection === 'signinglog') {
      this.signinglog.current.handleAccountChange(selectedAccount.AuthorityID);
    }
    if(currentSection==='accounts') {this.getKeypairs(selectedAccount.AuthorityID)}
    if(currentSection==='signing-keys') {this.getKeypairs(selectedAccount.AuthorityID)}
    if(currentSection==='models') {
      this.getModels(selectedAccount.AuthorityID)
      this.getKeypairs(selectedAccount.AuthorityID)
    }
    if(currentSection==='substores') {
      if (selectedAccount.ID) {
        this.getSubstores(selectedAccount.ID)
      }
      this.getModels(selectedAccount.AuthorityID)
    }
    if(currentSection==='systemuser') {this.getModels(selectedAccount.AuthorityID)}
  }

  handleAccountChange = (account) => {
    saveAccount(account)
    this.setState({selectedAccount: account})

    this.updateDataForRoute(account)
  }

  renderModels() {
    const id = sectionIdFromPath(window.location.pathname, 'models')
  
    switch(id) {
      case 'new':
        return <ModelEdit token={this.props.token} id={null} selectedAccount={this.state.selectedAccount} keypairs={this.state.keypairs} />
      case '':
        return <ModelList token={this.props.token} selectedAccount={this.state.selectedAccount} models={this.state.models} />
      default:
        return <ModelEdit token={this.props.token} id={id} selectedAccount={this.state.selectedAccount} keypairs={this.state.keypairs} />
    }
  }

  renderAccounts() {
    const id = sectionIdFromPath(window.location.pathname, 'accounts')
    const sub = subSectionIdFromPath(window.location.pathname, 'accounts')

    switch(id) {
      case 'upload':
        return <AccountForm token={this.props.token} />
      case 'new':
        return <AccountEdit token={this.props.token} />
      case 'account':
        return <AccountEdit id={sub} token={this.props.token} />
      case 'key-assertion':
        return <AccountKeyForm token={this.props.token} />
      default:
        return <AccountList token={this.props.token} selectedAccount={this.state.selectedAccount} keypairs={this.state.keypairs} />
    }
  }

  renderUsers() {
    const id = sectionIdFromPath(window.location.pathname, 'users')

    switch(id) {
      case 'new':
        return <UserEdit token={this.props.token} />
      case '':
        const urlParams = new URLSearchParams(window.location.search);
        const query = urlParams.get('query');
        return <UserList token={this.props.token} query={query} />
      default:
        return <UserEdit token={this.props.token} id={id} />
    }
  }

  renderKeypairs() {
    const id = sectionIdFromPath(window.location.pathname, 'signing-keys')

    switch(id) {
      case 'generate':
        return <KeypairGenerate token={this.props.token} selectedAccount={this.state.selectedAccount} />
      case 'new':
        return <KeypairAdd token={this.props.token} selectedAccount={this.state.selectedAccount} />
      case '':
        return <Keypair token={this.props.token} selectedAccount={this.state.selectedAccount} keypairs={this.state.keypairs} onRefresh={this.handleAccountChange} />
      case 'store':
        return <KeypairStore token={this.props.token} selectedAccount={this.state.selectedAccount} keypairs={this.state.keypairs} />
      default:
        return <KeypairEdit token={this.props.token} id={id} />
    }
  }

  render() {
    var currentSection = sectionFromPath(window.location.pathname);

    return (
      <div className="App">
          <Header token={this.props.token}
            accounts={this.state.accounts} selectedAccount={this.state.selectedAccount} 
            onAccountChange={this.handleAccountChange} />

          <div className="spacer" />

          {(isUserAdmin(this.props.token)||isUserSuperuser(this.props.token)) &&
           (currentSection==='models'||currentSection==='substores'||currentSection==='systemuser')? 
            <section className="row">
              <NavigationSubmenu items={submenuModels} selected={currentSection} />
            </section>
          : ''}
  
          {currentSection==='home'? <Index token={this.props.token} /> : ''}
          {currentSection==='notfound'? <Index token={this.props.token} error /> : ''}

          {currentSection==='signing-keys'? this.renderKeypairs(): ''}
          {currentSection==='models'? this.renderModels() : ''}

          {currentSection==='accounts'? this.renderAccounts() : ''}
          {currentSection==='signinglog'? <SigningLog 
              ref={this.signinglog}
              token={this.props.token} 
              onAccountChange={this.handleAccountChange}
              selectedAccount={this.state.selectedAccount} /> : ''}

          {currentSection==='substores'? <SubstoreList token={this.props.token}
            selectedAccount={this.state.selectedAccount} onRefresh={this.handleAccountChange}
            substores={this.state.substores} models={this.state.models} /> : ''}
          {currentSection==='systemuser'? <SystemUserForm token={this.props.token} models={this.state.models} /> : ''}

          {currentSection==='users'? this.renderUsers() : ''}

          <Footer />
      </div>
    )
  }
}

export default App;
