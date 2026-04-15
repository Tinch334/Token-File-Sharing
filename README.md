# TFS Packet format:
A TFS packet can have two forms:

## Server to client
In this form there's no auth data, since the client has no need to authenticate the server directly. Only message type and data is transmitted.
```
(1 byte ) Message type
(4 bytes) Data length, specifies <n bytes> |
(n bytes) Data                             |
(1 byte ) Continuation indicator           | > Can repeat
```

## Client to server
In this case there is a need for authentication, as such there is an auth block; However if the bytes corresponding to auth token length are zero then no auth token will be expected. Note that no auth information is only allowed with packets of type 1. There can be an arbitrary amount of data blocks. More specifically the bytes are:
```
(1 byte ) Message type
(2 bytes) Auth token length <n bytes> |
(n bytes) Auth token                  | > Auth block
(4 bytes) Data length, specifies <n bytes> |
(n bytes) Data                             |
(1 byte ) Continuation indicator           | > Can repeat
```
All the data must be in Big-Endian format. The continuation indicator informs whether there are more data packages afterwards. If there's another packet it's value must be 0xCC, otherwise it must be 0xAA.

# Packet types
The packet position in the list denotes it's code:
0. An error, the data is the error message.
1. A user connection request, the data is two blocks. The first contains the username and the second the password.
2. A token connection request, the data is the token
3. Close, indicates that the given token will no longer be used
4. Ok, indicates that the given operation succeeded, the data is optional, if it's present it contains a descriptive message of what succeeded.
5. Token response, used by server to inform host of it's token.
6. Echo send, requests an echo message, the data is the echoed value.
7. Echo receive
8. File list request, request a list of all the files in the server.
9. File download request
10. File upload request