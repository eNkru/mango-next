---
name: git-merge-cleanup
description: "Merge a GitHub PR (or confirm already merged), delete the feature branch, checkout main, and sync local. Use when user says merge the PR, PR merged clean up, switch back to main and delete the feat branch."
---

# Git Merge + Cleanup

After a feature PR is ready or already merged: merge if needed, delete the feature branch, return local repo to updated `main`.

**Does not** implement features or open PRs. For ship/PR creation use `git-ship-pr`. For Trellis archive/journal use `trellis-finish-work` (usually already done on the feature branch before merge).

## When to use

- User: "merged, switch back to main and delete the feat branch"
- User: "merge the PR, delete branch, back to main"
- User: "are you able to merge the pr and clean up"

## Step 1: Identify PR and branch

```bash
git branch --show-current
git status --porcelain
gh pr status
# or explicit:
gh pr view <number> --json number,state,mergeable,headRefName,baseRefName,url,title
```

If user said "merged" without a number, resolve the latest PR for the current branch or the branch they named:

```bash
gh pr list --head "<branch>" --state all --limit 1
```

If the working tree is dirty with unrelated changes, warn and ask before `reset --hard` / checkout.

## Step 2: Merge (if still open)

```bash
gh pr view <n> --json state,mergeable,title,url
```

| State | Action |
|-------|--------|
| `OPEN` + mergeable | `gh pr merge <n> --merge --delete-branch` (or `--squash` if user prefers) |
| `OPEN` + not mergeable | Report conflicts; stop |
| `MERGED` | Skip merge; still do local cleanup |
| `CLOSED` unmerged | Report; do not pretend success |

Default merge method: **merge commit** (`--merge`), matching this repo’s recent PR history. Use `--squash` only if user asks.

`--delete-branch` deletes the **remote** head branch when supported.

## Step 3: Local main sync

```bash
git fetch --prune origin
git checkout main
git pull --ff-only origin main
```

If pull fails due to ref lock / non-ff:

```bash
git fetch origin main
git checkout main
git reset --hard origin/main
```

Only use `reset --hard` when the user wants a clean match to remote and there is no precious local-only work on `main`.

## Step 4: Delete local feature branch

```bash
git branch -d <feature-branch> 2>/dev/null || git branch -D <feature-branch>
```

- Prefer `-d` (safe). Use `-D` only if branch is unmerged and user confirms discard.
- If branch already gone (`not found`), continue.

Remote delete if still present and not removed by `gh pr merge --delete-branch`:

```bash
git push origin --delete <feature-branch> 2>/dev/null || true
```

"remote ref does not exist" is OK.

## Step 5: Verify

```bash
git status
git branch --show-current
git log --oneline -5
git branch -a | grep <feature-slug> || echo "feature branch gone"
```

Expect:

- On `main`
- Clean working tree (unless user had other WIP)
- In sync with `origin/main`
- Feature branch absent locally (and preferably remotely)

## Step 6: Report

```text
PR: <url> → MERGED
Local: main @ <short-hash>
Branches removed: <name> (local/remote)
```

Optional: if a parent Trellis task is complete (all children archived), mention they can archive the parent with `task.py archive` — do not auto-archive parents unless user asks.

## Anti-patterns

- Merging without checking `mergeable` / CI when user did not waive checks
- `reset --hard` with uncommitted user work
- Deleting `main`
- Leaving local feature branch after "clean up" when delete was requested
- Creating a new Trellis task for merge cleanup (simple ops; no task unless user wants one)
