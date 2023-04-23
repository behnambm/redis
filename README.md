### A simple implementation of Redis server in Go

It listens on TCP port `6379` and decodes RESP format and responds according to get command that the client sends.

#### List of commands and features of server:

- Ping 
- Echo 
- Set 
- Set with expiry(milliseconds) 
- Get

---

This server is compatible with `redis-cli` so it makes it compatible with other redis clients too.

--- 

### Run:
`make run`

To test you can use:

`redis-cli PING`


# TODO

- Read port number from command line arguments
