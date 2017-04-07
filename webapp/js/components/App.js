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
var injectIntl = require('react-intl').injectIntl;
import Navigation from './Navigation'
import Footer from './Footer'


const LANGUAGES = {
	en: 'English',
	zh: 'Chinese'
}

var App = React.createClass({
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
		var M = this.props.intl.formatMessage;

    return (
      <div>
				<header className="p-navigation" role="banner">
					<div className="row">
						<div className="p-navigation__logo">
								<div className="nav_logo">
										<img src="/static/images/logo-ubuntu-white.svg" alt="Ubuntu" />
										<span>{M({id:"title"})}</span>
								</div>
						</div>

						<nav role="navigation" className="p-navigation__nav">
							<span className="u-off-screen"><a href="#navigation">Jump to site nav</a></span>
							<Navigation />
								{/*<form id="language-form" className="header-search">*/}
										{/* Add more languages here */}
										{/* this.renderLanguage('en') */}
										{/* this.renderLanguage('zh') */}
								{/*</form>*/}

						</nav>
					</div>
				</header>


				<div className="wrapper">

					{this.props.children}
				</div>

				<Footer />
      </div>
    )
  }
});

module.exports = injectIntl(App);
