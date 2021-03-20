# gruler

Gruler is a Rules Engine for defining actions on HTTP requests. It runs as a process listening in on a unix domain socket. 
The webserver writes a summary of the incoming request to the socket and receives a response containing a list of actions
that apply to that request. Rules are defined in json see example-rules.json.

## Rule
Each rule consists of:

- Rule Id: This uniquely identifies the rule in your rule set.
- Condition: A condition which if evaluates to true results in the corresponding action applying.
- Action: The action to apply to the request if the condition evaluates to true

### Rule Id

Rule id is the id of the rule and must be unique in the ruleset.

### Condition
There are two types of  conditions- terminal conditions and non-terminal conditions. 

Terminal conditions do not contain any nested conditions. Non-terminal conditions are recursive and contain other conditions

The supported terminal conditions are:

#### eq

This condition is written as `"{eq": {"request_field": "request_field_value"}}`

The `request_field` is the property of the request to examine, if the runtime value of the `request_field` 
equals to the specified `request_field_value` then the condition evaluates to `true`. 

For example `{"eq": {"request.method": "GET"}}` -- this condition will evaluate to true if the request is a GET request.

For a list of all available `request_field`s. See [Introspectable Request fields](#introspectable_request_fields)

#### in
This condition is written as `"{in": {"request_field": ["val-a", "val-b", "val-c", ...]"}}`

The `request_field` is the property of the request to examine, if the runtime value of the `request_field`
is one of the ones to the specified in the array on the right hand side then the condition evaluates to `true`.

For example `"{in": {"request.method": ["GET", "OPTIONS", "PATCH"]}}` -- this condition will evaluate to true if the 
request is a GET, OPTIONS or PATCH request.

For a list of all available `request_field`s. See [Introspectable Request fields](#introspectable-request-fields)

#### and
This condition is written as `{"and": [condition-a, condition-b, condition-c ...]"}`

This condition is used when you want the condition to be true if a set of conditions is true. This is similar to `&&` in
programming languages.

For example:
```
{
    "and": [
      {"eq": {"request.method": "GET"}},
      {"eq": {"request.header.host": "example.com"}},
      {"in": {"request.clientIp": ["72.1.1.1", "73.1.1.1", "74.1.1.1"]}}
    ]
}
```
this condition will evaluate to true if all the sub-conditions are true ie. it's a GET request, with Host header set 
to example.com and the client ip is one of the ones specified. 

Sub-conditions are recursive and may be nested inside of one another without any limit, so the subcondition can be any 
condition terminal or non-terminal.

## Introspectable request fields

TODO


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
