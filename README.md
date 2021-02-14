# gruler

## Communication Protocol

For communication between your application server and the gruler server, protocol buffers are used.

See request.proto and response.proto files to see their definitions.

Protobufs have no way to delimitate between different payloads to tell where one ends and the next begins, 
so the following protocol is used when making a request to gruler and reading responses.

### Requests
- The first 4 bytes must be an integer `n` specifying the size of the request not including the size of `n`
- The next `n` bytes will be the serialized payload

The same connection can and should be reused to send multiple requests. 

See test_clients.go for an example

### Responses
- The first 4 bytes are an integer `n` specifying the size of the response not including the size of `n`
- The next `n` bytes will be the serialized payload

See test_clients.go for an example
