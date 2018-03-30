# Prattle

## Description

Distributed Embedded Key-Value store


## Installation

```
go get -u github.com/gojekfarm/prattle
```

## Usage

You can build the example using the following steps:
```
cd example
glide install
go build
```

You then need to run two instances of the app on different tabs of your
terminal:
```
./example -http_port 8000 -rpc_port 9000 -members
'0.0.0.0:9000,0.0.0.0:9001'

./example -http_port 8001 -rpc_port 9001 -members
'0.0.0.0:9000,0.0.0.0:9001'
```

Sample payload for the '/add_key' API is shown below:
```
{
"key": "ping",
"value": "pong"
}
```

## Developing

```
try,
    make deps
    make test

or just,
    make
```

## License

```
Copyright 2018, GO-JEK Farm <http://gojek.farm>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
