## v1.4.0 (2025-06-05)

### Feat

- **hamqth**: init provider
- **watcher**: retail the file when changed

## v1.3.2 (2025-05-31)

### Feat

- **main**: allow to choose whether download remote file

### Fix

- **adif**: use text/template

## v1.3.1 (2025-05-31)

### Fix

- **main**: fix header parser
- **hamcq**: status code should be 200

## v1.3.0 (2025-05-31)

### Feat

- **hamcq**: init
- **webhook**: support custom method, header and body
- show version on startup

## v1.2.2 (2025-05-30)

### Fix

- **consts**: should be var

## v1.2.1 (2025-05-29)

### Fix

- **goreleaser**: only release to Forgejo

## v1.2.0 (2025-05-29)

### Feat

- **webhook**: init
- **watcher**: use nxadm/tail lib
- **pkgs**: use retryablehttp lib
- **goreleaser**: init

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
