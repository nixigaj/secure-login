const codeElement = document.querySelector('code');

fetch('/api')
	.then(response => {
		if (!response.ok) {
			throw new Error('Network response was not ok');
		}
		return response.text();
	})
	.then(data => {
		codeElement.textContent = data;
	})
	.catch(error => {
		console.error('There was a problem with the fetch operation:', error);
		codeElement.textContent = 'Error fetching data';
	});
