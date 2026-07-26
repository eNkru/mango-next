# GitHub links and issue reporting UI

## Goal

Give users direct access to the project's GitHub repository and a GitHub issue-reporting page from the React UI.

## Confirmed Facts

- The current shared application chrome is `frontend/src/shell/AppShell.tsx`.
- Its top bar already contains navigation, theme, UI-style, language, and logout controls.
- The React frontend currently has no GitHub repository, issue, feedback, or report link.
- Earlier legacy UI tasks used `https://github.com/eNkru/mango-next` as the repository URL, but this new task must confirm the desired destinations and interaction before implementation.
- `AppShell` has separate theme, UI-style, and language `<select>` controls; the immersive reader has its own intentionally separate top bar and stays out of scope.
- The app uses `lucide-react`, the shared `Icon` wrapper, semantic entries in `shell/icons.ts`, localized strings in `lib/i18n.tsx`, and `.mango-btn--icon` for compact icon-only controls.

## Requirements

- Add discoverable UI access to the project repository and its GitHub issue list.
- Present the GitHub repository link as an icon-only control rather than visible text; provide a localized accessible name and tooltip/title.
- Present the issue-reporting action as an icon-only control; show localized feedback/reporting help in a tooltip on pointer hover and keyboard focus rather than a persistent visible label.
- Replace visible theme, UI-style, and language selects with compact icon buttons that reveal their choices in accessible dropdown menus.
- Each preference menu exposes the current value and lets the user select a value without cycling settings accidentally.
- Preserve the established React shell visual language and provide accessible, localized labels.
- GitHub icon opens `https://github.com/eNkru/mango-next`; issue icon opens `https://github.com/eNkru/mango-next/issues`, each in a new tab with safe external-link attributes.
- Do not change backend routes, API contracts, or application architecture.

## Scope

### In scope

- `frontend/src/shell/AppShell.tsx` compact tools, external links, and preference dropdown interactions.
- `frontend/src/shell/icons.ts`, `frontend/src/lib/i18n.tsx`, and `frontend/src/styles/shell.css` as necessary for the controls.
- AppShell desktop and mobile layout behavior across flat/comic and light/dark themes.

### Out of scope

- The reader top bar and reader controls.
- Backend routes, GitHub issue templates/forms, API changes, and npm dependencies.
- Replacing the existing preference storage keys or theme system.

## Acceptance Criteria

- [ ] Users can reach the intended GitHub repository from the application UI.
- [ ] Users can reach `https://github.com/eNkru/mango-next/issues` from the issue-report icon; its localized tooltip appears on hover and focus without persistent text.
- [ ] GitHub and issue links are icon-only, have localized accessible names, and use `target="_blank" rel="noopener noreferrer"`.
- [ ] Theme, UI-style, and language appear as icon buttons whose menus expose the current setting and select the existing stored value; keyboard users can open, navigate, select, and dismiss each menu without losing focus.
- [ ] The compact controls work on desktop and mobile layouts in comic/flat and light/dark variants.
- [ ] The implementation does not add npm dependencies or require backend changes.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
