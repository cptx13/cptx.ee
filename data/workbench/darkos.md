---
title: R36S + dArkOS
icon: gamepad
status: running
order: 3
summary: A cheap R36S retro handheld, made good by a better operating system.
hardware: R36S, two good microSD cards, a card reader
software: dArkOS, EmulationStation, RetroArch
---
## guide

Based on the [dArkOS wiki](https://github.com/christianhaitian/dArkOS/wiki). dArkOS is the successor of ArkOS, for small Linux handhelds.

1. **Know your R36S.** There are many clones and several screen ("panel") versions; the wrong image or panel file boots to a black screen. For the R36S and its clones there is a dedicated build, [dArkOSRE-R36](https://github.com/southoz/dArkOSRE-R36/wiki), which asks for your variant after flashing.
2. **Back up the stock card** and put it away. The cards these handhelds ship with are the cheapest there are and fail early.
3. **Download the image for the R36S** from the releases page and check its hash.
4. **Flash it to a good microSD card** (Rufus, Raspberry Pi Imager or USB Image Tool), then put it in the handheld. The first boot unpacks the system and takes a while.
5. **Add games and BIOS files** to the card's folders, one folder per system.

## notes

- Use a well-known brand of microSD card, for the system and for games.
- [dArkOS wiki](https://github.com/christianhaitian/dArkOS/wiki) · [dArkOSRE-R36 installation](https://github.com/southoz/dArkOSRE-R36/wiki/Firmware-Installation) · [ArkOS starter guide](https://retrogamecorps.com/2023/03/27/arkos-starter-guide/), most of it still applies
