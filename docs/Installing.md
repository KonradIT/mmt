## How to install MMT

### Via binary:

Right now, binaries will be provided via the GitHub Actions builder: https://github.com/KonradIT/mmt/actions/workflows/build-artifacts.yaml

and also Releases tab: https://github.com/KonradIT/mmt/releases

Download the binary for your OS and move it to an executable PATH (eg: `/usr/bin/`).

### Build from source:

```bash
git clone https://github.com/konradit/mmt.git
cd mmt
go install
```

To install system-wide.

### Dependencies:

The program will call FFmpeg for several tasks including transcoding/merging/extracting metadata. Make sure it's installed and in your PATH.