---
title: "Replace Your Cloud Subscriptions in One Evening: Self-Hosted Photos, Notes, and Files on an Old Laptop"
author: cptx
date: 2026-07-16
draft: false
categories: [croquis]
tags: [self-hosting, workbench]
showTags: true
toc: true
summary: "A focused, evening-length playbook for building a private cloud — Immich for photos, Syncthing for notes and documents, Tailscale for remote access — on an old laptop you already own."
---

If you're paying for iCloud, Google One, Dropbox, *and* Obsidian Sync, the monthly total has a way of creeping up. The good news: a laptop you already have in a drawer can replace most of it. This is a focused playbook for getting a private cloud running in a single evening — about three hours, start to finish.

## What you're building

By the end of the evening you'll have:

- **Immich** as your Google Photos replacement: mobile auto-upload, face/object search, shared albums, web access.
- **Syncthing** as your Dropbox / iCloud Drive / Obsidian Sync replacement: continuous, peer-to-peer file sync for an Obsidian vault, documents, or any folder.
- **Tailscale** as the glue: a free, encrypted private network so you can reach the server from anywhere — your phone on cellular, your laptop at a coffee shop — without opening any ports on your router.

All three are open source. None of them phone home. You own all the data.

## What you need

**A computer that can stay on.** An old laptop is ideal — built-in battery acts as a UPS, low power draw, quiet. As a rough floor, aim for a 4-core CPU from 2018 or newer, 8 GB RAM (16 GB is comfortable, 32 GB is overkill but nice), and an SSD of at least 500 GB. A mini PC, NAS, or Raspberry Pi 5 works too. The instructions below assume a recent-ish Intel laptop running Ubuntu Server; almost everything translates directly to other x86 hardware.

**A USB stick**, 8 GB or larger. It'll get wiped.

**Your normal computer** for downloads and to ssh into the server later.

