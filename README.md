EMP
=====
EMP is a fully encrypted, distributed messaging service designed with speed in mind.
Originally based off of BitMessage, EMP makes modifications to the API to include
both Read Receipts that Purge the network of read messages, and an extra identification field
to prevent clients from having to decrypt every single incoming message.

Submodules
----------

This repository contains submodules.  You will need to run:
```
git submodule init
git submodule update
```

Launching
---------

Running `./start.sh` will start empd and open a local browser.

Running `./stop.sh` wil stop empd
