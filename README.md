# P.I. Phone Home

Daemon to run on a Raspberry Pi which periodically reports where it is on
the network. You'll need a webserver which you can see the access logs of.

You'll need to build it within a Docker container which you can access with:
```
make docker
```

Then, from within the container, run the tests and build:
```
make test arm
```

And deploy to Snappy Ubuntu Core:
```
export PHONE_URL=http://web.example.com/pi-phone-home
export SNAPPY_URL=ssh://user@rpi.example.com
make snappy
```
