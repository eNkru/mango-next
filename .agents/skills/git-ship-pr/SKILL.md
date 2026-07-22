---
name: git-ship-pr
description: "Ship work: ensure feature branch, confirm commit plan, commit, push, open GitHub PR. Use when user says ship, open PR, commit and push, create PR, or 'ok close the task and create PR'."
---

# Git Ship + Open PR

Turn finished local work into a remote branch and a GitHub pull request.

**Does not** merge the PR or archive Trellis tasks. After merge, use `git-merge-cleanup`. For Trellis archive/journal only, use `trellis-finish-work` (after code is committed).

## When to use

- User asks to commit + push + create PR
- User says "ok close the task in the same branch and create PR"
- End of Phase 3.4 when the plan includes shipping, not only local commit

## Preconditions

- User confirmed the commit plan (or explicitly said `ok` / `行` to a presented plan)
- Quality checks already run when this is a Trellis task (typecheck/tests as applicable)
- Do **not** commit secrets; do **not** amend/force-push unless user explicitly asks

## Step 1: Inspect git state

```bash
git status --porcelain
git branch --show-current
git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || true
git log --oneline -5
git diff --stat
```

Classify dirty paths:

| Kind | Action |
|------|--------|
| This task / this request | Include in commit plan |
| Unrelated WIP | Exclude; list as "not in commits" |
| Already clean | Skip commit; still may push + PR if branch is ahead |

## Step 2: Ensure feature branch

If current branch is `main` or `master` (or the default base) **and** there are commits or dirty files to ship:

```bash
git checkout -b <branch>
```

Branch naming (prefer project style from recent branches):

- `feat/<short-slug>`
- `fix/<short-slug>`
- `refactor/<short-slug>`
- `chore/<short-slug>`

If already on a feature branch, **keep it** (do not create a nested branch).

If the branch name is wrong for the change set, rename with user agreement:

```bash
git branch -m <new-name>
```

## Step 3: Commit plan (one-shot confirmation)

Learn message style from `git log --oneline -5`.

Group files into 1+ logical commits (not one commit per file). Present:

```text
Branch: <name> (create / already on)

Proposed commits (in order):
  1. <message>
     - <file>
  2. <message>
     - <file>

Excluded (not this work):
  - <file>

Reply 'ok' / '行' to execute. Edits or 'manual' to abort.
```

On rejection / `manual`: stop. Do not invent a second full plan unless the user asks.

If the user already said `ok` to a plan in the previous turn, execute without re-asking.

## Step 4: Commit

For each planned commit:

```bash
git add <files>
git commit -m "$(cat <<'EOF'
<message>

<optional body>
EOF
)"
```

Rules:

- No `--amend` unless user explicitly requests
- No empty commits
- Prefer HEREDOC for multi-line messages
- After Trellis archive/journal auto-commits, those are separate commits (see optional Step 6)

## Step 5: Push and create PR

```bash
git push -u origin HEAD
```

Then create PR against the base branch (usually `main`):

```bash
gh pr create --title "<title>" --body "$(cat <<'EOF'
## Summary
- <bullet>

## Test plan
- [ ] <check>
EOF
)"
```

- Title usually matches the primary work commit subject
- Body: summary + test plan (commands already run + manual checks)
- Return the PR URL to the user

If `gh` is unavailable, print the compare URL from `git push` remote output and stop.

## Step 6 (optional): Trellis close-out on same branch

Only when user asked to close the Trellis task **and** ship:

1. Code commits first (Steps 4–5 work commits may already be done)
2. `python3 ./.trellis/scripts/task.py archive <task-dir-or-name>`
3. `python3 ./.trellis/scripts/add_session.py --title "..." --commit "<work-hashes>" --summary "..."`  
   - Use **work** commit hashes only (not archive/journal hashes)
4. `git push` again so archive/journal commits land on the remote branch
5. If PR already exists, no need to recreate; push updates the PR

If user only wants PR and will archive later, skip this step.

## Step 7: Report

```text
Branch: <name>
Commits: <hashes + subjects>
PR: <url>
```

## Anti-patterns

- Committing on `main` without user saying so
- Shipping unrelated dirty files silently
- Force-push / amend by default
- Creating a PR with no push
- Archiving Trellis before code commits when the tree still has task code changes
