# i2c - a package for accessing the i2c bus on Linux

## Overview

The `i2c` package provides a native Go interface to the Linux
i2c/smbus device drivers.

Automated package documentation for this Go package should be
available from [![Go
Reference](https://pkg.go.dev/badge/zappem.net/pub/io/i2c.svg)](https://pkg.go.dev/zappem.net/pub/io/i2c).

## Getting started

Cross compiling to make something runable on a Raspberry Pi binary can
be done as follows:
```
$ GOARCH=arm GOOS=linux go build example/s35390.go
```

This example is for a specific i2c device: a [real time clock](https://www.ablic.com/en/doc/datasheet/real_time_clock/S35390A_E.pdf).

Another example looks at the two supported addresses for one of the [Bosch pressure sensors](https://community.bosch-sensortec.com/t5/Knowledge-base/BMP-series-pressure-sensor-design-guide/ta-p/7103):
```
$ GOARCH=arm GOOS=linux go build example/bpmx8x.go
```

## TODOs

Explore some different i2c Raspberry Pi hats, perhaps add some more
examples.

## License info

The `i2c` package is distributed with the same BSD 3-clause license as
that used by [golang](https://golang.org/LICENSE) itself.

## Reporting bugs and feature requests

The package `i2c` has been developed purely out of self-interest and a
curiosity for physical IO projects, primarily on the Raspberry
Pi. Should you find a bug or want to suggest a feature addition,
please use the [bug
tracker](https://github.com/tinkerator/i2c/issues).
