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
import React, {Component} from 'react';
import Navigation from './Navigation'
import NavigationUser from './NavigationUser'
import {T} from './Utils';

const LANGUAGES = {
    en: 'English',
    zh: 'Chinese'
}

class Header extends Component {

    constructor(props) {
        super(props)
        this.state = {
            language: window.AppState.getLocale()
        }
    }

    handleLanguageChange = (e) => {
        e.preventDefault();
        this.setState({language: e.target.value});
        window.AppState.setLocale(e.target.value);
        window.AppState.rerender();
    }

    renderLanguage(lang) {
        if (this.state.language === lang) {
            return (
                <button onClick={this.handleLanguageChange} value={lang} className="p-button--neutral">{LANGUAGES[lang]}</button>
            );
        } else {
            return (
                <button onClick={this.handleLanguageChange} value={lang}>{LANGUAGES[lang]}</button>
            );
        }
    }

    handleToggleMenu = (e) => {
        e.preventDefault();

        var navPrimary = document.querySelector('.p-navigation__nav');
        navPrimary.classList.toggle('show');
    }

    render() {

        return (
        <div>
            <header className="p-navigation p-navigation--light" role="banner">
                <div className="p-navigation__logo">
                        <a href="/" className="p-navigation__link">
                            <img src="/static/images/serial-vault-logo.svg" alt={T("title")} width="180px" height="44px" />
                        </a>
                </div>
                <a href="#navigation" className="p-navigation__toggle--open" title="menu" onClick={this.handleToggleMenu}>
                    <img src="/static/images/navigation-menu-plain.svg" width="30px" alt="menu" />
                </a>
                <nav className="p-navigation__nav">
                    <span className="u-off-screen"><a href="#navigation">Jump to site</a></span>
                    <Navigation token={this.props.token}
                        accounts={this.props.accounts} selectedAccount={this.props.selectedAccount} onAccountChange={this.props.onAccountChange} />
                    <NavigationUser token={this.props.token} />
                        {/*<form id="language-form" className="header-search">*/}
                                {/* Add more languages here */}
                                {/* this.renderLanguage('en') */}
                                {/* this.renderLanguage('zh') */}
                        {/*</form>*/}

                </nav>
            </header>

        </div>
        )
    }
}

export default Header;
