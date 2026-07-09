#!/bin/sh
# Generate manpage for packaging into release archives / Homebrew cask.
set -e
rm -rf manpages
mkdir manpages
go run ./cmd/emu docs --format man-with-section | gzip -c -9 >manpages/emu.1.gz
