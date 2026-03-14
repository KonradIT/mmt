## Config file:

The location of the file is: `~/.mmt.yaml`. It is not created automatically.

By default mmt will not use any config file, but you can change some aspects of the software only via this config file, as well put the values of the different CLI flags into the file to save time.

You can check out some sample config files for different cameras and importing modes: [./config](https://github.com/KonradIT/mmt/tree/development/config)

Example, if you only want to use the software to import media from GoPros wirelessly:

```yaml
camera: gopro
output: /home/user/Videos
input: 10.5.5.9
connection: connect
```

Supported keys for `import` command:

- input
- output
- camera
- name
- date
- buffer
- range
- sort_by
- skip_aux
- connection
- tag_names

### Location:

Location.* keys will impact anything related to order of import that uses location info.

Default values:

- format: 1
- fallback: `NoLocation`
- order: "date", "location", "camera"

```yaml
location:
  format: 1 # Different formats supported: 1 and 2 (default 1)
  fallback: "NoLocation" # Leave empty to not make a folder for ungeolocated footage
  order: # Default order is:
  - date
  - location
  - camera
```

Location formats:

- 1:

```go
if len(address.City) < 9 && address.State != "" {
    return fmt.Sprintf("%s %s %s", address.City, address.State, address.Country)
}
return fmt.Sprintf("%s %s", address.City, address.Country)
```

- 2:

```go
return address.Country
```

### Camera: DJI

Prefix: `dji`

| Key | Description | Default | Example |
|-----|-------------|---------|---------|
| srt | Folder for SRT files | "" | `  srt: 'srt files'` | 

### Camera: GoPro

Prefix: `gopro`
| Key | Description | Default | Example |
|-----|-------------|---------|---------|
| gps_fix | Only use the following GPS fix values | `2, 3 // ==> 3d lock, 2d lock` | ` 2` | 
| gps_accuracy | Only GPS accuracy values equal or above to this setting will be used | 500 | 450 |
