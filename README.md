# switchbot-meter-cli

A simple CLI tool for SwitchBot Meter

## Install

Download the latest release.

## Usage

```
Usage of switchbot-meter:
  -a string
        MAC address of the meter device.
  -d    Run as daemon.
  -t uint
        Specify a timeout in seconds before exits. This option has no effect when used with '-d' (default 10)
```

### Example

Get a single metric from the meter.

```
$ sudo switchbot-meter -a aa:bb:cc:dd:ee:ff
{"addr":"aa:bb:cc:dd:ee:ff","bat":100,"temp":23.8,"humi":50,"ts":1639840189947}
```

## For Developers

### Build

```
$ make
```
