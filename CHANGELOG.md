## v1.1.2 (2025-05-27)

## v1.1.1 (2025-05-27)

## v1.1.0 (2025-05-27)

### Feat

- **winres**: add Windows rescourse files
- **clublog**: init provider
- **git**: init provider
- **log**: optimize log display

### Fix

- **git**: use auth on pull and push
- **git**: parse config in main
- **log**: remove dup provider field

### Refactor

- **providers**: make all creating log to DEBUG
- **clublog**: make API_KEY private
- **providers**: update GetName format

## v1.0.0 (2025-05-24)

### Feat

- **s3**: fix aws-chunked encoding not support
- support download remove file if larger
- define general interface for provider
- **s3**: init
- **main**: only monitor one adif file
- project first version

### Fix

- **sync**: replace the file lost the hard link
- **s3**: upload the whole file
