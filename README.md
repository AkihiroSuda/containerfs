# ContainerFS exposes containers' FS to the host

[![Build Status](https://travis-ci.org/AkihiroSuda/containerfs.svg?branch=master)](https://travis-ci.org/AkihiroSuda/containerfs)

ContainerFS exposes Docker containers' FS to the host.

It should be useful only for ad-hoc jobs such as interactive debugging.
DO NOT USE IN PRODUCTION.

## Install

    $ go get github.com/AkihiroSuda/containerfs

## Usage

    $ sudo containerfs /mnt/containerfs
    $ docker run --name foo bar
    $ sudo ls /mnt/containerfs/foo

