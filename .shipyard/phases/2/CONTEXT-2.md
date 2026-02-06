# Phase 2: Discussion Decisions

## OCR Checksums (R1)
- **Decision**: Download each tessdata_fast .traineddata file and compute SHA256 checksums locally
- **Rationale**: Most reliable approach — ensures checksums match the exact files users will download from the tessdata_fast repo

## Password Flag Error Message (R2)
- **Decision**: List all three alternatives in the error message when --password is used without --allow-insecure-password
- **Alternatives to list**: `--password-file`, `PDF_CLI_PASSWORD` environment variable, and interactive prompt
- **Rationale**: Gives users full context on all secure password input methods
