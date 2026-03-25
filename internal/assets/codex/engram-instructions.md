## Engram Persistent Memory — Protocol

You have access to Engram, a persistent memory system that survives across sessions and compactions.

### WHEN TO SAVE (mandatory — not optional)

Call mem_save IMMEDIATELY after any of these:
- Bug fix completed
- Architecture or design decision made
- Non-obvious discovery about the codebase
- Configuration change or environment setup
- Pattern established (naming, structure, convention)
- User preference or constraint learned

Format for mem_save:
- **title**: Verb + what — short, searchable (e.g. "Fixed N+1 query in UserList")
- **type**: bugfix | decision | architecture | discovery | pattern | config | preference
- **scope**: project (default) | personal
- **topic_key** (optional, recommended for evolving decisions): stable key like architecture/auth-model
- **content**: **What** / **Why** / **Where** / **Learned** (omit if none)

Topic update rules:
- Different topics must not overwrite each other
- Same evolving topic → reuse same topic_key (upsert)
- Unsure about key → call mem_suggest_topic_key first
- Know exact ID → use mem_update

### WHEN TO SEARCH MEMORY

On any variation of "remember", "recall", "what did we do", "recordar", "acordate", "qué hicimos":
1. Call mem_context — checks recent session history (fast, cheap)
2. If not found, call mem_search with relevant keywords
3. If found, use mem_get_observation for full untruncated content

Also search PROACTIVELY when:
- Starting work on something that might have been done before
- User mentions a topic you have no context on

### SESSION CLOSE PROTOCOL (mandatory)

Before ending a session or saying "done" / "listo" / "that's it", call mem_session_summary:

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

### PASSIVE CAPTURE — automatic learning extraction

When completing a task or subtask, include a "## Key Learnings:" section at the end of your response with numbered items. Engram will automatically extract and save these.

Example:
## Key Learnings:

1. bcrypt cost=12 is the right balance for our server performance
2. JWT refresh tokens need atomic rotation to prevent race conditions

You can also call mem_capture_passive(content) directly with any text containing a learning section.

### AFTER COMPACTION

If you see a compaction message or "FIRST ACTION REQUIRED":
1. IMMEDIATELY call mem_session_summary with the compacted summary content
2. Call mem_context to recover additional context from previous sessions
3. Only THEN continue working
