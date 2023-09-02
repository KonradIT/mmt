# MMT

**M**edia **M**anagement **T**ool

Or, what to do if your desk looks like this:

![](https://i.imgur.com/qmgLaxg.jpg)

## Backstory:

I've been using an assortment of scripts over the years to manage media from my different action cameras and drones, it's clear a centralized and unified solution is needed.

This tool draws inspiration from my [dji-utils/offload.sh](https://github.com/KonradIT/djiutils/blob/master/offload.sh) script as well as the popular [gopro-linux tool](https://github.com/KonradIT/gopro-linux/blob/master/gopro#L262) and @deviantollam's [dohpro](https://github.com/deviantollam/dohpro)

Right now the script supports these cameras:

-   GoPro:
    - HERO2 - HERO5
    - MAX
    - Fusion
    - HERO6 - HERO12
-   Insta360: X2, GO2, X3
-   DJI: Osmo Pocket 1/2, DJI Osmo Action 1/2/3, Mavics, Minis
-   Android: All, but with Pixel 6 (Google Camera) specific fixes

Feel free to PR!

I plan have the tool read a directory, use a config file and act accordingly to offload media from any type of drive

## To-do:

Refer to [Issues](https://github.com/KonradIT/mmt/issues)

## Features:

- Import videos and photos from the most popular action cameras (GoPro, Insta360, DJI)
- Fix nonsensical filenames and file structures:
  - `GH011273.MP4` and `GH021273.MP4` will become `GH1273-01.MP4` and `GH1273-02.MP4` respectively
  - `VID_20221012_102725_10_586.insv` and `VID_20221012_102725_00_586.insv` will become `102725/VID_20221012_102725_10_586.insv` and `102725/VID_20221012_102725_00_586.insv` therefore making organizing Insta360 footage easier
- Group *multi shots*/related files together, such as GoPro bursts, timelapses and Insta360 timelapse photos
- Update camera firmware
- Merge GoPro chaptered videos together
- Sort files into folders depending on:
  - Camera Name (eg: `HERO9 Black`, `Mavic Air 2`)
  - Location (eg: `El Escorial, España`)
- Apply LUT profiles to photos

## Installing:

Download binary from [the releaser Github Action](https://github.com/KonradIT/mmt/actions/workflows/build-artifacts.yaml)

## Running:

Different commands are supported, [refer to the wiki](https://github.com/KonradIT/mmt/wiki/commands)

[How to configure mmt](https://github.com/KonradIT/mmt/wiki/configfile)

## How it looks:

![](https://i.imgur.com/MjYKhfj.png)
