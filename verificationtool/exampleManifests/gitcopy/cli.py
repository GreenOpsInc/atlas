import os
import click
import requests
import json
import base64
from dotenv import load_dotenv

__base_url__ = "https://api.github.com"


def set_header(token):
    headers = {
        'Authorization': f"Bearer {token}",
        'Content-Type': 'application/json'
    }
    return headers


def get_path_and_content(token, resp, tree_items):
    if resp.get("type") == "file" and resp.get("content") is not None:
        tree_item = {
            "path": resp.get("path"),
            "content": base64.b64decode(resp.get("content")).decode("ascii")
        }
        tree_items.append(tree_item)
        return

    resp = requests.get(resp.get("url"), headers=set_header(token))
    if resp.status_code != 200:
        raise Exception("ERROR while getting content: ", resp.json())

    resp = resp.json()
    if type(resp) == list or resp.get("tree"):
        for res in resp:
            get_path_and_content(token, res, tree_items)
    else:
        get_path_and_content(token, resp, tree_items)


def get_tree_items(token, owner, repo, src, dest):
    url = f"{__base_url__}/repos/{owner}/{repo}/contents/{src}"
    src_path_and_content = requests.get(url, headers=set_header(token))
    if src_path_and_content.status_code != 200:
        raise Exception("ERROR while getting repository tree: ",
                        src_path_and_content.json())

    tree_partial_item = []
    for each in src_path_and_content.json():
        get_path_and_content(token, each, tree_partial_item)

    tree_items = []
    for each in tree_partial_item:
        url = f"{__base_url__}/repos/{owner}/{repo}/git/blobs"
        body = {
            "content": each.get("content")
        }
        blob_resp = requests.post(
            url, headers=set_header(token), data=json.dumps(body))
        if blob_resp.status_code != 201:
            raise Exception("ERROR while uploading blob: ", blob_resp.json())

        path = each.get("path").replace(src, dest)
        tree_item = {
            "path": path,
            "sha": blob_resp.json().get("sha"),
            "mode": "100644",
            "type": "blob"
        }
        tree_items.append(tree_item)
    return tree_items


def update_main_branch(token, owner, repo, sha):
    url = f"{__base_url__}/repos/{owner}/{repo}/git/refs/heads/main"
    resp = requests.patch(url, headers=set_header(token), data=json.dumps({
        "sha": sha
    }))
    if resp.status_code != 200:
        raise Exception("ERROR while updating main branch: ", resp.json())
    print(resp.json())


def get_main_branch_sha(token, owner, repo):
    url = f"{__base_url__}/repos/{owner}/{repo}/git/refs"
    repo_resp = requests.get(url, headers=set_header(token))
    if repo_resp.status_code != 200:
        raise Exception(
            "ERROR while getting main branch sha: ", repo_resp.json())

    for each in repo_resp.json():
        if each.get("ref") == "refs/heads/main":
            return each.get('object').get('sha')


def create_tree(token, owner, repo, src, dest):
    tree_items = get_tree_items(token, owner, repo, src, dest)
    main_branch_sha = get_main_branch_sha(token, owner, repo)

    url = f"{__base_url__}/repos/{owner}/{repo}/git/trees"
    body = {
        "base_tree": main_branch_sha,
        "tree": tree_items,
    }
    resp = requests.post(url, headers=set_header(token), data=json.dumps(body))
    if resp.status_code != 201:
        raise Exception("ERROR while creating tree: ", resp.json())
    else:
        return resp.json().get("sha")


def create_commit(token, owner, repo, src, dest):
    tree_sha = create_tree(token, owner, repo, src, dest)
    main_branch_sha = get_main_branch_sha(token, owner, repo)

    url = f"{__base_url__}/repos/{owner}/{repo}/git/commits"
    body = {
        "message": f"moved items from {src} to {dest}",
        "tree": tree_sha,
        "parents": [main_branch_sha]
    }
    resp = requests.post(url, headers=set_header(token), data=json.dumps(body))
    if resp.status_code != 201:
        raise Exception("ERROR while creating commit: ", resp.json())
    else:
        update_main_branch(token, owner, repo, resp.json().get("sha"))


@click.group()
def main():
    """
    Simple CLI to modify Github files using just one command
    """
    load_dotenv()


@main.command()
@click.argument("owner")
@click.argument("repo")
@click.argument("src")
@click.argument("dest")
def copy(owner, repo, src, dest):
    """
        <owner> <REPO_NAME> <SOURCE_PATH> <DESTINATION_PATH>
    """
    token = os.getenv('GITHUB_ACCESS_TOKEN')
    create_commit(token, owner, repo, src, dest)
    print(f"SUCCESS, files copied from {src} to {dest}")


if __name__ == "__main__":
    main()
