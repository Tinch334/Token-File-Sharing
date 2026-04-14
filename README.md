# TFS Packet format:
A TFS packet is composed of a message type followed by an arbitrary amount of data blocks. More specifically the bytes are:

```
(1 byte ) Message type
(4 bytes) Data length, specifies <n bytes> |
(n bytes) Data                             |
(1 byte ) Continuation indicator           | > Can repeat
```
The data length and the data must be in Big-Endian format. The additional block indicator informs whether there are more data packages afterwards. If there's another packet it's value must be 0xCC, otherwise it must be 0xAA.

# Packet types
The packet position in the list denotes it's code:
1. An error, the data is the error message.
2. A user connection request, the data is two blocks. The first contains the username and the second the password.
3. A token connection request, the data is the token
4. Echo, requests an echo message from the server, the data is the echoed value.
5. File list request, request a list of all the files in the server.