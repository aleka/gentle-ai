# SDD Phase — Common Protocol

This file contains boilerplate that is **identical** across all SDD phase skills (explore, propose, spec, design, tasks, apply, verify, archive). Sub-agents should load this alongside their phase-specific SKILL.md.

---

## Skill Registry

1. Check if the orchestrator provided a `SKILL: Load` instruction in your launch prompt. If yes, load that skill.
2. If no skill path was provided, search for the skill registry yourself:
   a. Try: `mem_search(query: "skill-registry", project: "{project}")` — if found, `mem_get_observation(id)` for full content
   b. Fallback: read `.atl/skill-registry.md` if it exists
3. From the registry, load any skills whose triggers match your current task.
4. If no registry exists, proceed with your phase skill only.

---

## Engram Upsert Note

When saving artifacts with `mem_save`, always set `topic_key` to the artifact's canonical key (e.g., `sdd/{change-name}/proposal`).

`topic_key` enables upserts — saving again updates, not duplicates.

---

## Return Envelope

Every phase MUST return a structured envelope to the orchestrator. Include ALL of these fields:

| Field | Description |
|-------|-------------|
| `status` | `success`, `partial`, or `blocked` |
| `executive_summary` | 1-3 sentence summary of what was done |
| `detailed_report` | (optional) Full phase output, or omit if already inline |
| `artifacts` | List of artifact keys/paths written |
| `next_recommended` | The next SDD phase to run, or "none" |
| `risks` | Risks discovered, or "None" |

Example:

```markdown
**Status**: success
**Summary**: Proposal created for `{change-name}`. Defined scope, approach, and rollback plan.
**Artifacts**: Engram `sdd/{change-name}/proposal` | `openspec/changes/{change-name}/proposal.md`
**Next**: sdd-spec or sdd-design
**Risks**: None
```
