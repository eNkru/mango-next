# Task PRD: Global UI Context and State Store

## Goal

Consolidate fragmented `localStorage` usage and UI preferences (theme, language, reader settings) into a single React Context/Store to eliminate state synchronization drift across tabs and components.

## Target Area

- `frontend/src/context/` or `frontend/src/pages/reader/useReaderPrefs.ts`

## Requirements

1. Create a `UserPrefsContext` (or store) that manages theme settings, language/i18n, and reading preferences.
2. Synchronize store state updates with `localStorage` automatically.
3. Provide easy-to-use hooks (e.g. `useUserPrefs()`) for components to consume settings without manual listener setup.

## Acceptance Criteria

- [ ] All reader and application preference hooks consume unified state from `UserPrefsContext`.
- [ ] Theme/language/reader layout changes update instantly across all mounted components.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
