# Deployment

cptx.ee is deployed on the `cptx` VM as systemd user services under the `exedev` user.

## Automatic deploys (normal path)

Just push to `main` on GitHub (web UI, phone, laptop — anywhere). The VM polls
GitHub itself, so nothing needs to reach into the VM:

- **`cptx-ee-deploy.timer`** — fires every minute
- **`cptx-ee-deploy.service`** — runs `deploy.sh`, which fetches `origin/main`
  and, if it moved since the last good deploy, deploys it safely (below).
  If nothing changed it exits immediately.

Watch a deploy: `journalctl --user -u cptx-ee-deploy -n 20`
(problems are logged with `FAILED:` — `journalctl --user -u cptx-ee-deploy | grep FAILED`)

### Safe deploy

1. `git reset --hard` to the new commit and build it with `mise exec -- go build`
   into `~/.local/bin/cptx.ee.new`. The live binary is not touched.
2. **Build fails** → the worktree goes back to the running commit, the site keeps
   serving the old version, and that commit is not retried. Push a fix.
3. **Build passes** → the live binary becomes `~/.local/bin/cptx.ee.prev`, the new
   one is moved into place, and `cptx-ee.service` is restarted.
4. **Health check**: `http://127.0.0.1:8080/` must answer within 90 s. If it doesn't,
   the previous binary and commit are restored automatically and the bad commit
   is not retried.

State lives in `~/.local/state/cptx-ee/` (`deployed-sha`, `prev-sha`, `failed-sha`).
`deploy.sh` no longer uses `cptx-ee-build.service`; it builds directly.

### Rollback

- **Instant (on the VM):** `~/cptx-v2/deploy.sh rollback` — restores the previous
  binary + commit and restarts. Main on GitHub is unchanged, so it stays rolled
  back only until the next push.
- **Permanent:** `git revert <sha> && git push origin main` — redeploys within ~2 min.

Note: `deploy.sh` hard-resets the worktree, so never leave uncommitted or
unpushed work on the VM — it will be discarded.

The `origin` remote uses a tokenless public HTTPS URL, so deploys need no
credentials, and since 2026-11-15 the VM holds **no GitHub token at all** (no
`~/.git-credentials`, no `credential.helper`). It can fetch but not push: branch
clean-up and anything that writes to GitHub happen from a workstation or the
GitHub UI. If the repo ever goes private, give the VM a fine-grained token
limited to this repo with **Contents: read-only**.

GitHub Actions cannot be used for deploys: hosted runners can't reach this VM.

## Manual deploy

## Prerequisites

