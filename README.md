# HoneyPoke

![HoneyPoke Logo](honeypoke.png)

## What is HoneyPoke?

HoneyPoke is a ~~Python~~ Go application that shows you what attackers are poking around with. It sets up listeners on certain ports and records whatever is sent to it. 

This information can be logged to different places, the currently supported outputs are:
* ElasticSearch 6
* ElasticSearch 7
* ElasticSearch 8

## What about HoneyPoke Python?

I'm not planning to support it anymore, I had too much trouble with memory leaks. Hopefully, in the long run, the new Go version will handle things better.

You can see the old code here: https://github.com/bocajspear1/honeypoke

## Quick Start (Only look at the parts about ElasticSearch and Kibana)

__A full tutorial on installing HoneyPoke and using it with ElasticSeach and Kibana is available in the [wiki](https://github.com/bocajspear1/honeypoke/wiki/Full-Install-(With-ElasticSearch-Kibana)).__

## Pre-Reqs

* Install the libpcap dev libraries (`libpcap-dev`).
* Change the port your SSH server is listening on so you can place a HoneyPoke listener there instead.
* Have Go >=1.13
* Download the GeoLite IP database from MaxMinds. Requires free account.

## Installation

1. Clone or download this repo
2. Build using `make`

## Setup and Usage

1. Copy `config.json.default`  to `config.json` Modify the config file. 
    * `recroders` enables and disables recorders. This done with the `enabled` key under the respective loggers. Some may need extra configuation, which is in the `config` key.
    * The `udp_ports` key sets the UDP listeners that you will be creating. Its just a list of ports.
    * The `tcp_ports` key sets the TCP listeners that you will be creating. It has the `port` key for the port and `ssl` as a boolean to indicate if the listener should use SSL (See **SSL Connections** below for more details) `config.json.sample` contains a sample list of ports. 
    * `ignore_tcp_ports` is used ignore TCP ports. This is useful for things like ElasticSearch and SSH so that these connections are not recorded as missing ports.
    * `interface` is used by the missed port watcher to indicate what interface to listen on.
    * `user` is the user you want the script to drop privileges to.
    * `group` is the group you want the script to drop privileges to.
2. Run HoneyPoke with `./honeypoke`


**Note:** Be sure you have nothing listening on the selected ports, or else HoneyPoke will not fully start.

**Note:** HoneyPoke is run using sudo (aka root). It will drop privileges though, and it will not process any connections until permissions are dropped. The script should report when privileges are dropped.


## SSL Connections

By setting the `ssl` key to `true`, the port will expect SSL connections. This means the socket will ignore non-SSL connections. Invalid SSL connections will produce a blank input, so only enable SSL on ports that are expected to SSL, such as 443.

SSL expects two files, the key in `honeypoke_key.pem` and the cert in `honeypoke_cert.pem`. Both should be created by the `prepare.sh` script.

You can manually create a self-signed cert with the following command:
```
openssl req -new -x509 -days 365 -nodes -out honeypoke_cert.pem -keyout honeypoke_key.pem
```

## Binary and Large files

Binary data is converted into the Python/Golang bytes format (`'\x00'`). This ensures the data is stored safely, but also keeps strings in the binary readable. For small binary data (<4096 bytes), HoneyPoke will send the data as is to the output. If the data is larger than 4096 bytes, HoneyPoke will store the output into a file in the `large` directory and the location to the file is logged instead of the entire contents.

See [here](https://stackoverflow.com/questions/43337544/read-bytes-string-from-file-in-python3) if you want to load the Python bytes format for manipulation or conversion.

## Missed ports

Ports that have no listener are recorded by HoneyPoke in the `missed.txt` file in a special format:
```
<PORT> TCP|   UDP
```
Use this is to modify your listeners with new ports.

## Contributing

Go at it! Open an issue, make a pull request, fork it, etc.

## License

This project is licensed under the Mozilla Public License Version 2.0 license