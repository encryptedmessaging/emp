EMP
=====
EMP is a fully encrypted, distributed messaging service designed with speed in mind.
Originally based off of BitMessage, EMP makes modifications to the API to include
both Read Receipts that Purge the network of read messages, and an extra identification field
to prevent clients from having to decrypt every single incoming message.

**You can check out some more detailed information in the GitHub Wiki!**

Submodules
----------

This repository contains submodules.  You will need to run:
```
git submodule init
git submodule update
```

Required Tools
---------
In order to compile and run this software, you will need:

* The [Go Compiler (gc)](http://golang.org/doc/install)
* For Downloading Dependencies: [Git](http://git-scm.com/book/en/Getting-Started-Installing-Git)
* For Downloading Dependencies: [Mercurial](http://mercurial.selenic.com/wiki/Download)

Building and Launching
---------

* `make build` will install the daemon to ./bin/emp
* `make start` will set up the config directory at ~/.config/emp/, then build and run the daemon, outputting to the log file at ~/.config/emp/log/log_<date>
* `make stop` will stop any existing emp daemon
* `make clean` will remove all build packages and log files
* `make clobber` will also remove all the dependency sources

**Running as root user is NOT recommended!**

Configuration
---------
All configuration is found in `~/.config/emp/msg.conf`, which is installed automatically with `make start`. An example is found in `./script/msg.conf.example`. The example should be good for most users, but if you plan on running a "backbone" node, make sure to add your external IP to msg.conf in order to have it circulated around the network.

Debian/Ubuntu Installation
---------
* Add the APT repository with `add-apt-repository 'deb http://emp.jar.st/repos/apt/debian unstable main'`
* Download and install the JARST GPG Key with `wget -O key http://emp.jar.st/repos/apt/debian/conf/jarst.gpg.key && sudo apt-key add key; rm -f key`
* Update the APT Database: `sudo apt-get update`
* You can now install EMP with `sudo apt-get install emp`

**Note:**
Configuration of the Debian installation will be stored in `/usr/share/emp/` instead of the home directory.

Support
---------
Support is available through our [Google group](https://groups.google.com/forum/#!forum/encryptedmessaging).
