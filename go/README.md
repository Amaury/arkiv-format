- **magic.zst**: zstd("arkiv001"), unencrypted.
- **prefix.zst.aes**: zstd(8 random bytes) → OpenSSL enc AES‑256‑CBC (PBKDF2 SHA‑256, 10k).
  - Read path: decrypt → decompress → Base64 (single line) → `PREFIX_BASE64`.
- **index.zst.aes**: text lines, canonical `LC_ALL=C sort -u` byte-wise.
  - Lines: `"PATH"` or `"PATH"=HASH_DATA` (lower hex).
  - Write: escape `\` → `\\`, `"` → `\\"`, then wrap with `"`.
  - Read: **do not unescape** the inner substring when hashing `HASH_NAME`.
- **meta/<HASH_NAME>.tar.zst.aes**: one-entry tar named exactly `PATH` (the raw
  substring), with type+mode+uid+gid+mtime+linkname; size=0 for regular files.
- **data/<HASH_DATA>.zst.aes**: zstd(raw bytes) + OpenSSL enc; dedup by `HASH_DATA`.
- Hashes: `SHA-512/256( PREFIX_BASE64 || bytes )`.

