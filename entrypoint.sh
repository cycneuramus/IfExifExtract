#!/bin/sh -e

case $1 in
	extract)
		exec ./IfExifExtract \
			-srcDir="$SRC_DIR" \
			-dstDir="$DST_DIR" \
			-exifKey="$EXIF_KEY" \
			-exifQuery="$EXIF_QUERY"
		;;
esac
