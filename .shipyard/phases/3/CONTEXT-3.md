# Phase 3: Discussion Decisions

## Password File Binary Content (R9)
- **Decision**: Warning only — print warning to stderr but still return the password content
- **Rationale**: Avoids breaking users who legitimately use binary-looking passwords. A hard error would be too strict.
