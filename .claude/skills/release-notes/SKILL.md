---
name: release-notes
description: Generate release notes for the gon project by comparing the latest GitHub release with HEAD. Use this skill whenever the user asks to generate release notes, buat release notes, create a release, prepare a changelog, or wants to publish a new version. Also trigger when user mentions "release", "release notes", "changelog", "versi baru", "gh release", or any variant of publishing a new version of gon. This skill handles the full workflow: detecting new commits since the last release, categorizing changes, generating formatted notes in the established template, and optionally creating a GitHub release with `gh release create`.
---

## Overview

This skill generates release notes for the **gon** project by:
1. Finding the latest GitHub release and its tag
2. Collecting commits since that tag
3. Categorizing commits into the established release note sections
4. Generating formatted notes in Indonesian with emoji sections
5. Optionally publishing via `gh release create`

---

## Step 1: Discover latest release and new commits

```bash
# Get latest release tag
LATEST_TAG=$(gh release list --limit 1 --json tagName -q '.[0].tagName')
echo "Latest release: $LATEST_TAG"

# Fetch tags so git can compare
git fetch --tags --quiet

# Commits since the last release
git log ${LATEST_TAG}..HEAD --oneline
```

If `git log` returns empty, tell the user: "Tidak ada commit baru sejak release $LATEST_TAG — belum ada yang perlu di-release."

Also view the latest release notes for reference context:
```bash
gh release view $LATEST_TAG
```

---

## Step 2: Collect full commit info

Get the full commit messages to understand changes in depth:

```bash
git log ${LATEST_TAG}..HEAD --format="%H %s%n%b" --no-merges
```

Also look at the actual diffs for context if needed:
```bash
git log ${LATEST_TAG}..HEAD --name-only --format="" | sort | uniq
```

---

## Step 3: Determine next version

Suggest a version bump based on the nature of changes:

| Change type | Version bump |
|-------------|-------------|
| Breaking changes present | Minor (e.g., v0.4.0 → v0.5.0) |
| New features only | Minor |
| Bug fixes / internal only | Patch (e.g., v0.4.0 → v0.4.1) |

Ask the user to confirm the version: "Berdasarkan perubahan yang ada, saya sarankan **vX.Y.Z**. Apakah versi ini OK?"

---

## Step 4: Categorize commits

Map commits to sections based on keywords in commit messages:

| Commit keywords | Section |
|----------------|---------|
| `feat`, `add`, `new`, `implement` | 🚀 New Features |
| `fix`, `bug`, `patch`, `resolve` | 🐛 Bug Fixes |
| `refactor`, `restructure`, `rename`, `reorganize`, `migrate`, `move`, `hexagonal`, `architecture` | 🏗️ Architecture |
| `break`, `remove`, `drop`, `delete` (user-facing) | ♻️ Breaking Changes |
| `chore`, `internal`, `test`, `ci`, `build`, `docs`, `install`, `config`, `clean` | 🔧 Internal |

Use judgment — a single commit can touch multiple concerns. Prioritize the most user-visible section.

---

## Step 5: Generate release notes

Write the release notes in **Bahasa Indonesia** following this exact template:

```markdown
## What's Changed

### 🚀 New Features

- **Feature name** — deskripsi singkat apa yang berubah atau ditambahkan, dari sudut pandang pengguna

### 🏗️ Architecture

- **Perubahan arsitektur** — jelaskan strukturnya secara singkat

### ♻️ Breaking Changes

- **Hal yang berubah** — apa yang bisa break, dan apa yang harus dilakukan pengguna

### 🔧 Internal

- Deskripsi perubahan internal

### 🐛 Bug Fixes

- **Bug yang diperbaiki** — jelaskan apa yang salah sebelumnya
```

**Rules:**
- Only include sections that have content — skip empty sections
- Descriptions should be clear and user-facing, not just repeating the commit message
- Use **bold** for the feature/change name, then `—` (em dash) then description
- Write in Indonesian (same style as existing releases)
- Each bullet point should give enough context to understand the change without reading the code

---

## Step 6: Review with user

Present the generated release notes to the user and ask:
- "Apakah versi dan release notes-nya sudah sesuai?"
- "Ada yang mau ditambah, diubah, atau dihapus?"

Wait for confirmation before proceeding to publish.

---

## Step 7 (optional): Publish GitHub release

If the user wants to publish with `gh release`, run:

```bash
gh release create <VERSION> \
  --title "<VERSION>" \
  --notes "<RELEASE_NOTES>" \
  --latest
```

If the user wants to also attach build artifacts (binaries), check if there's a CI workflow that handles it automatically (`.github/workflows/release.yml`). If yes, tell the user the release tag will trigger the CI build. If not, inform them binaries need to be built and attached separately.

Check the release workflow:
```bash
cat .github/workflows/release.yml 2>/dev/null || echo "No release workflow found"
```

After publishing:
```bash
gh release view <VERSION>
```

Confirm the release URL and show it to the user.

---

## Example release notes output

```markdown
## What's Changed

### 🚀 New Features

- **Custom help command** — command `help` kini menampilkan daftar perintah yang tersedia langsung dari REPL
- **Header support** — flag `--header` kembali hadir, memungkinkan pengiriman custom HTTP header

### 🐛 Bug Fixes

- **CLI command lookup** — perbaikan bug saat command tidak ditemukan di registry menyebabkan panic
- **Output formatting** — hasil response kini ditampilkan dengan benar untuk semua metode HTTP

### 🔧 Internal

- Refactor validasi input di REPL untuk penanganan error yang lebih konsisten
- Perbaikan konstruksi argumen CLI untuk one-shot mode
```
