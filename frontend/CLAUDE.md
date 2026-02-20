# Frontend

React + Vite frontend for the learn_kids educational platform. UI and content are in Russian.

## Commands

```bash
npm install          # Install dependencies
npm run dev          # Dev server at http://localhost:5173 (proxies /api to :8080)
npm run build        # TypeScript check + Vite production build
npm run lint         # ESLint
npm test             # Run Vitest tests
npm run test:watch   # Watch mode
```

After making changes: run `npm run lint` and `npm test`.

## Architecture

Feature-based structure under `src/features/`. Cross-cutting code lives in `src/shared/`.

- **Routing:** `src/App.tsx` — React Router v7
- **API client:** `src/shared/api.ts` — typed `fetch` wrapper
- **Shared types:** `src/shared/types.ts`
- **Features:** `auth/`, `lessons/`, `progress/`, `parent-dashboard/`, `theme/`, `i18n/`
- **Translations:** `src/i18n/` — Russian and English via i18next

## Key Conventions

- New pages go in a feature folder under `src/features/`
- Shared UI components go in `src/components/`
- All user-facing text must use i18next `t()` — add keys to both `ru` and `en` translation files
- Lesson content is rendered with `react-markdown` + `remark-gfm` + `rehype-highlight`
- Material types: `text`, `image`, `youtube`, `code`, `simulator`, `wokwi`
- API types must match `src/shared/types.ts` — update if the API contract changes
