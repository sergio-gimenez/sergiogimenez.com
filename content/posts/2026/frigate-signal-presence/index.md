---
title: "Camera Alerts on Signal, and the Presence Bug That Woke Us at 07:09"
date: 2026-09-08
draft: true
description: "Person detection with Frigate on an N100, alert clips delivered over Signal, and why Home Assistant's OpenWrt integration kept telling me my house was empty."
tags: ["homelab", "home-assistant", "frigate", "openwrt", "signal", "presence"]
categories: ["articles"]
series: ["Homelab Networking From Scratch"]
series_order: 2
---

# Camera Alerts on Signal, and the Presence Bug That Woke Us at 07:09

This morning my partner's phone got a Signal message with a video clip of our hallway, timestamped 07:09, because my alarm system had decided the flat was empty. Both of us were in it.

The camera was right and the person detection was right. What was wrong was Home Assistant's idea of who was home, and fixing that turned out to be the interesting part of the project, so it's most of what this post is about.

## What I was actually trying to build

One Tapo C200 in the hallway. When it sees a person and nobody is home, I want a message with the video clip, on a phone, within a few seconds. No camera cloud subscription, no Home Assistant exposed to the public internet, and no dependency on being connected to my VPN at the moment the alert fires.

That last constraint is what makes this awkward, and it's worth explaining before the parts list makes sense.

My Home Assistant is reachable at `assistant.home.sergiogimenez.com` through Pi-hole and Nginx Proxy Manager, and from outside the house only over the Netmaker mesh. I'm not opening it to the internet; the standard route for push notifications is the Companion app, which normally gets its notifications through Google's Firebase servers. But I run GrapheneOS and installed the app from F-Droid, which is the `minimal` build with no Firebase in it at all. That build only receives notifications while its WebSocket to HA is alive, which means only while the VPN is up, and I can't keep the VPN up permanently because it fights with Android Auto in the car.

So: the notification has to leave my house by some path that doesn't involve my phone reaching Home Assistant. Signal it is.

## Frigate on the N100, and a pile of small surprises

Detection runs in an LXC on a GMKtec N100 node, with `/dev/dri` passed through so the iGPU does the work. Inference lands around 10ms per frame, which for one camera is barely a rounding error on that little box.

The install did not go smoothly, and none of the problems were interesting enough to deserve the time they took, so here they are compressed:

The community-scripts installer built me a container and then didn't install Frigate in it. Bare Debian, no docker, no service. I finished it by hand and got a native systemd install rather than the containerised one, which matters for the next bullet.

That native install reads `/config/config.yml`. There is also a `/opt/frigate/config/config.yml` which looks exactly like the file you want, and is ignored. I edited the wrong one for a while and could not understand why my changes did nothing.

Frigate 0.17 changed the config schema in ways that fail validation and then quietly drop you into safe mode with a CPU detector: `record.events` became `record.alerts` and `record.detections`, and `live.streams` went from a list to a mapping. Safe mode itself was caused by the CPU detector's model file not existing in this install, which is a strange failure to hit on a machine that has a perfectly good iGPU sitting there. The fix is to not use the CPU detector:

```yaml
detectors:
  ov:
    type: openvino
    device: GPU
model:
  width: 300
  height: 300
  input_tensor: nhwc
  input_pixel_format: bgr
  path: /openvino-model/ssdlite_mobilenet_v2.xml
  labelmap_path: /openvino-model/coco_91cl_bkgr.txt
```

Then the best one. Everything was healthy, the camera was streaming, the detector was warm, and no events ever fired. `detect.enabled` defaults to `false`. I walked past that camera a lot of times before I found it.

The working camera block records the 1080p main stream and detects on the 720p substream at 5fps, which keeps the N100 idle:

```yaml
cameras:
  pasillo:
    ffmpeg:
      inputs:
        - path: rtsp://127.0.0.1:8554/pasillo
          input_args: preset-rtsp-restream
          roles: [record]
        - path: rtsp://127.0.0.1:8554/pasillo_sub
          input_args: preset-rtsp-restream
          roles: [detect]
    detect:
      enabled: true
      width: 1280
      height: 720
      fps: 5
    objects:
      track: [person]
```

One more thing that isn't a bug: framing matters more than I expected. My first real test failed because the camera was pointed at a desk and caught a sliver of my shoulder at the edge of frame. MobileNet wants a head and a torso. Point the camera at the walking path, not at the room.

## Getting a video clip into Signal

Signal has no official bot API. What exists is `signal-cli`, and around it `signal-cli-rest-api`, which you can run in Docker and talk to over HTTP. You link it as a device on an existing account, exactly like Signal Desktop, so no second phone number is needed.

The catch, and it's a real one: a linked device sends *as you*. Messages land in the conversation from your side, so my own phone treats them as sent messages and stays silent. My partner's phone buzzes. For our case that's fine, since the point is that one of us finds out; if you want your own phone to ring you need a separate account on its own number.

Setup is a small LXC, Docker, and the REST API in `json-rpc-native` mode. Linking is a QR code from `/v1/qrcodelink`, scanned from Signal on the phone under Linked devices. Be quick about it: the QR is single-use and expires in about two minutes, and fetching a new one invalidates the old. My first attempt died to exactly that.

Sending text worked immediately. Sending the clip did not, and the failure was silent, which is the worst kind. The API accepts an `attachments` field containing URLs, returns HTTP 201, and delivers a message with no attachment. It only actually sends files given as `base64_attachments`. So there's a small Python service on the same container listening on port 8081: it takes a Frigate event id, fetches `http://192.168.34.102:5000/api/events/<id>/clip.mp4`, base64s it, and posts that to the Signal API. Home Assistant calls it with a `rest_command`:

