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
    - HERO7 - HERO11
-   Insta360: X2, GO2, X3
-   DJI: Osmo Pocket 1/2, Mavic (most of them)
-   Android: All, but with Pixel 6 (Google Camera) specific fixes

Feel free to PR!

I plan have the tool read a directory, use a config file and act accordingly to offload media from any type of drive

## Installing:

Download from the releases tab, additionally, a github action will run for every push.

## Running:

-   import - **import camera footage**
    -   `--input`: Either one of these:
        -   A directory pointing to your SD card, on Windows it would be a letter (eg: `E:\`)
        -   USB Ethernet IP (v4) bound to a GoPro Connect connection (HERO8/9 Black)
    -   `--output`: Destination folder, a hard drive, etc...
    -   `--name`: Project name, eg: `Paragliding Weekend Winter 2021`
    -   `--camera`: Type of device being imported. Values supported: `gopro, insta360, dji, android`
    -   `--buffersize`: Buffer size for copying files. Default is `1000 bytes`
    -   `--date`: Date format. Default is `dd-mm-yyyy`
    -   `--range`: Date range, for example: `12-03-2021,15-03-2021`
    -   GoPro specific:
        -   `connection`: `sd_card`/`connect`
        -   `skip_aux`: Skips `.THM`, `.LRV` files
        -   `sort_by`: Sort by: `camera`, `days` (defaults to both)
-   update - **updates your camera**
    -   `--input`: A directory pointing to your SD card, MTP or GoPro Connect not supported
    -   `--camera`: Type of device being updated. Values supported: `gopro, insta360`
-   list: - **list devices plugged in**

## How it looks:

![](https://i.imgur.com/04m55zg.png)

## To-do:

Refer to [Issues](https://github.com/KonradIT/mmt/issues)