console.log("Welcome to register.html")

argon2
	.hash({
		pass: 'password',
		salt: 'passhash'
	})
	.then(hash => {
		document.querySelector('code').innerText =
			`Encoded: ${hash.encoded}\n` +
			`Hex: ${hash.hashHex}\n`;

		argon2
			.verify({
				pass: 'password',
				encoded: hash.encoded
			})
			.then(() => document.querySelector('code').innerText += 'Verified OK')
			.catch(e => console.error('Error: ', e));
	})
	.catch(e => console.error('Error: ', e));
