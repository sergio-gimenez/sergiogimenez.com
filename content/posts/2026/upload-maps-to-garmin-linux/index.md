---
title: "Uploading Maps to My Garmin Forerunner 965 on Linux"
date: 2026-02-14
lastmod: 2026-02-14
categories:
    - "articles"
tags:
    - "garmin"
    - "linux"
    - "debian"
draft: false
---

I recently got a Forerunner 965 and wanted to load some custom maps for an upcoming trip. Turns out, Garmin devices use MTP instead of standard USB mass storage, so the filesystem doesn't just show up in your file manager.

Here's how I got it working on Debian 13.

## The Setup

First, make sure you have the MTP tools installed:

```bash
apt list --installed | grep -i mtp
```

You should see `mtp-tools`, `jmtpfs`, and `libmtp9`. If not, install them.

## Mounting the Device

```bash
mkdir -p ~/garmin-mount
jmtpfs ~/garmin-mount
```

If you get a "device is busy" error, it's probably already auto-mounted by GVFS. Check with `mount | grep mtp`.

## Copying the Map

Navigate to the mount point and you'll see:

```
~/garmin-mount/
└── Internal Storage/
    ├── GARMIN/          ← Maps go here
    ├── Music/
    └── ...
```

Copy your `.img` map file to the `GARMIN` folder:

```bash
cp Spain_East_OFM.img ~/garmin-mount/Internal\ Storage/GARMIN/
```

That's it. Unmount with:

```bash
fusermount -u ~/garmin-mount
```

## Quick Commands

```bash
# Check if device is detected
lsusb | grep -i Garmin

# Mount
jmtpfs ~/garmin-mount

# Copy map
cp your-map.img ~/garmin-mount/Internal\ Storage/GARMIN/

# Unmount
fusermount -u ~/garmin-mount
```

The map should now show up in your watch under **Settings → Map → Map Manager**.
