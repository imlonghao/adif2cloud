# adif2cloud

A tool for monitoring ADIF files and sending new QSO records to cloud services in real-time.

## Features

- Real-time monitoring of ADIF file changes
- Automatic transmission of new QSO records to cloud services
- Configuration file support
- Graceful shutdown handling

## Installation

```bash
go install git.esd.cc/imlonghao/adif2cloud@latest
```

## Configuration

Create a `config.yaml` file in your running directory with content similar to the following example:

```yaml
source: /path/to/your/adif_file.adi

target:
  - type: wavelog
    api_url: "https://your.wavelog.domain/index.php/api/qso"
    api_key: "wlyourwavelogapikey"
    station_profile_id: 1
  - type: s3
    endpoint: "https://s3.your-region.amazonaws.com" # e.g., s3.amazonaws.com or your MinIO endpoint
    region: "your-region" # e.g., us-east-1
    access_key_id: "YOUR_AWS_ACCESS_KEY_ID" # Your S3 access key ID
    secret_access_key: "YOUR_AWS_SECRET_ACCESS_KEY" # Your S3 secret access key
    bucket_name: "your-adif-backup-bucket" # Required: Name of your S3 bucket
    use_path_style: false # Optional: Set to true for MinIO or S3 compatible services requiring path-style addressing (defaults to false if omitted)
    file_name: "adif_file.adi" # Optional: Name of the file to upload to S3 (defaults to the source file name if omitted)
  - type: git
    repo_url: "https://github.com/your-username/your-repo.git"
    branch: "main"
    file_name: "adif_file.adi"
    commit_author: "Your Name <your.email@example.com>"
    commit_email: "your.email@example.com"
    auth_username: "your-github-username"
    auth_password: "your-github-password"
    auth_ssh_key: "/path/to/your-ssh-key"
    auth_ssh_key_passphrase: "your-ssh-key-passphrase"
  - type: clublog
    email: "your.email@example.com"
    password: "your-clublog-password"
    callsign: BA0AN
  - type: webhook
    url: "https://example.com/webhook"
    method: "GET"
    headers:
      X-Header: "value"
    body: |
      {"callsign": "{{.call}}"}
  - type: hamcq
    key: "YOUR_API_KEY"
```

## Usage

1. Ensure that the `config.yaml` file is correctly configured
2. Run the program:

```bash
adif2cloud
```

The program will start monitoring the specified ADIF files and automatically send new QSO records to the cloud service.

## Exit

Press `Ctrl+C` to gracefully exit the program. 
