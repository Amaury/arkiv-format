# Arkiv format

Documentation and tools for the "Arkiv file format", a tar-like deduplicated, compressed and encrypted file format based on standard Unix tools.

> **TL;DR**
>
> Arkiv format is an **experimental, educational** archive format that combines
> **deduplication**, **encryption**, and **compression** in a way that stays
> friendly to **standard Unix tools**. Archives are **immutable**,
> internally a plain **tar** with a small, well-defined structure, and allow
> **partial extraction** (a single file, a subtree, or the whole archive)
> without decrypting/decompressing everything.

---

## 1. Introduction

**Arkiv format** demonstrates how to combine **deduplication**, **encryption**,
and **compression** while remaining **fully operable with standard Unix tools**.

Compared to a classic **tar** file:

- Tar has **no deduplication**.
- Tar has **no built‑in encryption**.
- Tar has **no built‑in compression**.
- Tar does **not** perform **integrity checks** of archived contents.

Arkiv format adds all of the above.

Arkiv archives are **immutable**: once created, they are not modified in place.

---

## 2. Goals & philosophy

- **Unix‑friendly:** operable with `tar`, `cut`, `sed`, `openssl`, `zstd`, etc.  
- **Transparent:** readable and manipulable even without helper scripts.  
- **Simple structure:** a plain `.tar` file with a few well‑defined members.  
- **Immutability:** the archive is built in one pass and then considered read‑only.

---

## 3. Algorithms & tooling

- **Signatures:** `SHA‑512/256`  
  - Typically **faster** than `SHA‑256` and `SHA‑512` on modern **64‑bit** CPUs.  
  - Produces a compact 256‑bit digest with strong security properties.  
  - Used for naming `data/` blobs (file content) and `meta/` blobs (path name).

- **Compression:** `zstd`  
  - Very fast, good compression ratio, **built‑in integrity checking**.

- **Encryption:** `AES‑256‑CBC` via `openssl`  
  - With `-pbkdf2 -md sha256 -salt`.  
  - A single password (provided via `ARKIV_PASS` environment variable) encrypts/decrypts all encrypted members.

---

## 4. Archive structure

Arkiv files use the extension **`.arkiv`**.  
An Arkiv archive is a **plain tar** file that contains:

```
backup.arkiv (tar)
├── magic.zst                 # compressed "arkiv001"
├── prefix.zst.aes            # encrypted, compressed 8-byte salt for hashing
├── index.zst.aes             # encrypted, compressed plaintext index
├── meta/
│   ├── 5059…945e.tar.zst.aes # metadata tar for file1
│   ├── 9ab1…c02f.tar.zst.aes # metadata tar for file2
│   └── …                     # (one per index entry)
└── data/
    ├── 7f…21.zst.aes         # data blob for syslog content
    └── …                     # (only for regular files)
``` 

### 4.1 `magic.zst`
- Compressed file containing the literal string: `arkiv001`
- Used to quickly **identify** a valid Arkiv archive.
- `001` is the format version number.
- The file is compressed in order to check its integrity with zstd.

### 4.2 `prefix.zst.aes`
- Encrypted and compressed file containing **8 random bytes**.  
- This 8‑byte value **salts all signatures**: its base 64-encoded version is **prepended**
to the exact byte sequence being hashed (path text for `HASH_NAME`, file data for `HASH_DATA`).

### 4.3 `index.zst.aes`
- Encrypted, compressed **plaintext index** listing every path in the archive.  
- Two line forms (no spaces):
  - **Directories and special files (symlinks, FIFOs):**
    ```
    "PATH"
    ```
  - **Regular files (with data blob):**
    ```
    "PATH"=HASH_DATA
    ```

**Example (hashes truncated for readability):**
```
"/etc/cron.d"
"/etc/cron.hourly/logrotation"=7f…21
"/etc/mtab"
"/home/user/save/logrotation"=7f…21
```

Notes:
- `/etc/cron.d` is a directory; `/etc/mtab` is a symbolic link.
- `/etc/cron.hourly/logrotation` and `/home/user/save/logrotation` share the **same content** (`HASH_DATA`), illustrating **deduplication**.
- **Directories and special files** do **not** have a corresponding entry in `data/` (only regular files do).

### 4.4 `meta/`
- Stores **metadata** for **every** path listed in the index.  
- Each entry is an encrypted & compressed **one‑file tar** named exactly like the original path:
  ```
  meta/<HASH_NAME>.tar.zst.aes
  ```
