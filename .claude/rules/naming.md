# Naming

- `MixedCaps`, not underscores. Exported = capital first letter.
- Package name: short, lowercase, no underscore, no plural (`user` not `users`, not `userutil`).
- No stutter: avoid `user.User` → prefer `user.Profile` or just `User` inside pkg `user`.
- Interface names: single-method interfaces get `-er` suffix (`Reader`, `Writer`, `Hasher`).
- Getter: `Name()` not `GetName()`. Setter: `SetName()`.
- No `utils`/`common`/`helpers` dumping-ground package — name by what it does.