```yaml
rest_command:
  signal_alert:
    url: http://192.168.34.210:8081/alert
    method: POST
    headers:
      Content-Type: application/json
    payload: >-
      {"event_id": "{{ event_id }}"}
```

A 4MB clip goes through without complaint.

## The automation, and the condition that broke everything

```yaml
- id: frigate_pasillo_person_away
  alias: Frigate - persona en pasillo estando fuera
  mode: queued
  max: 5
  triggers:
    - trigger: mqtt
      topic: frigate/events
  conditions:
    - condition: template
      value_template: >-
        {{ trigger.payload_json.type == 'end'
           and trigger.payload_json.after.camera == 'pasillo'
           and trigger.payload_json.after.label == 'person'
           and trigger.payload_json.after.has_clip }}
    - condition: numeric_state
      entity_id: zone.home
      below: 1
  actions:
    - variables:
        event_id: "{{ trigger.payload_json.after.id }}"
    - action: rest_command.signal_alert
      data:
        event_id: "{{ event_id }}"
```

`zone.home` below 1 means zero people home, which is nicer than naming individuals because adding a person to Home Assistant automatically extends the rule. It's also the line that sent the 07:09 clip, because Home Assistant sincerely believed the flat was empty.

## Why the presence was wrong

Phone GPS was never an option here, for the same reason push wasn't: the phones can't reach Home Assistant from outside. Presence had to come from the network itself, from the router seeing the phones on Wi-Fi.

Home Assistant has an OpenWrt integration in the UI, called **luci**. I added it against my main router, assigned the trackers to each person, and watched it flap. Both phones would go home, away, home, away, on a cycle of thirty to sixty seconds. Overnight it was worse: hours of "away" while the phones sat two metres from me on a bedside table.

Two separate things were going on, and I only found the second one after fixing the first.

My main router at `192.168.34.1` has no radios. It's a wired gateway; the Wi-Fi comes from a Cudy AP at `.195`. So the trackers I'd built presence on were reading DHCP leases and the ARP table, not Wi-Fi at all, and a sleeping phone that hasn't sent a packet in ten minutes falls out of both. If you look at the `openwrt_luci_rpc` library the integration uses, it filters on the neighbour table's `reachable` flag, and dozing phones are not reachable. That behaviour is fine for a laptop; it's useless for presence.

Adding the AP as a second luci instance didn't help either, and the reason is the same one: the AP's ARP table only knows about devices talking to the AP itself. My phone is associated to its radio, plainly visible in `iwinfo assoclist`, and completely absent from anything luci reads.

So I wrote my own. A shell script on the AP, polling `ubus call iwinfo assoclist` on both radios every twenty seconds, with a 300 second grace period, publishing MQTT discovery messages to the Mosquitto broker that Frigate already uses. It worked on the first try, it was about forty lines, and I was pleased with myself right up until I went looking for prior art.

Then I checked whether Home Assistant already had something, and it does. There are *two* OpenWrt integrations in Home Assistant, and I'd only found one:

- **luci**, in the UI, config-flow, reads ARP and DHCP leases. This is the one everybody finds.
- **ubus**, YAML only, marked `quality_scale: legacy`, no UI, no documentation link that shows up when you search for "openwrt home assistant". It calls `hostapd get_clients`, which is genuine Wi-Fi association.

Reading `homeassistant/components/ubus/device_tracker.py` confirmed it, so I deleted my script, removed the package I'd installed on the AP, cleared the retained MQTT topics, and replaced the lot with this:

```yaml
device_tracker:
  - platform: ubus
    host: 192.168.34.195      # the AP with the radios, not the gateway
    username: root
    password: !secret cudy_password
    dhcp_software: none       # this AP serves no DHCP
    consider_home: 300
    new_device_defaults:
      track_new_devices: false
```

`dhcp_software: none` is what makes it work on an access point with no DHCP server, and it's the setting I expected to be a problem and turned out to be the whole trick. Entity names come from `known_devices.yaml`, where I named the two phones and left everything else untracked.

`track_new_devices: false` I added afterwards, when every new device that associated to the AP started throwing `The see service is not supported for this entity` into the log every twelve seconds, because the entity id it wanted was already owned by a disabled luci entry.

I then polled both trackers every 30 seconds for ten minutes:

```
15:29:49 home not_home home not_home 1
15:30:19 home not_home home not_home 1
...
15:39:12 home not_home home not_home 1
```

Twenty samples, zero transitions, on the same phones that had been flipping every minute an hour earlier.

## The honest accounting

The whole stack is one LXC for Frigate, one for the Signal bridge, an MQTT broker, five lines of YAML for presence, and one automation. It's stable and it costs nothing per month.

It also exists entirely because I won't pay 6.50 € a month for Nabu Casa, which would give me remote GPS presence from both phones and push notifications with the clip attached, without any of the above. If you don't have a specific reason to keep everything local, that subscription is a much better use of a Saturday than what I just described.

I do have a reason, and I got something out of the detour that the subscription wouldn't have given me: I now know that association-based presence is a different thing from lease-based presence, that Home Assistant ships both and only advertises the worse one, and that the useful one is hidden behind the word "legacy". If you've been fighting phones that go away while they're sitting on your table, that's probably your bug too.

## Still open

Both people should carry a second tracker, so that rebooting the access point doesn't instantly mark the house empty and fire an alert at nobody. The lease-based luci trackers are actually good for that, as a backup that only ever says "home".

And I want to point the same automation at a second camera, which means finding out whether the N100 stays this relaxed with two streams. I suspect it will.