- SSH access to `cptx` (configured in `~/.ssh/config`)
- [mise](https://mise.jdx.dev/) installed on the remote at `~/.local/bin/mise`
- The repository cloned to `~/cptx-v2` on the remote
- The `mise.toml` in the repo must be trusted (`mise trust ~/cptx-v2/mise.toml`)

## Systemd units

All four user units are kept in this repo and installed (copied) to
`~/.config/systemd/user/` on the VM:

| Unit | Role | Enabled |
|---|---|---|
| `cptx-ee.service` | Runs `~/.local/bin/cptx.ee` on :8080 from `~/cptx-v2`; restarts on failure | yes (`default.target`) |
| `cptx-ee-deploy.timer` | Fires `cptx-ee-deploy.service` every minute | yes (`timers.target`) |
| `cptx-ee-deploy.service` | Runs `deploy.sh` (safe build → swap → health check) | via the timer |
| `cptx-ee-build.service` | Manual bootstrap build for a fresh VM only | no |

The server does not depend on the build unit: `deploy.sh` builds and swaps the
binary, and the binary in `~/.local/bin` survives reboots, so a reboot simply
starts the last good build.

## Installing / updating unit files (on the VM)

Unit files are *not* applied by a deploy. After a change to any unit reaches
`main` (and the VM has pulled it), run on the VM:

```sh
export XDG_RUNTIME_DIR=/run/user/$(id -u)
cd ~/cptx-v2
cp cptx-ee.service cptx-ee-build.service cptx-ee-deploy.service cptx-ee-deploy.timer ~/.config/systemd/user/
systemctl --user daemon-reload
systemd-analyze --user verify ~/.config/systemd/user/cptx-ee*.service
```

Check that what is installed matches the repo:

```sh
for f in cptx-ee.service cptx-ee-build.service cptx-ee-deploy.service cptx-ee-deploy.timer; do
  diff -q ~/cptx-v2/$f ~/.config/systemd/user/$f || echo "DIFFERS: $f"; done
```

### Fresh VM bootstrap

```sh
mise trust ~/cptx-v2/mise.toml
cp ~/cptx-v2/cptx-ee*.service ~/cptx-v2/cptx-ee-deploy.timer ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user start cptx-ee-build.service      # first binary
systemctl --user enable --now cptx-ee.service cptx-ee-deploy.timer
```

## Restarting the server

```sh
ssh cptx 'export XDG_RUNTIME_DIR=/run/user/$(id -u) && \
  systemctl --user restart cptx-ee.service'
```

## Checking status

```sh
ssh cptx 'export XDG_RUNTIME_DIR=/run/user/$(id -u) && \
  systemctl --user status cptx-ee.service'
```

## Viewing logs

```sh
ssh cptx 'export XDG_RUNTIME_DIR=/run/user/$(id -u) && \
  journalctl --user -eu cptx-ee.service --no-pager'
```

## Before pushing code: `scripts/check.sh`

Run it in a scratch clone (never in `~/cptx-v2`). On the VM:

```sh
git clone -q --depth 1 -b <branch> https://github.com/cptx13/cptx.ee /tmp/chk && cd /tmp/chk
~/.local/bin/mise trust -q mise.toml        # mise ignores an untrusted mise.toml
~/.local/bin/mise exec -- scripts/check.sh
cd ~ && rm -rf /tmp/chk
```

It builds, lints content frontmatter (fails on vault-only fields), starts a test
server on `:18080` (`CHECK_PORT` to change), diffs the generated URLs against
`urls.txt` (fails if any URL would disappear) and requests every URL. Takes ~5 s.
After an *intended* URL change, run it with `--update-urls` and commit `urls.txt`.

## Automation (GitHub Actions)

| Workflow | When | What |
|---|---|---|
| `check` | every push / PR | `scripts/check.sh` (build, vet, tests, lint, URL diff, smoke). On `main` it then waits until `https://cptx.ee/healthz` reports the new commit and runs `scripts/smoke.sh` against the live site — a red run means the deploy didn't land |
| `check` → `lighthouse` | after deploy-verify on `main` | Lighthouse scores for four live pages in the run summary (informational, never fails the run) |
| `check` → `newsletter` | after deploy-verify on `main` | `scripts/newsletter-send.py`: each issue published in the last 21 days becomes one bilingual Buttondown email, once (slug `cptx-issue-NN`). The repository variable `NEWSLETTER_MODE=send` (set 2026-11-15) sends at once, and sends a draft still waiting; without it (`draft`) emails wait in Buttondown for you to send. Needs the `BUTTONDOWN_API_KEY` secret; without it the job does nothing |
| `uptime` | every 15 min | `GET /healthz`, 3 tries; a failed run is emailed by GitHub |
| `linkcheck` | 1st of each month | `scripts/linkcheck.py` — fails on external links that are gone (404/410/DNS) |

Failure emails: GitHub → Settings → Notifications → Actions → "Send notifications for failed workflows only" (email).
GitHub pauses scheduled workflows after 60 days without repo activity; re-enable from the Actions tab.

Manual live check: `scripts/smoke.sh [https://cptx.ee] [commit]`.

## Newsletter (Buttondown)

- **Account:** `cptx` on [Buttondown](https://buttondown.com) (free plan). Open and click tracking and UTM parameters are off — the privacy policy promises it. Tags need a paid plan, so each issue is one email with English, then Spanish.
- **Sign-up form:** `/newsletter/` posts to `https://buttondown.com/api/emails/embed-subscribe/cptx` (`buttondownUser` in `newsletter.go`; empty = form closed). The CSP allows `form-action https://buttondown.com`. Buttondown sends the confirmation email. Its redirects (Settings → Subscribing → Redirects) point to `https://cptx.ee/newsletter/thanks/` (after subscribing) and `https://cptx.ee/newsletter/welcome/` (after confirming); both pages are `newsletterNotices` in `newsletter.go`. Email look: accent `#b54d00`, Modern template, UTM/open/click tracking off.
- **Weekly issue:**
  1. `scripts/newsletter-draft.py` writes `content/newsletter/NNN.md` and `NNN.es.md` (`draft: true`): what changed (commit subjects, to rewrite), what was written, this week years ago, articles kept that week in `data/reads.tsv` (else up to 5 fresh posts from the feed cache), new bedsidetable entries, and empty sections (hidden while empty).
  2. Edit, set `draft: false` in both, add `/newsletter/<n>/` and `/newsletter/<n>/es/` to `urls.txt`, push through the usual branch → check → main.
  3. After the deploy is verified, CI hands the issue to Buttondown and, with `NEWSLETTER_MODE=send`, it goes out to every confirmed subscriber. Approving the issue for the site is approving the email.
- **API key:** GitHub secret `BUTTONDOWN_API_KEY` (Actions). Never in chat, never on the VM.

## Library refresh (`/archives/library/`)

`data/library/books.json` is in git; covers and thumbnails live only on the VM in
`~/cptx-data/library` (outside the repo, so deploys never touch them). Fetches run
on the VM.

1. **Export** the books library from Libib as CSV.
2. **Trim** it to the public columns, same header as today:
   `libib_id,isbn13,isbn10,title,creators,publisher,publish_date,pages`
   (no prices, notes, loans or tags) and replace `data/library/libib-books.csv` on a branch.
3. **Shelve** new books: add `key<TAB>shelf` lines to `shelves.tsv` (key = ISBN-13, or the
   Libib id without one); fixes go in `overrides.tsv`.
4. **Import** on the VM in a scratch clone of that branch:
   ```sh
   git clone -q -b <branch> https://github.com/cptx13/cptx.ee /tmp/lib && cd /tmp/lib
   ~/.local/bin/mise trust -q mise.toml
   python3 scripts/library-import.py ~/cptx-data/library     # only new books are fetched
   cp ~/cptx-data/library/books.json data/library/books.json
   ~/.local/bin/mise exec -- scripts/check.sh --update-urls  # new book pages → urls.txt
   ```
   The import prints the books still without a cover. Hand-made covers go to
   `~/cptx-data/library/covers/<key>.jpg`; the server makes missing thumbnails at start-up
   (when replacing a cover, also delete `~/cptx-data/library/thumbs/<key>.jpg`).
5. **Commit** `libib-books.csv`, `*.tsv`, `books.json` and `urls.txt`; push the branch,
   wait for `check`, then fast-forward `main`. Delete `/tmp/lib`.

`LIBRARY_DIR` / `LIBRARY_MEDIA_DIR` override the two locations; `GOOGLE_BOOKS_KEY` lifts
Google's anonymous quota (the VM's IP is often rate-limited).

## Go version

The Go toolchain is pinned in `mise.toml` (currently `1.27.4`); `go.mod` only sets
the minimum (`1.26.1`). To upgrade: change the version in `mise.toml`, push, and
watch `journalctl --user -u cptx-ee-deploy -n 20`. `mise exec` installs the new
toolchain on the first build; if that build fails, the safe deploy keeps the
current version running.
