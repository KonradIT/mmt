# MMT

**M**edia **M**anagement **T**ool

Manage your action camera/drone files intelligently.

This tool draws inspiration from my [dji-utils/offload.sh](https://github.com/KonradIT/djiutils/blob/master/offload.sh) script as well as the popular [gopro-linux tool](https://github.com/KonradIT/gopro-linux/blob/master/gopro#L262) and @deviantollam's [dohpro](https://github.com/deviantollam/dohpro)


### Camera support:

-   GoPro:
    - HERO2 - HERO5
    - MAX
    - HERO6 - HERO13
    - HERO 2024
    - MAX2
-   Insta360: X2, GO2, X3, X4
-   DJI:
    - Mavic drones (tested with Air 2, Air 2S, Mini 3 Pro, Mavic 3)
    - Osmo Action cameras (tested with Action 3)
    - Osmo Pocket cameras (tested with Pocket 1)
    - Osmo Nano camera
-   Android: All, but with Pixel 6 (Google Camera) specific fixes
- Autel Lite drone

### Features:

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

### Installing:

Download binary from [the releaser Github Action](https://github.com/KonradIT/mmt/actions/workflows/build-artifacts.yaml)

### Running:

Different commands are supported, [refer to the wiki](https://github.com/KonradIT/mmt/wiki/commands)

[How to configure mmt](https://github.com/KonradIT/mmt/wiki/configfile)

### Development:

pkg/* hosts different implementations for each camera.