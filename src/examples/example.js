var http = require('http');

http.globalAgent.maxSockets = 1000;

http.get("http://peaceful-falls-6706.herokuapp.com/sample?n=5", function(res){
	var body = "";

	res.on("data", function(chunk){ body += chunk });
	res.on("end", function() {
		var parsed = JSON.parse(body);
		onURLList(parsed);
	});
});

var onURLList = function(response) {
	var counter = 0;

	for (var i = response.urls.length - 1; i >= 0; i--) {
		counter++;
		http.get(response.urls[i].url, (function(currIndex) {
			return function(res) {
				var body = "";
				var current = response.urls[currIndex];

				res.on("data", function(c) { body += c; });
				res.on("end", function() {
					console.log("Got response for", currIndex, current, body);
					var parsed = JSON.parse(body);
					current.status_code = res.statusCode;
					current.data = parsed;

					counter--;
					if ( counter === 0 ) {
						console.log(response.urls);
						process.exit(0);
					}
				});
			};
		})(i));
	};
};