- That inner tar contains **one member** whose type and attributes (mode, uid, gid, mtime, symlink target, etc.) encode the **file kind and metadata**.

### 4.5 `data/`
- Stores the **content blobs** for **regular files** only:
  ```
  data/<HASH_DATA>.zst.aes
  ```
- Each blob is **compressed** with zstd, then **encrypted** with openssl.  
- Multiple paths can reference the **same** `HASH_DATA` file → **deduplication**.

---

## 5. Deduplication

- Regular‑file contents are addressed by `HASH_DATA`.  
- If two files have identical bytes, they **share** the same `data/<HASH_DATA>.zst.aes`.  
- **Metadata is per path** (in `meta/`), so permissions/ownership/timestamps can differ even for identical contents.

---

## 6. Extraction

Because the index lists every path and is **sorted** lexicographically (`sort`), you can:

1. **Extract a single file**.  
2. **Extract a whole subtree** (e.g. `"/etc/cron.d"` and everything beneath it).  
3. **Extract the entire archive**.

Example — extract one subtree:
```sh
./arkiv-extract backup.arkiv /tmp/cron.d "/etc/cron.d"
```

Example — extract the entire archive:
```sh
./arkiv-extract backup.arkiv /restore
```

> Under the hood, only the required `meta/…` and `data/…` members are read, decrypted, and decompressed.
> You **do not** need to decrypt/decompress the entire tar to restore a subset.

---

## 7. Integrity & security

**Integrity:**
- `zstd` performs integrity checks when decompressing `magic.zst`, `prefix.zst.aes`, `index.zst.aes`, `data/…` and `meta/…` blobs.
- The content‑addressed layout also enables verifying a regular file’s data by recomputing `SHA‑512/256(prefix || data)` and comparing it to `HASH_DATA` from the index.

**Security:**
- Every sensitive member (`prefix.zst.aes`, `index.zst.aes`, each `meta/*.tar.zst.aes`, each `data/*.zst.aes`) is encrypted with `AES‑256‑CBC` using PBKDF2 (`-md sha256 -salt`) via `openssl`.

---

## 8. Command reference

### 8.1 `arkiv-build`
**Synopsis**
```sh
arkiv-build ARCHIVE.arkiv PATH...
```

**Description**

Builds an **immutable** Arkiv archive from the given inputs.

For each input path:
- If it is a **directory (not a symlink)**, Arkiv includes the directory itself **and** all its descendants (recursively, symlinks not followed).
- For **every path** (file/dir/symlink/FIFO), Arkiv writes a **meta** entry (`meta/<HASH_NAME>.tar.zst.aes`) capturing the metadata (type, mode, uid/gid, mtime, link target…).
- For **regular files**, Arkiv writes or reuses a **data** blob (`data/<HASH_DATA>.zst.aes`) with **zstd** + **AES‑256‑CBC**; identical contents are deduplicated.
- `index.zst.aes` receives one line per path:
  - Regular file: `"PATH"=HASH_DATA`
  - Directory / symlink / FIFO: `"PATH"`

**Environment**

- `ARKIV_PASS`: password used to encrypt all members (except `magic.zst`).

**Examples**

```sh
# Build an archive from a directory and two files
ARKIV_PASS='s3cr3t' arkiv-build backup.arkiv /etc /var/log/syslog /home/user/notes.txt
```

### 8.2 arkiv-tree
**Synopsis**

```sh
arkiv-tree ARCHIVE.arkiv [PREFIX]
```

**Description**

Lists archive entries (like `ls -l`) from the metadata:

- If `PREFIX` is given, only entries under that subtree are listed.
- Uses metadata tars (`meta/<HASH_NAME>.tar.zst.aes`) to display file type, permissions, ownership, and timestamps.

**Environment**

`ARKIV_PASS`: password used to decrypt all members.

**Examples**

```sh
# List entire archive
ARKIV_PASS='s3cr3t' arkiv-tree backup.arkiv

# List a subtree
ARKIV_PASS='s3cr3t' arkiv-tree backup.arkiv /etc/cron.d
```

### 8.3 arkiv-extract
**Synopsis**

```sh
arkiv-extract ARCHIVE.arkiv [DEST] [PREFIXES]
```

**Description**

Extracts the whole archive, one file or a complete subtree into DEST:
- The tool selects the exact `"PATH"` entry and, if it’s a directory, all entries beneath it.
- For each selected entry, it restores the type and metadata (best‑effort), and for regular files it restores the content from `data/<HASH_DATA>.zst.aes`.

