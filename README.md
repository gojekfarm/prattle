# Prattle
Distributed Embedded Key-Value store

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
