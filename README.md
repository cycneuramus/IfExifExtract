## Overview

This tool traverses a directory tree, finds all JPEG files containing a given EXIF metadata value, and copies them to another directory.

---

```
Usage of IfExifExtract:
  -dstDir string
    	Directory to receive matching files
  -exifKey string
    	EXIF key to query
  -exifQuery string
    	EXIF value to match
  -srcDir string
    	Directory to scan
```
### Example

To extract all JPEG files where the EXIF key `Subject` has a value of `John Smith` and/or `Jane Smith`:

``` 
IfExifExtract \
	-srcDir=/path/to/source \
	-dstDir=/path/to/destination \
	-exifKey=Subject \
	-exifQuery="John Smith, Jane Smith"
```
