#!/usr/bin/env python3

import requests
import os

releases = requests.get(
    "https://git.esd.cc/api/v1/repos/imlonghao/adif2cloud/releases"
).json()
release = [x for x in releases if x["tag_name"] == os.environ["GITHUB_REF_NAME"]]

if len(release) == 0:
    print(f"Release {os.environ['GITHUB_REF_NAME']} not found")
    exit(1)

release = release[0]

github_release = requests.post(
    "https://api.github.com/repos/imlonghao/adif2cloud/releases",
    json={
        "tag_name": release["tag_name"],
        "name": release["name"],
        "body": release["body"],
        "target_commitish": release["target_commitish"],
    },
    headers={"Authorization": f'Bearer {os.environ["GITHUB_TOKEN"]}'},
).json()

for asset in os.listdir("dist"):
    if not asset.endswith((".tar.zst", ".zip")):
        continue
    requests.post(
        f"https://uploads.github.com/repos/imlonghao/adif2cloud/releases/{github_release['id']}/assets",
        params={
            "name": asset,
        },
        headers={
            "Authorization": f'Bearer {os.environ["GITHUB_TOKEN"]}',
            "Content-Type": "application/octet-stream",
        },
        data=open(f"dist/{asset}", "rb").read(),
    )
