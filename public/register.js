console.log("Welcome to register.html")

const pass = 'password';
const salt = 'passhash';

const hashPromise = argon2.hash({ pass: pass, salt: salt });
console.log(hashPromise)
hashPromise.then(handleHashSuccess).catch(handleError);

function handleHashSuccess(hash) {
	const encoded = hash.encoded;
	const hashHex = hash.hashHex;
	const codeElement = document.querySelector('code');
	codeElement.innerText = 'Encoded: ' + encoded + '\n' +
		'Hex: ' + hashHex + '\n';

	const verifyPromise = argon2.verify({ pass: pass, encoded: encoded });
	verifyPromise.then(handleVerifySuccess).catch(handleError);
}

function handleVerifySuccess() {
	const codeElement = document.querySelector('code');
	codeElement.innerText += 'Verified OK';
}

function handleError(e) {
	console.error('Error: ', e);
}
