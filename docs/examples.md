# Example Use Cases

This page shows practical examples of how to use `mmt`, including sample config files.



## 🧪 Importing media

### GoPro over USB Ethernet (GoPro Connect)

```bash
.\mmt.exe import --input "172.23.120.51" --output "C:\Users\konrad\Videos\Projects" --name "MadridXmas" --range "22-12-2022,30-12-2022" --camera gopro --connection connect --skip_aux false
```

### GoPro over WiFi

```bash
.\mmt.exe import --input "10.5.5.9" --output "." --camera gopro --connection connect
```

### Insta360 SD card

```bash
.\mmt.exe import --input "F:\" --output "C:\Users\konrad\Projects" --name "Skiing" --camera insta360
```

### DJI drone SD card

```bash
.\mmt.exe import --input "F:\" --output "C:\Users\konrad\Projects" --name "Skiing" --camera dji
```

Files will be imported into the same project folder, sorted by date and location.



## ⚙️ Config file examples

Rather than passing in flags every time, you can define them in a config file.

> Location: `~/.mmt.yaml`

Example use cases below — these live in the [`config/`](https://github.com/KonradIT/mmt/tree/development/config) folder of the repo.


### Wireless GoPro

[`config/wireless_gopro.yaml`](https://github.com/KonradIT/mmt/blob/development/config/wireless_gopro.yaml):

```yaml
input: 10.5.5.9
camera: gopro
output: C:\Users\konrad\Videos\Projects
range: week
skip_aux: false
connection: connect
```


### SD card GoPro

[`config/sd_card_gopro.yaml`](https://github.com/KonradIT/mmt/blob/development/config/sd_card_gopro.yaml):

```yaml
input: "F:\\"
camera: gopro
output: C:\Users\konrad\Videos\Projects
range: week
skip_aux: false
connection: sd_card
```



### HiLight tags

[`config/gopro_hilights.yaml`](https://github.com/KonradIT/mmt/blob/development/config/gopro_hilights.yaml):

```yaml
tag_names:
  - "Marked 1"
  - "Good Stuff"
  - "Important"
```



### Custom location ordering

[`config/location_tweaks.yaml`](https://github.com/KonradIT/mmt/blob/development/config/location_tweaks.yaml):

```yaml
location:
  format: 1
  fallback: ''
  order:
    - camera
    - date
    - location
```


## 🧪 Test with a config

```bash
.\mmt.exe import --config config/wireless_gopro.yaml
```



## 📁 Example output

```
C:\USERS\KONRAD\VIDEOS\PROJECTS\ELESCORIALUAV
├───21-12-2022
│   └───Air 2S
│       ├───El Escorial España
│       │   ├───photos
│       │   └───videos
│       └───San Lorenzo de El Escorial España
│           ├───photos
│           └───videos
```

