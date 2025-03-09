import requests
import argparse
import os
import sys
import json

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("token")
    parser.add_argument("tag")
    args = parser.parse_args()
    
    auth_token = args.token
    tag = os.path.basename(args.tag)
    release_name = f"Apollo v{tag}"
    request_headers = {
        "Accept": "application/vnd.github+json",
        "Authorization": f"Bearer {auth_token}",
        "X-GitHub-Api-Version": "2022-11-28"
    }
    body = {
        "tag_name": f"release/{tag}",
        "name": release_name,
        "draft": True,
        "prerelease": False,
        "generate_release_notes": False
    }
    res = requests.post(
        url="https://api.github.com/repos/Assifar-Karim/apollo/releases",
        headers=request_headers,
        data=json.dumps(body)
    )

    if not res.ok:
        print(res.text)
        sys.exit(1)
    json_res = res.json()
    release_id = json_res["id"]
    binaries = [
        "worker-linux-amd64.tar.gz",
        "worker-linux-arm64.tar.gz",
        "coordinator-linux-amd64.tar.gz",
        "coordinator-linux-arm64.tar.gz"
    ]
    request_headers["Content-Type"] = "application/octet-stream"
    for binary in binaries:
        path = os.path.join("bin", binary)
        bin_req = requests.post(
            url=f"https://uploads.github.com/repos/Assifar-Karim/apollo/releases/{release_id}/assets?name={binary}",
            headers=request_headers,
            data=open(path, "rb").read()
        )
        if not bin_req.ok:
            print(f"Could not upload {binary} to release!")
        
    
if __name__ == "__main__":
    main()