## Message Format

Request Format:

|-method name length (1 byte) |  +  |-- request length (4 byte) --|  +  |-.. method name (up to 127 bytes)..-|  +   |--.. request (up to 2^31-1 bytes) ..--|

---------------------
Response Format:

|--response length(4 byte) --|  +  |-- error length(4 byte) --|   +   |-.. response (up to 2^31 -1 bytes) ..-|  +  |-.. error (up to 2^31 -1 bytes) ..-|


## Client Design

## Server Design



## TODO List

- [ ] Write a design document

- [x] Finish writing encoder, decoder and their unit test

- [ ] Going to consider supporting multiplexing in the future - that is multiple gRPC calls can share the same TCP connection. 
Each gRPC call will have a unique request ID in the case

- [x] On client side, all TCP connections are stored in a channel

- [] Change print to log
     - Requirements:
        - Print log to stderr and also save to a file.
        - Need to include the timestamp, file:line, and (hopefully) the function

- [x] Add pprof

- [ ] A config file for server and client
    
- [x] Finish Server Part

- [] Server unit test

- [x] replace binary.Write with a more efficient one. (avoid using reflection)

- [] Remove logging logic in the customrpc package

- [x] Benchmark codec

- [] Try telnet server port

- [] Resume broken TCP connections