"""Portable release gate. Canonical source: Docs deploy/verify_release.py."""
import argparse
import hashlib
import json
from pathlib import Path
import re


def validate_release(manifest, phase="package", root=None, baseline=None):
    if not baseline or baseline.get("schemaVersion") != 1 or not baseline.get("packages"):
        raise ValueError("approved frontend baseline is required")
    receipt = manifest.get("frontendDependencies", {})
    if receipt.get("schemaVersion") != 1 or any(
        not re.fullmatch(r"[a-f0-9]{64}", receipt.get(key, ""))
        for key in ("packageJsonSha256", "lockfileSha256")
    ):
        raise ValueError("complete frontend build dependency receipt is required")
    expected_baseline = baseline.get("baselineSha256")
    if not re.fullmatch(r"[a-f0-9]{64}", expected_baseline or "") or receipt.get("baselineSha256") != expected_baseline:
        raise ValueError("frontend baseline mismatch or missing build dependency receipt")
    actual_packages = receipt.get("packages", {})
    if "@yueli/ui" not in actual_packages:
        raise ValueError("frontend receipt must include the approved UI package")
    for name, actual in actual_packages.items():
        expected = baseline["packages"].get(name)
        if not expected or any(actual.get(key) != expected.get(key) for key in ("version", "sha256")):
            raise ValueError(f"frontend dependency differs from approved baseline: {name}")
    components = manifest.get("components", {})
    for role in ("api", "web"):
        if role not in components:
            raise ValueError(f"missing release component: {role}")
    snapshot = manifest.get("sourceSnapshot", {})
    snapshot_hash = snapshot.get("sha256", "")
    if not re.fullmatch(r"[a-f0-9]{64}", snapshot_hash):
        raise ValueError("source snapshot SHA256 is required")
    for role in ("api", "web"):
        component = components[role]
        if component.get("snapshotSha256") != snapshot_hash:
            raise ValueError(f"{role}: source snapshot mismatch")
        for key in ("sourceSha256", "artifactSha256"):
            if not re.fullmatch(r"[a-f0-9]{64}", component.get(key, "")):
                raise ValueError(f"{role}: {key} is required")
        if not component.get("image"):
            raise ValueError(f"{role}: image is required")
        action = component.get("action")
        if action == "reuse":
            original = component.get("original", {})
            if any(original.get(key) != component[key] for key in ("sourceSha256", "artifactSha256", "image")):
                raise ValueError(f"{role}: reuse provenance mismatch")
            if not original.get("imageId"):
                raise ValueError(f"{role}: reuse image ID is required")
        elif action == "deploy":
            artifact = component.get("artifact")
            if not artifact:
                raise ValueError(f"{role}: artifact is required")
            if root is not None:
                artifact_path = (root / artifact).resolve()
                if not artifact_path.is_relative_to(root.resolve()):
                    raise ValueError(f"{role}: artifact outside release directory")
                if hashlib.sha256(artifact_path.read_bytes()).hexdigest() != component["artifactSha256"]:
                    raise ValueError(f"{role}: artifact digest mismatch")
        else:
            raise ValueError(f"{role}: action must be deploy or reuse")
        if phase == "candidate":
            result = manifest.get("candidate", {}).get(role, {})
            if result.get("passed") is not True or not result.get("imageId"):
                raise ValueError(f"{role}: candidate validation is incomplete")
            if action == "reuse" and result["imageId"] != component["original"]["imageId"]:
                raise ValueError(f"{role}: running reuse image changed")
    return True


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("manifest", type=Path)
    parser.add_argument("--phase", choices=("package", "candidate"), default="package")
    parser.add_argument("--baseline", type=Path, required=True,
                        help="current approved baseline exported by Workspace; never use an old release's own snapshot")
    args = parser.parse_args()
    try:
        baseline = json.loads(args.baseline.read_text(encoding="utf-8"))
        validate_release(json.loads(args.manifest.read_text(encoding="utf-8")), args.phase, args.manifest.parent, baseline)
    except (ValueError, OSError) as error:
        parser.exit(1, f"Release rejected: {error}\n")
    print("Release gate passed: API and Web accounted for")


if __name__ == "__main__":
    main()
