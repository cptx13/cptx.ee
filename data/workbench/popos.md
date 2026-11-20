---
title: Pop!_OS
icon: laptop
status: running
order: 4
summary: The computer where everything else is made: Linux, from System76.
hardware: my computer, a USB stick
software: Pop!_OS, COSMIC
---
## guide

Based on System76's [installing Pop!_OS](https://system76.com/support/install-pop). It needs a 64-bit CPU, 4 GB of RAM (8 GB is better) and 20 GB of disk.

1. **Download the right image**: the generic one for Intel/AMD graphics, or the NVIDIA one. Check it against the published checksum.
2. **Flash it to a USB stick** ([live disk creation](https://system76.com/support/pop-live-disk/)).
3. **Turn off Secure Boot** in the firmware settings.
4. **Boot from the stick** with the computer's boot-menu key.
5. **Install**: language and keyboard, then **Clean Install** (wipes the disk) or **Custom** to keep other partitions. Encrypting the disk is offered; say yes on a laptop.
6. **Create the user**, let it finish, take the stick out and reboot.

## notes

- Installed the Intel/AMD image on a machine with an NVIDIA card? `sudo apt install system76-driver-nvidia`.
- A Pop!_OS live stick is also the rescue disk: it can [refresh](https://system76.com/support/live-disk) the system without touching `/home`.
- [Pop!_OS docs](https://pop-os.github.io/docs/)
