![Build](https://github.com/shawc71/gruler/actions/workflows/go.yml/badge.svg)

# gruler

Gruler is a [Rules Engine](https://martinfowler.com/bliki/RulesEngine.html) for defining actions on HTTP requests. It runs as a process listening in on a unix domain socket, the webserver (which can be any language that speaks [Protocol Buffers](https://developers.google.com/protocol-buffers)) writes a summary of the incoming request to the socket and receives a response containing a list of actions
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

For a list of all available `request_field`s. See [Introspectable Request fields](#introspectable-request-fields)

#### in
This condition is written as `"{in": {"request_field": ["val-a", "val-b", "val-c", ...]"}}`

The `request_field` is the property of the request to examine, if the runtime value of the `request_field`
is one of the ones to the specified in the array on the right hand side then the condition evaluates to `true`.

For example `"{in": {"request.method": ["GET", "OPTIONS", "PATCH"]}}` -- this condition will evaluate to true if the 
request is a GET, OPTIONS or PATCH request.

For a list of all available `request_field`s. See [Introspectable Request fields](#introspectable-request-fields)

The supported non-terminal conditions are:

#### and
This condition is written as `{"and": [condition-a, condition-b, condition-c ...]"}`

This condition is used when you want the condition to be true if a set of conditions is true. This is similar to `&&` in
programming languages.

For example:
```
{
    "or": [
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

#### or
This condition is written as `{"or": [condition-a, condition-b, condition-c ...]"}`

This condition is used when you want the condition to be true if one ofa set of conditions is true.
This is similar to `||` in programming languages.

For example:
```
{
    "or": [
      {"eq": {"request.method": "GET"}},
      {"eq": {"request.header.host": "example.com"}},
    ]
}
```
this condition will evaluate to true if one of the sub-conditions is true ie it's a GET request or the Host header set
to example.com

Sub-conditions are recursive and may be nested inside of one another without any limit, so the subcondition can be any
condition terminal or non-terminal.

#### not
This condition is written as `"not": {condition}`.

This negates the result of the nested condition. This is similar to `!` in programming languages.

For example `"not": {"eq": {"request.method": "PUT"}}` will evaluate to true if the request method is not PUT. The
nested condition can be aby condition.

### Actions

Each rule can contain one action. Actions specify what to do when the condition evaluates to true. The following actions are supported

#### setHeader
This action will attach a header to the request and pass it on to your application, the purpose of this is if you want your application code to have different behavior based on the corrosponding rule condition evaluating to true.

Example:

```
    "action": {"type": "setHeader", "headerName": "Internal-Foo", "headerValue": "Value"}
```

Will attach a header with the name "`Internal-Foo`" and value "`Value`" to the request

#### block
This action will block the request and it will not be forwarded to your application

```
    {
      "rule_id": "test-in-condition",
      "condition": {"in": {"request.queryParam.userId": ["joe", "jane", "victor"]}},
      "action": {"type": "block", "statusCode": 503}
    }
```

Will return status code 503 and block the request without forwarding it to our application if userId query param has one of the given values


#### throttle

Rate limits requests, once the rate limit is exceeded all requests are rejected with status code 429

```
    {
      "rule_id": "test-static-throttle",
      "condition": {
        "and": [
          {"eq": {"request.header.host": "foo.example.com"}},
          {"eq": {"request.clientIp": "7.7.7.7"}}
        ]
      },
      "action": {"type": "throttle", "max_tokens": 50, "refill_amount": 5, "refill_time": 1 }
    }
```

will limit requests from ip `7.7.7.7` to 5 requests per second bursting upto 50.

Throttle configuration requires the following:

`refill_amount`: Think of this as the number of requests you want to allow in a given unit of time

`refill_time`: The time unit that applies to refill amount.

Together the values of these two fields determine the rate of your throttle, for example `refill_amount` 5 and `refill_time` 1 essentially means a rate of 5 requests per second.

`max_tokens`: Determines the maximum "burst". For example of you have throttle that sets the rate to 5 requests per second but affected party only uses say 2 requests in a given second you can use the `max_tokens` field to allow the unused capaacity to carry over upto the number specified as a value of this field. So if it's value was 50, any unused capacity would carry over upto a maximum of 50 and user could potentially make 50 requests per second by saving up enough capacity from previous time units. If you don't want to allow any bursts, just set the value of `max_tokens` to be the same as `refill_amount`.

`each_unique`: Optional, this creates a dynamic throttle. For example:

```
    {
      "rule_id": "test-eachUnique-throttle",
      "condition": {
        "and": [
          {"eq": {"request.header.host": "www.example.com"}}
        ]
      },
      "action": {
        "type": "throttle",
        "max_tokens": 5,
        "refill_amount": 5,
        "refill_time": 1,
        "each_unique": "request.clientIp"
      }
    }
```

this allows 5 requests per second for each request with the `host` header set to `www.example.com` for each unique IP Address.

## Introspectable request fields

The following fields are available to be used in rules:

| Field  | Description  |
|---|---|
| request.clientIp  |  Returns the ip address of the client eg `10.1.1.1` |
| request.method  |  The request method eg `POST` |
| request.httpVersion  |  The http version used by the request eg `HTTP/1.1`|
| request.header.$header_name  |  Will return the value of the header $header_name. Eg `request.headers.host` will return the value of the `Host` header |
| request.queryParam.$param | Will return the value of the query param $param. Eg `request.queryParam.user` will return the value of the `user` query param  |
| request.rawUri  |  The raw uri of the request|
| request.rawQueryParams  |  The raw query param string associated with the request eg `user=foo&productId=1&year=2021|



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
