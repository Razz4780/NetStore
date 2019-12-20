# NetStore
Server-client application to share files over TCP

## Server
Application accepts two parameters:
1. `dir` - the location to search for files, default value `.`.
2. `port` - the port number, default value `5551`.

## Client
Application accepts one parameter - `server`. This is a valid IP address to be passed to `net.Dial` function. 
Application creates writes new files to directory `tmp` inside working directory.

## Protocol

### Requests

1. For the list of filenames - value 1 of type uint16.
2. For a file chunk - value 2 of type uint16, chunk offset of type uint32, chunk size of type uint32, 
filename length of type uint16, filename.

### Responses

1. With filenames - value 1 of type uint16, filenames field length of type uint32, filenames separated with null bytes
(with null byte after the last filename).
2. With refusal - value 2 of type uint16, refusal cause of type uint32. Refusal causes: 1 for bad filename,
2 for bad offset (greater than file size), 3 for bad chunk size (0).
3. With file chunk - value 3 of type uint16, chunk length of type uint32, chunk contents.
