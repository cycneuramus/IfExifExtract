## Overview

This tool traverses a directory tree, finds all JPEG files containing a given EXIF metadata value, and copies them to another directory.

## Example

To extract all JPEG files where the EXIF key `Subject` has the value of `John Smith`:

``` 
export SRC_DIR="/path/to/imgs"
export DST_DIR="/path/to/extracted"
export EXIF_KEY="Subject"
export EXIF_VAL="John Smith"

go run main.go
```
