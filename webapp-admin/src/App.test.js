import React from 'react';
import ReactDOM from 'react-dom';
import App from './App';
import Messages from './components/messages';
import {IntlProvider} from 'react-intl';

// Mock the AppState method for locale
window.AppState = {getLocale: function() {return 'en'}};

const token = { role: 200, name: 'Steven Vault' }
const tokenUser = { role: 100, name: 'Steven Vault' }

it('renders without crashing', () => {
  const div = document.createElement('div');
  ReactDOM.render((
    <IntlProvider locale={'en'} messages={Messages['en']}>
      <App token={token} />
    </IntlProvider>
  ), div);
});
