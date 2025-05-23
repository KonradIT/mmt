## Commands:

### import: import camera footage

-   `--input`: Either one of these:
	-   A directory pointing to your SD card, on Windows it would be a letter (eg: `E:\`)
	-   USB Ethernet IP (v4) bound to a GoPro Connect connection (GoPro HERO8, HERO9) / OpenGoPro (>HERO9)
	-   `10.5.5.9` if connected to a GoPro wirelessly
-   `--output`: *MANDATORY* Destination folder, a hard drive, etc...
-   `--name`: Project name, eg: `Paragliding Weekend Winter 2021`
-   `--camera`: *MANDATORY* Type of device being imported. Values supported: `gopro, insta360, dji, android`
-   `--buffersize`: Buffer size for copying files. Default is `1000 bytes`
-   `--date`: Date format. Default is `dd-mm-yyyy`
-   `--range`: Date range, for example: `12-03-2021,15-03-2021`
-   `--skip_aux`: Skips `.THM`, `.LRV` files on GoPro cameras and .SRT files on DJI systems
-   `--sort_by`: Sort by: `camera, location` (default: `camera, location` true)
-   `--camera_name`: Camera name

- Helper shortcuts:
    -  `--use_gopro`: Use the first GoPro available. Order: USB Ethernet > SD card. Errors out if no GoPro is found
    -  `--use_insta360`: Use the first Insta360 available. Errors out if no Insta360 is found

-   GoPro specific:
	-   `--connection`: `sd_card`/`connect`
	-   `--tag_names`: Tag names for number of HiLight tags in last 10s of video, each position being the amount, eg: 'marked 1,good stuff,important' => num of tags: 1,2,3"

Examples:

To import media from a GoPro HERO9/10/11 over USB Ethernet (GoPro Connect), while retaining LRV Proxy files and only the days 22-30 of December

```bash
.\mmt.exe import --input "172.23.120.51" --output "C:\Users\konrad\Videos\Projects" --name "MadridXmas" --range "22-12-2022,30-12-2022" --camera gopro --connection connect --skip_aux "false"
```

To import media from a GoPro camera over WiFi to the current directory

```bash
.\mmt.exe import --input "10.5.5.9" --output "." --camera gopro --connection connect
```

To import media from an Insta360 camera's SD card

```bash
.\mmt.exe import --input "F:\" --output "C:\Users\konrad\Projects" --name "Skiing" --camera insta360
```

To import media from a DJI drone's SD card

```bash
.\mmt.exe import --input "F:\" --output "C:\Users\konrad\Projects" --name "Skiing" --camera dji
```

The cool part is, both Insta360 and DJI footage will be on the same project directory, and the files will be organized by day/location

```
C:\USERS\KONRA\VIDEOS\PROJECTS\ELESCORIALUAV
├───21-12-2022
│   └───Air 2S
│       ├───El Escorial España
│       │   ├───photos
│       │   └───videos
│       └───San Lorenzo de El Escorial España
│           ├───photos
│           └───videos
```

### update: updates your camera

-   `--input`: A directory pointing to your SD card, MTP or GoPro Connect not supported
-   `--camera`: Type of device being updated. Values supported: `gopro, insta360`

### apply-lut: Applies LUT profile to photos

- `--input`: path to one JPG file or directory of JPG files
- `--lut`: Path to .CUBE LUT
- `--intensity`: 0-100
- `--resize`: Resize image. Use `0` to keep aspect ratio
- `--quality`: 0-100

```bash
./mmt.exe apply-lut --input "DJI_0468.JPG" --lut '..\GoPro LUTs Color Grading Pack by IWLTBAP\LUTs by IWLTBAP (CUBE)\IWLTBAP Kumate.cube' --intensity 75 --resize 2000x0 --quality 90
```

### merge: merges videos together

-   `--input`: Files to merge. Specify multiple times

```bash
 .\mmt.exe merge --input "G:\Edits\Extremadura trip\13-03-2021\HERO9 Black\videos\GH1276-01.MP4" --input "G:\Edits\Extremadura trip\13-03-2021\HERO9 Black\videos\GH1276-02.MP4"
```

### list: list devices plugged in

Will show you devices plugged in to make importing easier.

```
📷 Devices:
        🎥 0 - C: (C:)
        🎥 1 - F: (F:)
        🎥 2 - G: (G:)
🔌 GoPro cameras via Connect (USB Ethernet):
        📹 0 - 172.23.120.51 (HERO11 Black - H22.01.02.01.00)
```

### calendar: display an ascii calendar of days when media was captured