**Environment**

- `ARKIV_PASS`: password used to decrypt all members.

**Examples**

```sh
# Extract the whole archive in the current directory
ARKIV_PASS='s3cr3t' arkiv-extract backup.arkiv

# Extract the whole archive in the given destination directory
ARKIV_PASS='s3cr3t' arkiv-extract backup.arkiv ./restore/

# Extract a single file
ARKIV_PASS='s3cr3t' arkiv-extract backup.arkiv ./restore/notes.txt "/home/user/notes.txt"

# Extract a whole directory recursively
ARKIV_PASS='s3cr3t' arkiv-extract backup.arkiv ./restore/cron.d "/etc/cron.d"
```

---

## 9. Working without the helper scripts

You can manipulate an Arkiv archive using only Unix tools.

**Integrity check**
```sh
$ tar xOf backup.arkiv magic.zst | zstd -q -d -c
```
Should print:
```sh
arkiv001
```

**Definition of ARKIV_PASS environment variable**
```sh
$ export ARKIV_PASS="s3cr3t"
```

**Inspect the index:**
```sh
$ tar xOf archive.arkiv index.zst.aes \
  | openssl enc -d -aes-256-cbc -pbkdf2 -md sha256 -pass env:ARKIV_PASS \
  | zstd -q -d -c \
  | sed -E 's/^"([^"]*)".*/\1/'
```

**Extract a data blob:**
```sh
# fetch data hash for file 'foo.txt':
# extract file 'index.zst.aes' from archive
# + decrypt + decompress + get data hash
$ HASH_DATA="$(tar xOf archive.arkiv index.zst.aes \
  | openssl enc -d -aes-256-cbc -pbkdf2 -md sha256 -pass env:ARKIV_PASS \
  | zstd -q -d -c \
  | grep '"foo.txt"=' \
  | cut -d'=' -f2)"

# fetch data content of file 'foo.txt':
# extract file 'data/$HASH_DATA.zst.aes' from archive
# + decrypt + decompress
$ tar xOf archive.arkiv "data/$HASH_DATA.zst.aes" \
  | openssl enc -d -aes-256-cbc -pbkdf2 -md sha256 -pass env:ARKIV_PASS \
  | zstd -d -o "foo.txt"
```

**Apply metadata to extracted file:**
```sh
# get prefix:
# extract file 'prefix.zst.aes' from archive
# + decrypt + base 64 encode
$ PREFIX="$(tar xOf archive.arkiv prefix.zst.aes \
  | openssl enc -d -aes-256-cbc -pbkdf2 -md sha256 -pass env:ARKIV_PASS \
  | zstd -q -d -c \
  | openssl base64 -A)"

# compute name hash for file 'foo.txt'
$ HASH_NAME="$(printf "%s%s" "$PREFIX" "foo.txt" \
  | openssl dgst -r -sha512-256 | cut -d' ' -f1)"

# fetch metadata file (under the name 'foo.txt.meta/foo.txt'):
# extract file 'meta/$HASH_NAME.tar.zst.aes' from archive
# + decrypt + decompress
$ mkdir foo.txt.meta
$ tar xOf archive.arkiv "meta/$HASH_NAME.tar.zst.aes" \
  | openssl enc -d -aes-256-cbc -pbkdf2 -md sha256 -pass env:ARKIV_PASS \
  | zstd -q -d -c \
  | tar -xpf - "foo.txt" -C foo.txt.meta
$ META_PATH="foo.txt.meta/foo.txt"

# get metadata
$ MODE="$(stat -c '%a' "$META_PATH" 2>/dev/null || echo 0644)"
$ UID="$(stat -c '%u' "$META_PATH" 2>/dev/null || echo 0)"
$ GID="$(stat -c '%g' "$META_PATH" 2>/dev/null || echo 0)"
$ MTIME="$(stat -c '%Y' "$META_PATH" 2>/dev/null || date +%s)"
$ TIMESTAMP="$(date -d "@$MTIME" "+%Y%m%d%H%M.%S")"

# update of extracted file's rights, user/group and modification date
$ chmod "$MODE" "foo.txt"
$ chown "$UID:$GID" "foo.txt" # may fail if not root
$ touch -t "$TIMESTAMP" "foo.txt"

# delete temporary file
$ rm -rf "foo.txt.meta"
```
---

## Appendix A. License

Copyright © 2025, Amaury Bouchard <amaury@amaury.net>

Published under the terms of the MIT license.

