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
import React, {Component} from 'react'

class Header extends Component {

    render() {
        return (
            <header className="p-navigation" role="banner">
                <div className="row">
                    <div className="p-navigation__logo">
                        <div className="nav_logo">
                            <img src="/static/images/logo-ubuntu-white.svg" alt="Ubuntu" />
                            <span>Serial Vault</span>
                        </div>
                    </div>
                    <a href="#navigation" className="p-navigation__toggle--open" title="menu">Menu</a>
                    <a href="#navigation-closed" className="p-navigation__toggle--close" title="close menu">Menu</a>
                    <nav className="p-navigation__nav">
                        <span className="u-off-screen">
                            <a href="#main-content">Jump to main content</a>
                        </span>
                        <ul className="p-navigation__links">
                        </ul>
                    </nav>
                </div>
            </header>
        )
    }
}

export default Header