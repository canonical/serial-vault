import React, { Component } from 'react';
import Header from './components/Header'
import './App.css';

class App extends Component {
  render() {
    return (
      <div className="App">
        <Header />
        <div className="row">
          <h2>System User Assertion</h2>
          <div className="p-card">
            <h3 className="p-card__title">Usage</h3>
            <p className="p-card__content">
              A system-user assertion allows a local system user to be created on specified devices.<br />
              Once the assertion is created, download the file to a USB drive and insert it into the device.
            </p>
          </div>


        </div>
      </div>
    );
  }
}

export default App;
