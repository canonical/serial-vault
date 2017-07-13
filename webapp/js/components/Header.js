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
'use strict'
var React = require('react');
import Navigation from './Navigation'
import {T} from './Utils';

const LANGUAGES = {
    en: 'English',
    zh: 'Chinese'
}

var Header = React.createClass({
    getInitialState: function() {
        return {language: window.AppState.getLocale()}
    },

    handleLanguageChange: function(e) {
        e.preventDefault();
        this.setState({language: e.target.value});
        window.AppState.setLocale(e.target.value);
        window.AppState.rerender();
    },

    renderLanguage: function(lang) {
        if (this.state.language === lang) {
            return (
                <button onClick={this.handleLanguageChange} value={lang} className="p-button--neutral">{LANGUAGES[lang]}</button>
            );
        } else {
            return (
                <button onClick={this.handleLanguageChange} value={lang}>{LANGUAGES[lang]}</button>
            );
        }
    },

  render: function() {

    return (
      <div>
        <header className="p-navigation--light" role="banner">
            <div className="row">
                <div className="p-navigation__logo">
                        <div className="nav_logo">
                            <a href="/" className="p-navigation__link">
                                <img src="/static/images/logo-ubuntu-black.svg" alt="Ubuntu" height="20px"/>
                                <span>{T("title")}</span>
                            </a>
                        </div>
                </div>

                <nav role="navigation" className="p-navigation__nav">
                    <span className="u-off-screen"><a href="#navigation">Jump to site nav</a></span>
                    <Navigation token={this.props.token} />
                        {/*<form id="language-form" className="header-search">*/}
                                {/* Add more languages here */}
                                {/* this.renderLanguage('en') */}
                                {/* this.renderLanguage('zh') */}
                        {/*</form>*/}

                </nav>
            </div>
        </header>

      </div>
    )
  }
});

module.exports = Header;