**A Tailscale account** — free, sign up at [tailscale.com](https://tailscale.com) with whichever single-sign-on you prefer.

**Patience for one long-running task afterward.** The setup tonight is quick. Pulling your existing photo library out of Google or iCloud and importing it into Immich is a separate, multi-hour job you'll want to run on another day.

## Conventions in this guide

Anywhere you see angle brackets, substitute your own value:

- `<your-username>` — the Linux user you'll create during install (keep it short and lowercase, no spaces).
- `<your-hostname>` — what you want to call the server on your network. Something like `home-server` works.
- `<server-tailscale-ip>` — the Tailscale IP assigned to the server after Phase 2. Looks like `100.x.y.z`.
- `<your-device-name>` — display name you give a device when pairing it in Syncthing.

All commands prefixed with `$` are typed on your normal computer. All others run on the server.

## Scope: tonight vs. next time

**Tonight (the must-haves):**

1. Wipe Windows, install Ubuntu Server.
2. Remote access via Tailscale so you can work from your normal laptop.
3. Docker.
4. Immich running, one phone uploading new photos.
5. Syncthing running, a folder (e.g. an Obsidian vault) syncing between server and one other device.

**Deferred to a later session (none of these block having a working system):**

- Hardware-accelerated video transcoding (Intel Quick Sync) in Immich.
- Battery charge thresholds via TLP, to keep a laptop battery healthy long-term.
- Restic + offsite cloud backups.
- Bulk-import of an existing photo library from Google Takeout or iCloud.
- Onboarding additional users (partner, kids, housemates).

## Pre-flight (15 min, before sitting down)

On your normal computer, download:

- Ubuntu Server 24.04 LTS ISO — [ubuntu.com/download/server](https://ubuntu.com/download/server)
- Balena Etcher — [etcher.balena.io](https://etcher.balena.io) (or Rufus on Windows)

Flash the Ubuntu ISO to your USB stick with Etcher. Takes about 5 minutes.

While that runs, sign up for Tailscale and install the Tailscale app on your phone and your normal computer. You can sign in to all of them now.

## Phase 1 — Install Ubuntu Server (30–40 min)

1. Plug the USB stick into the target machine. Power on and open the boot menu (the key varies — F12, F9, Esc, or Del depending on the manufacturer). Pick the USB stick.
2. Choose **Try or Install Ubuntu Server**.
3. Walk through the installer:
   - **Language** and **Keyboard**: your preference.
   - **Network**: connect to your Wi-Fi (or use ethernet if the machine has it).
   - **Proxy**: leave blank.
   - **Mirror**: accept default.
   - **Storage**: choose **Use entire disk**. This wipes the existing OS. Accept the LVM default.
   - **Profile setup**:
     - Your name: anything.
     - Server name: `<your-hostname>`
     - Username: `<your-username>`
     - Password: pick something strong. You'll occasionally need to type it.
   - **Ubuntu Pro**: skip.
   - **SSH**: ✅ **Install OpenSSH server**. This is critical — it's how you'll work after Phase 2.
   - **Featured snaps**: skip everything.
4. Let it install (~10 min). Reboot when prompted. Remove the USB stick when it says to.
5. After reboot, log in at the text console with `<your-username>` and your password.

## Phase 2 — Get remote access working (10 min)

Goal: stop using the server's keyboard. Move to your normal computer.

On the server, find its local IP:

```bash
ip a | grep inet
```

Look for a `192.168.x.x` or `10.x.x.x` address on the active interface. Note it.

Install Tailscale on the server:

```bash
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up
```

It prints a URL. Open the URL on your normal computer and log in. The server now appears at [login.tailscale.com/admin/machines](https://login.tailscale.com/admin/machines) with a Tailscale IP (something like `100.x.y.z`). That's your `<server-tailscale-ip>`.

If you haven't already, install Tailscale on your normal computer ([tailscale.com/download](https://tailscale.com/download)) and sign in with the same account. Both devices now see each other on a private encrypted network, regardless of where they are.

From your normal computer's terminal:

```bash
$ ssh <your-username>@<server-tailscale-ip>
```

You're in. From here on, all commands run over ssh — you can put the server keyboard away.

## Phase 3 — Make it behave like a server (10 min)

If you're using a laptop, tell it to ignore the lid being closed:

```bash
sudo sed -i 's/^#*HandleLidSwitch=.*/HandleLidSwitch=ignore/' /etc/systemd/logind.conf
sudo sed -i 's/^#*HandleLidSwitchExternalPower=.*/HandleLidSwitchExternalPower=ignore/' /etc/systemd/logind.conf
sudo sed -i 's/^#*HandleLidSwitchDocked=.*/HandleLidSwitchDocked=ignore/' /etc/systemd/logind.conf
sudo systemctl restart systemd-logind
```

Update everything:

```bash
sudo apt update && sudo apt upgrade -y
```

If you're using a laptop, close the lid and tuck the machine somewhere with airflow near your router. Your ssh session keeps running.

## Phase 4 — Docker (10 min)

```bash
sudo apt install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
```

Log out and back in so the group change takes effect:

```bash
exit
```

Then ssh back in and test:

```bash
docker run --rm hello-world
```

You should see a hello-world message.

## Phase 5 — Immich (30 min)

Create the app directory and grab Immich's official compose files:

```bash
mkdir -p ~/immich-app
cd ~/immich-app
wget -O docker-compose.yml https://github.com/immich-app/immich/releases/latest/download/docker-compose.yml
wget -O .env https://github.com/immich-app/immich/releases/latest/download/example.env
```

Edit the environment file:

```bash
nano .env
```

Two things to change:

- `UPLOAD_LOCATION=./library` → set an absolute path you'll remember, e.g. `UPLOAD_LOCATION=/home/<your-username>/immich-library`
- `DB_PASSWORD=postgres` → set to something random. Generate one in another terminal with `openssl rand -hex 16`. You will never type this; just paste it in.

Save (Ctrl+O, Enter, Ctrl+X).

Create the library directory:

```bash
mkdir -p /home/<your-username>/immich-library
```

Start Immich:

```bash
docker compose up -d
```

The first run pulls a few GB of images and takes 3–10 minutes. Watch progress with:

```bash
docker compose logs -f
```

When the logs settle, Ctrl+C out of the logs view. Test locally:

```bash
curl http://localhost:2283
```

Should return some HTML. Now from your normal computer's browser, go to:

```
http://<server-tailscale-ip>:2283
```

You'll see the Immich welcome screen. Create the admin account — this is the first user and "owner" of the instance.

## Phase 6 — Phone upload (10 min)

On your phone:

1. Install **Immich** from the App Store / Play Store.
2. Install **Tailscale** from the App Store / Play Store. Sign in with the same account. Toggle the VPN on.
3. Open Immich. For server URL, enter: `http://<server-tailscale-ip>:2283`
4. Log in with the admin account you just created.
5. **Settings → Backup**: enable foreground *and* background backup. Pick the camera roll album. Start the backup.

Photos start uploading. Leave it running while you finish Phase 7. Your full library will take hours or longer the first time — that's fine, it resumes if interrupted.

## Phase 7 — Syncthing for notes and documents (20 min)

Install on the server:

```bash
sudo apt install -y syncthing
sudo systemctl enable --now syncthing@$USER.service
```

Wait 10 seconds and confirm it's running:

```bash
systemctl status syncthing@$USER.service
```

Should say `active (running)`.

Syncthing's web UI only listens on localhost by default. The simplest way to reach it: an ssh tunnel from your normal computer. Open a **new terminal** locally:

```bash
$ ssh -L 8384:localhost:8384 <your-username>@<server-tailscale-ip>
```

Leave that terminal open. In your local browser, visit [http://localhost:8384](http://localhost:8384) — you're looking at the server's Syncthing UI.

On the server's Syncthing UI:

1. **Actions → Settings → GUI**: set a username and password. Save.
2. Remove the "Default Folder" (it points at `~/Sync` and we don't need it).
3. **Add Folder**:
   - Folder Label: `Obsidian Vault` (or whatever you want to sync first)
   - Folder Path: `/home/<your-username>/obsidian-vault`
   - Save.

Now install Syncthing on the device that holds the original copy of the folder you want to sync:

- **macOS**: `brew install syncthing` then `brew services start syncthing`, or use the [Syncthing-macOS app](https://github.com/syncthing/syncthing-macos/releases).
- **Windows**: [SyncTrayzor](https://github.com/canton7/SyncTrayzor/releases).
- **Linux**: `sudo apt install syncthing` and enable the user service as above.

Open Syncthing's web UI on that machine (usually [http://localhost:8384](http://localhost:8384)). Top right, you'll see this device's **Device ID** — a long string. Copy it.

Back on the server's Syncthing UI (still via ssh tunnel):

1. **Add Remote Device** → paste the Device ID → name it `<your-device-name>` → Save.
2. A few seconds later, your other device's Syncthing UI will show a popup: "Device wants to connect" → Accept.
3. On the other device, you'll get a folder share notification → Accept → choose the local path (e.g. `~/Documents/obsidian-vault`).

Wait a couple of minutes. Files appear on both sides. If you're using Obsidian, point it at the synced folder. Done.

For mobile Syncthing — note that the *official* Android app was discontinued in 2024. The actively-maintained community fork is **Syncthing-Fork** on F-Droid or the Play Store. On iOS, the option is **Möbius Sync** (paid, but well-regarded). Add either to your phone the same way you added another desktop.

## Phase 8 — Sanity check (5 min)

Three quick checks before calling it a night:

1. **Photos**: open Immich on your phone, confirm the photo count is climbing. (A full first sync will take many hours — let it run.)
2. **Sync**: create a test file on your other device in the shared folder, confirm it appears on the server within a minute: `ls /home/<your-username>/obsidian-vault`.
3. **Remote access**: disconnect your laptop from home Wi-Fi (use a cellular hotspot or coffee-shop network). ssh in again over Tailscale. It just works.

Close the server's lid. You're running your own cloud.

## What to do next (in priority order)

1. **Backups.** This is the most important follow-up. Plug a USB drive into the server, set up `restic` to snapshot `/home/<your-username>/immich-library` and your synced folders nightly. Then add a second restic repository pointing at an offsite location — Backblaze B2 (~$6/TB/month) or Hetzner Storage Box are popular. Sync is *not* a backup: if a file gets corrupted or deleted, that propagates everywhere.
2. **Battery charge limits** (laptop servers only): `sudo apt install tlp tlp-rdw`, edit `/etc/tlp.conf` to set start/stop charge thresholds around 75/80. Keeps the battery healthy long-term while it acts as your UPS.
3. **Hardware-accelerated video transcoding**: if your CPU has Intel Quick Sync or you have a GPU, Immich's docs have a specific config to enable it. Smoother video playback for not much effort.
4. **Onboard the rest of the household.** Same flow: install Tailscale and Immich on each person's phone, point it at the server. Each user gets their own Immich account.
5. **Bulk-import your old photo library**. Export from Google Photos via Google Takeout or from iCloud via Apple's data export. Use Immich's CLI ([immich-go](https://github.com/simulot/immich-go) is the community favorite) to bring everything in. Plan to run this overnight.
6. **Cancel the cloud subscriptions** — but only after you've verified the setup for a couple of weeks and your backups are working.

## Troubleshooting cheatsheet

**Can't ssh in after reboot.** The server's *local* IP may have changed. Always use the *Tailscale* IP — it's stable.

**Immich web UI won't load.** From an ssh session: `cd ~/immich-app && docker compose ps` to see if containers are up. `docker compose restart` to bounce them. `docker compose logs --tail=100` to see what's wrong.

**Syncthing folders show "Out of Sync" forever.** Almost always a permissions issue. Check the folder is owned by your user: `ls -la /home/<your-username>/obsidian-vault`. If not: `sudo chown -R <your-username>:<your-username> /home/<your-username>/obsidian-vault`.

**Fan running constantly the first day.** Normal. Immich is generating thumbnails and running machine learning on every photo. It quiets down once the initial pass finishes.

**Updating Immich later.** `cd ~/immich-app && docker compose pull && docker compose up -d`. Monthly-ish is fine.

**Updating Ubuntu later.** `sudo apt update && sudo apt upgrade -y`. Monthly-ish. Reboot if it asks.

## Closing thoughts

Self-hosting trades a monthly fee for an upfront time investment plus the responsibility of being your own IT. For most people, that's a good trade only if (a) you find it interesting, and (b) you actually run the backups. Don't skip the backups.

The setup above is intentionally minimal. There's a whole rabbit hole of additional services — Nextcloud for collaborative documents, Jellyfin for movies, Vaultwarden for passwords, Home Assistant for smart-home stuff — that run happily on the same machine once you're comfortable with Docker. But none of that is necessary to start cancelling subscriptions tonight.

Good luck.
