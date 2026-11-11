#!/usr/bin/env python3
"""Supply DIMA workflow context without executing any work-item actions."""

import json
from pathlib import Path
import re
import sys


def main():
    try:
        payload = json.load(sys.stdin)
    except (ValueError, UnicodeError):
        return
    if not isinstance(payload, dict):
        return
    prompt = payload.get("prompt")
    if not isinstance(prompt, str):
        return

    has_reference = re.search(
        r"(?<![a-z0-9])dima(?![a-z0-9])|openworkitemid|project\.alipay\.com",
        prompt,
        re.IGNORECASE,
    )
    bare_id = re.fullmatch(r"(?:[0-9]{19,20}|WI-[0-9]+)", prompt.strip(), re.IGNORECASE)
    if not (has_reference or bare_id):
        return

    skill = Path(__file__).resolve().parents[1] / "SKILL.md"
    print(
        f"[ob-operator DIMA] 先读取 {skill}，按本次用户范围执行。"
        "修复流程：工单与关联文档 → 本仓库回归用例 RED → 最小修复 GREEN → "
        "Go/Kubernetes 回归 → 三个独立 subagent 审查 → "
        "已授权的 PR/MR 提交、最新 HEAD 三方复审与 DIMA 回填读回。"
        "目标分支以工单组件/版本和实际 remote 为准。"
        "仅分析、不提交、不回填等限制优先；流程配置、状态询问和背景链接不启动修复或外部写入。"
    )


if __name__ == "__main__":
    main()
