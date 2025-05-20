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

Create a `config.yaml` file in your running directory:

```yaml
source: /path/to/your/adif_file.adi

target:
- type: wavelog
  api_url: "https://<your.wavelog.domain>/index.php/api/qso"
  api_key: wlyourwavelogapikey
  station_profile_id: 3
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
