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
    - HERO6 - HERO11
-   Insta360: X2, GO2, X3
-   DJI: Osmo Pocket 1/2, DJI Osmo Action 1/2/3, Mavics, Minis
-   Android: All, but with Pixel 6 (Google Camera) specific fixes

Feel free to PR!

I plan have the tool read a directory, use a config file and act accordingly to offload media from any type of drive

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

## Installing:

Download from the releases tab, additionally, a github action will run for every push.

## Running:

-   import - **import camera footage**
    -   `--input`: Either one of these:
        -   A directory pointing to your SD card, on Windows it would be a letter (eg: `E:\`)
        -   USB Ethernet IP (v4) bound to a GoPro Connect connection (GoPro HERO8, HERO9) / OpenGoPro (>HERO9)
        -   `10.5.5.9` if connected to a GoPro wirelessly
    -   `--output`: Destination folder, a hard drive, etc...
    -   `--name`: Project name, eg: `Paragliding Weekend Winter 2021`
    -   `--camera`: Type of device being imported. Values supported: `gopro, insta360, dji, android`
    -   `--buffersize`: Buffer size for copying files. Default is `1000 bytes`
    -   `--date`: Date format. Default is `dd-mm-yyyy`
    -   `--range`: Date range, for example: `12-03-2021,15-03-2021`
    -   GoPro specific:
        -   `connection`: `sd_card`/`connect`
        -   `skip_aux`: Skips `.THM`, `.LRV` files
        -   `sort_by`: Sort by: `camera` (default: `camera` true)
-   update - **updates your camera**
    -   `--input`: A directory pointing to your SD card, MTP or GoPro Connect not supported
    -   `--camera`: Type of device being updated. Values supported: `gopro, insta360`
-   merge - **merges videos together**
    -   `--input`: Files to merge. Specify multiple times
-   list: - **list devices plugged in**

## Configuration file:

By default mmt will not use any config file, but you can change some aspects of the software only via this config file, as well put the values of the different CLI flags into the file to save time.

The default location is: `~/.mmt.yaml`.

```yaml
input:
camera:
model:
...
location:
  format: 1 # Different formats supported: 1 and 2 (default 1)
  fallback: "NoLocation" # Leave empty to not make a folder for ungeolocated footage
  order: # Default order is:
  - date
  - location
  - device
```

## How it looks:

![](https://i.imgur.com/04m55zg.png)

## To-do:

Refer to [Issues](https://github.com/KonradIT/mmt/issues)