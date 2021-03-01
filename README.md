# gruler

Gruler is a Rules Engine for defining actions on HTTP requests. It runs as a process listening in on a unix domain socket. 
The webserver writes a summary of the incoming request to the socket and receives a response containing a list of actions
that apply to that request. Rules are defined in json see example-rules.json.

## Rule
Each rule consists of:

- Rule Id: This uniquely identifies the rule in your rule set.
- Condition: A condition which if evaluates to true results in the corresponding action applying.
- Action: The action to apply to the request if the condition evaluates to true

TODO: Explain the rule authoring syntax.

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
