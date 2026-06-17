// Source: commitlint.js.org getting-started — CITED
// ESM via .mjs: package.json has no "type": "module", so a bare .js may resolve as
// CJS and break `export default` under Node 22 (RESEARCH Open Question 2). The .mjs
// extension forces ESM unconditionally.
//
// config-conventional's type-enum already accepts the repo's GSD commit style
// (docs(04): ..., feat(04): ..., chore(04): ... — knowledge/git.md), so no override
// is needed. Wired into the commit-msg hook by plan 04-03 (npx --no-install commitlint).
export default {
  extends: ['@commitlint/config-conventional'],
};
