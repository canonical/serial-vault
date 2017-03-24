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
const sections = ['assertion']

export function sectionFromHash(hash) {
  return hash === '' ? 'home' : (
    sections.find(section => (
      hash.startsWith(`#${section}`)
    )) || ''
  )
}

export function routeFromHash(path) {
  const parts = path.split('/')

  return (parts[0] === '#device' && parts[1]) || ''
}

export function isError(response) {
  return ((response.statusCode >= 300) || (response.statusCode === 0));
}

export function handleError(err) {
  if (err.response) {
    this.setState({message: err.response.data.message})
  } else {
    this.setState({message: err.message})
  }
}