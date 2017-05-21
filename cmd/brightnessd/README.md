brightnessd
===========

Commandline brightness deamon which sets brightness for laptops based on the sun times of your location.


Installation
------------

```
go get github.com/rikvdh/go-tools/cmd/brightnessd
```

Usage
-----

```
brightnessd -lat <latitude> -long <longitude> -min 4 -max 80
```

```
Usage of ./brightnessd:
  -lat string
        Latitude for the user location (default "51.697816")
  -long string
        Longitude for the user location (default "5.303675")
  -max float
        maximum brightness (default 80)
  -min float
        minimal brightness (default 4)
```