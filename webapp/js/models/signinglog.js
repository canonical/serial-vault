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
var Ajax = require('./Ajax');

var SigningLog = {
    url: 'signinglog',

	list: function (fromID) {
		var data = {};
		if (fromID) {
			data.fromID = fromID;
		}
		return Ajax.get(this.url, data);
	},

	filters: function() {
		return Ajax.get(this.url + '/filters');
	},

	delete:  function(log) {
		return Ajax.delete(this.url + '/' + log.id, {});
	},

	download: function(data) {
		// Convert the filtered data to a CSV array format
		var lines = ['data:text/csv;charset=utf-8,ID,Make,Model,Serial Number,Revision,Fingerprint,Date'];

		data.forEach(function(d, index){
			var line = [d.id,d.make,d.model,d.serialnumber,d.revision,d.fingerprint,d.created].join(",");
			lines.push(line);
		});

		// Convert the lines array to a downloadable URI
		var csvContent = lines.join("\n");
		var encodedUri = encodeURI(csvContent);

		// Create a temporary link so we can name the download file
		var downloadLink = document.createElement("a");
		downloadLink.href = encodedUri;
		downloadLink.download = "SigningLogs.csv";

		// Click the link and remove it
		document.body.appendChild(downloadLink);
		downloadLink.click();
		document.body.removeChild(downloadLink);
	}
}

module.exports = SigningLog;
