`con-send` is a module that facilitates the creation, signing, and sending of transactions. It includes the following components:

### Features
- **Transaction Signing**: Uses ECDSA to sign transactions securely.
- **Malicious Behavior Simulation**: Supports testing with various malicious behaviors, such as invalid keys, corrupted signatures, and altered data.
- **HTTP Integration**: Sends signed transactions to a specified server endpoint.

### Files
- **`signer.go`**: Implements the core functionality for signing transactions and handling malicious behavior.
- **`signer_test.go`**: Contains unit tests to validate the functionality of the `signer` package, including tests for valid and malicious transactions.

### Usage
1. **Transaction Signing**:
    - Create a `Transaction` struct with the sender, receiver, amount, and private key.
    - Use the `Sign` function to generate a signed transaction.

2. **Testing**:
    - Run the tests in `signer_test.go` to ensure the module behaves as expected under various scenarios.

### Example
```go
privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
tx := Transaction{
     From:   "Alice",
     To:     "Bob",
     Amount: 100,
     Key:    privKey,
}
signedTx := Sign(tx, None)
fmt.Println(string(signedTx))
```

### HTTP Endpoint
The module includes an HTTP handler (`transactionHandler`) to process incoming transaction requests. It expects a POST request with a JSON payload containing transaction details.

### Testing Framework
The module uses `testify` for assertions in unit tests. Run the tests using:
```bash
go test ./tests
```