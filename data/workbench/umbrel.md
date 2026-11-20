---
title: UmbrelOS
icon: server
status: running
order: 5
summary: A home server on an old ThinkPad.
kit: an old ThinkPad, UmbrelOS, Immich, a USB stick, an ethernet cable
---
## how it's set up

The steps follow Umbrel's own [guide for x86 computers](https://github.com/getumbrel/umbrel/wiki/Install-umbrelOS-on-x86-systems). An old laptop is enough: a 64-bit dual-core CPU, 4 GB of RAM (8 GB is better) and 32 GB of disk to start, more for photos.

1. **Make the installer.** Download the umbrelOS USB installer and flash it to a USB stick (4 GB or more) with balenaEtcher. The stick is wiped.
2. **Boot the ThinkPad from the stick**, with the ethernet cable plugged into the router. If it starts the old system instead, change the boot order in the BIOS (on ThinkPads, F12 at power-on opens the boot menu).
3. **Pick the internal disk** when the installer asks: type its number and press enter. Everything on that disk is erased.
4. **Shut down, take the stick out, start again.** The first boot takes a few minutes. Ignore the `umbrel login` prompt on the screen; it is only for the terminal.
5. **Open `http://umbrel.local`** from any device at home (or the laptop's IP address, from the router's list) and create the account.

## Immich, for photos

1. In umbrelOS, open the **App Store** and install [Immich](https://apps.umbrel.com/app/immich).
2. Open it and create the admin account.
3. Install the Immich app on the phone, point it to the server, and turn on backup.

## resources

- [Install umbrelOS on x86 systems](https://github.com/getumbrel/umbrel/wiki/Install-umbrelOS-on-x86-systems)
- [Installing an app on Umbrel](https://umbrel.com/support/apps/how-to-install-an-app)
- [Immich](https://immich.app/)
- [Build your own decentralized Dropbox with Syncthing](https://xn--gckvb8fzb.com/build-your-own-decentralized-dropbox-with-syncthing/)
