Message Protocol Specification
==============================


Message
------
All Messages are sent through zeromq.  This ensures integrity and size.
```
Magic Number | uint32 | Verify valid message server (63004c7)
Type         | string | Payload description
Payload      | []byte | Payload / Encrypted content
```


Peer
----
```
IP Address | net.IP | IP Address
Port       | uint16 | Port number to connect on
Last Seen  | int64  | Standard Unix timestamp
```

Version
-------
```
Version   | uint32 | Current version running on a given peer
Timestamp | int64  | Standard Unix timestamp
UserAgent | string | A null terminated UA string
```





