# MMT

**M**edia **M**anagement **T**ool

Or, what to do if your desk looks like this:

![](https://i.imgur.com/qmgLaxg.jpg)

## Backstory:

I've been using an assortment of scripts over the years to manage media from my different action cameras and drones, it's clear a centralized and unified solution is needed.

This tool draws inspiration from my [dji-utils/offload.sh](https://github.com/KonradIT/djiutils/blob/master/offload.sh) script as well as the popular [gopro-linux tool](https://github.com/KonradIT/gopro-linux/blob/master/gopro#L262) and @deviantollam's [dohpro](https://github.com/deviantollam/dohpro)

Right now the script supports these cameras:

-   GoPro: Pretty much all of them
-   Insta360: X2 (GO 2 to follow)
-   DJI: Tested with Osmo Pocket, Spark and Mavic Air 2, but should work on Osmo action and other drones as well
-   Android: photos and videos recorded with OnePlus 7T, but possibly most Android phones

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

-   [ ] Auto detect camera using clues from SD card.
-   [ ] **HiLight parsing**: I've found that the best way to see which clip I will use later on is to put some tags at the end of it (press mode button on GoPro, or shout "Oh Shit", or use the app/pebble app). Then when I run a script that prints the number of hilight tags during the last 30 seconds of each video. That lets me know the clips are important. This tool should let you label each tag count (eg: --tag-labels="good,great,important") for each hilight count.
-   [ ] **Sort by location**: Should be on root, so:

    ```
    - Mexico City, Mexico:
       - 2020-01-02:
    	     ...
    - New York, NY, United States:
       - 2017-07-01:
    	     ...
    - Madrid, Spain:
       - 2020-09-02:
    	     ...

    ```

    To get location info: GoPro ([GPMF](https://github.com/stilldavid/gopro-utils)) DJI (SRT file) Insta360 (???)

-   [x] **Date range**: Import from only certain dates (allow for: `today`, `yesterday` and `week`, `--date-start` and `--date-end`)
-   [X] **Sort by resolution/framerate**: use ffmpeg for getting resolution/framerate
-   [ ] **Extract info from each clip**: Eg: km travelled, altitude changes, number of faces, shouts, etc...
-   [ ] **Merging chapters**: GoPro only, merge chapters from separate files losslessly using ffmpeg
-   [ ] **Generate GIF for burst photos**: Move each burst sequence to a separate folder and make a GIF
-   [ ] **Merge timelapse photos**: Using ffmpeg
-   [ ] **Generate DNG from GPR**: Using [gpr tool](https://github.com/gopro/gpr)
-   [x] **Proxy file support**
-   [ ] **H265 to H264 conversion**: Using ffmpeg
-   [x] **Update camera firmware?** (Done: GoPro, Insta360)
-   [ ] **Use goroutines**
-   [ ] **Tests**
-   [X] **Import media from GoPro's Webcam mode (USB Ethernet)**
-   [ ] **GUI counterpart using fyne.io**
