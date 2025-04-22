# KeyCLI - RSA Key Pair Generator

## Features

- Generate RSA private and public key pairs.
- Specify the key size (2048 or 4096 bits recommended).
- Save the keys to custom file paths.

## Usage

Run the utility with the following flags:

- `-private`: Path to save the private key (default: `private_key.pem`).
- `-public`: Path to save the public key (default: `public_key.pem`).
- `-size`: Key size in bits (default: `2048`).

### Example Commands

1. Generate a key pair with default settings:
	```bash
	./keycli
	```

2. Generate a key pair with a custom private key path:
	```bash
	./keycli -private my_private_key.pem
	```

3. Generate a key pair with a custom public key path:
	```bash
	./keycli -public my_public_key.pem
	```

4. Generate a key pair with a custom key size (e.g., 4096 bits):
	```bash
	./keycli -size 4096
	```

5. Generate a key pair with all custom settings:
	```bash
	./keycli -private my_private_key.pem -public my_public_key.pem -size 4096
	```

## Output

- The private key will be saved to the file specified by the `-private` flag.
- The public key will be saved to the file specified by the `-public` flag.

## Notes

- Ensure you have write permissions to the specified file paths.
- Use a key size of at least 2048 bits for secure encryption.