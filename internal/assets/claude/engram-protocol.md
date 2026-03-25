## Engram Persistent Memory — Protocol

You have access to Engram, a persistent memory system that survives across sessions and compactions.
This protocol is MANDATORY and ALWAYS ACTIVE.

### PROACTIVE SAVE TRIGGERS (mandatory — do NOT wait for user to ask)

Call `mem_save` IMMEDIATELY after any of these:
- Architecture or design decision made
- Team convention documented or established
- Workflow change agreed upon
- Tool or library choice made with tradeoffs
- Bug fix completed (include root cause)
- Feature implemented with non-obvious approach
- Notion/Jira/GitHub artifact created or updated with significant content
- Configuration change or environment setup done
- Non-obvious discovery about the codebase
- Gotcha, edge case, or unexpected behavior found
- Pattern established (naming, structure, convention)
- User preference or constraint learned

Self-check after EVERY task: "Did I make a decision, fix a bug, learn something non-obvious, or establish a convention? If yes, call mem_save NOW."

Format for `mem_save`:
- **title**: Verb + what — short, searchable (e.g. "Fixed N+1 query in UserList")
- **type**: bugfix | decision | architecture | discovery | pattern | config | preference
- **scope**: `project` (default) | `personal`
- **topic_key** (recommended for evolving topics): stable key like `architecture/auth-model`
- **content**: **What** / **Why** / **Where** / **Learned** (omit Learned if none)

Topic update rules:
- Different topics MUST NOT overwrite each other
- Same topic evolving → use same `topic_key` (upsert)
- Unsure about key → call `mem_suggest_topic_key` first
- Know exact ID to fix → use `mem_update`

### WHEN TO SEARCH MEMORY

On any variation of "remember", "recall", "what did we do", "recordar", "acordate", "qué hicimos":
1. Call `mem_context` — checks recent session history (fast, cheap)
2. If not found, call `mem_search` with relevant keywords
3. If found, use `mem_get_observation` for full untruncated content

Also search PROACTIVELY when:
- Starting work on something that might have been done before
- User mentions a topic you have no context on
- User's FIRST message references the project, a feature, or a problem

### SESSION CLOSE PROTOCOL (mandatory)

Before ending a session or saying "done" / "listo" / "that's it", call `mem_session_summary`:

## Goal
[What we were working on this session]

## Instructions
[User preferences or constraints discovered — skip if none]

## Discoveries
- [Technical findings, gotchas, non-obvious learnings]

## Accomplished
- [Completed items with key details]

## Next Steps
- [What remains to be done — for the next session]

## Relevant Files
- path/to/file — [what it does or what changed]

### AFTER COMPACTION

If you see a compaction message or "FIRST ACTION REQUIRED":
1. IMMEDIATELY call `mem_session_summary` with the compacted summary content
2. Call `mem_context` to recover additional context from previous sessions
3. Only THEN continue working
