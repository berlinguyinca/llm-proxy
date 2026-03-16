# Code Review Team - Multi-Agent Collaboration Skill

## Overview

This skill enables a team of AI agents to collaboratively review codebases for problems, missing features, and nonsensical ideas. The system brings together:

- **5 Software Developers** - Each specializes in different aspects (architecture, performance, security, UX, testing)
- **1 Security Expert** - Reviews code for vulnerabilities, best practices, and compliance
- **1 Intern/Test Engineer** - Creates comprehensive test plans and identifies edge cases

All agents work together to find issues, propose solutions, and reach consensus on the best approach.

## Team Structure

```
┌─────────────────────────────────────────┐
│           Lead Architect Agent          │ ← Synthesizes findings, coordinates consensus
└─────────────────────────────────────────┘
                    ↑
        ┌───────────┼───────────┬──────────────┐
        ↓           ↓           ↓              ↓
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│  Arch Agent  │ │ Perf Agent   │ │ UX/Dev Agent │ │ Testing Agent│ ← Intern/Test Engineer
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
        ↑           ↑           ↑              ↑
        └───────────┴───────────┴──────────────┘
                  │
        ┌─────────┼─────────┐
        ↓         ↓         ↓
┌─────────────────┐ ┌─────────────────┐
│ Security Expert │ │  Intern Agent   │
│    (Senior)     │ │ (Junior, eager) │
└─────────────────┘ └─────────────────┘
```

## Roles and Responsibilities

### Lead Architect Agent
- **Role**: Team coordinator and consensus synthesizer
- **Focus**: Big picture issues, architectural decisions, overall impact assessment
- **Tools**: Cross-references all other findings, prioritizes fixes, documents trade-offs

### Developer Agents (5 specialized roles)
1. **Architect Dev** - System design, patterns, complexity reduction
2. **Performance Dev** - Efficiency, memory usage, optimization opportunities
3. **API/Backend Dev** - API design, data models, business logic issues
4. **Frontend/UX Dev** - User experience, documentation clarity, accessibility
5. **DevOps/SRE Dev** - Deployment configs, monitoring, scalability

### Security Expert Agent
- **Role**: Senior security reviewer (not junior intern)
- **Focus**: Vulnerabilities, auth flaws, data exposure, compliance
- **Tools**: Static analysis recommendations, threat modeling

### Intern/Test Engineer Agent
- **Role**: Junior developer with enthusiasm for quality
- **Focus**: Edge cases, test coverage, error handling, documentation gaps
- **Approach**: Asks "what if?" questions that catch oversights

## Workflow Process

### Phase 1: Individual Discovery (Parallel)
Each agent runs independently:
```
Agent explores codebase → identifies issues → documents findings
```

**Example output format:**
```markdown
## [Agent Role] Issue Report: {ISSUE-TYPE}

### Severity: Critical/High/Medium/Low
### Location: filepath:line_number
### Problem Description: ...
### Impact: ...
### Proposed Solution: ...
### Trade-offs: ...
```

### Phase 2: Initial Synthesis (Lead Architect)
Lead agent collects findings and creates summary:
```
"Team has identified X issues, Y missing features, Z nonsensical patterns"
Priority breakdown by severity and complexity"
```

### Phase 3: Collaborative Solutioning (All Agents)
For each major issue (>10% codebase impact or critical):
```
1. Lead Architect proposes solution framework
2. All agents discuss alternatives
3. Each agent votes (emoji: 👍 approve, 🤔 concerns, 👎 reject)
4. Refine best proposal with combined insights
```

### Phase 4: Consensus Decision
Lead agent documents final recommendation:
```markdown
## Recommendation: {DECISION}

### Why This Approach?
- Pros from each agent's perspective
- Mitigated risks
- Implementation cost/benefit

### Alternative Options Considered
- Option A (rejected): {reason}
- Option B (secondary): {when to use instead}
```

### Phase 5: Implementation Plan
Lead architect creates actionable plan with assignments.

## How to Use This Skill

### Via `/gsd-skills` CLI:
```bash
/gsd-skills list        # See available skills
/gsd-skills apply code-review-team --help    # Get usage details

# Activate the skill for current project
/gsd-skills activate code-review-team

# Now use in any analysis phase
Task(prompt="Apply code-review team analysis...", subagent_type="code-review-team")
```

### Via Prompt Context:
```yaml
skill: "code-review-team"
team_config:
  parallel_execution: true
  consensus_threshold: 2      # Minimum votes for agreement
  priority_levels: ["critical", "high", "medium", "low"]
```

### Direct Command Usage:
```bash
# Run the full team review on current directory
node ~/.config/opencode/get-shit-done/bin/gsd-tools.cjs code-review --dir . --output report.md
```

## Team Meeting Format

When agents meet to discuss issues, use this structure:

### Agent Introductions (Brief)
```
Arch: "Found 3 complexity issues in package selection"
Perf: "Discovered 2 memory leaks in pkg/memory/"
API: "API endpoints missing rate limiting on sensitive data"
UX: "Documentation unclear on error handling patterns"
DevOps: "Docker images could be reduced by 60%"
Sec: "⚠️ Hardcoded credentials in config files!"
Intern: "Found edge case where empty arrays crash the proxy"
```

### Consensus Discussion (Focused)
```
Arch: "So we have critical auth issues, optimization opportunities, and..."
Sec: "Let's address security first - hardcoded keys are blocker"
Perf: "The memory leaks could also cause denial of service"
Lead Architect: "Priority order: 1. Security, 2. Critical bugs, 3. Optimization"
```

### Final Decision Document
```markdown
## Team Consensus Decision

**Issue**: Hardcoded API keys in config files  
**Decision**: Remove all hardcoded keys, use environment variables with proper defaults  
**Support**: Sec (critical), DevOps (deployment security), UX (developer clarity)  
**Implementation Cost**: Low - just update env file loading  
**Timeline**: Can be done in Wave 1  
```

## Example Analysis Output

### Individual Findings Summary
```
┌─────────────────────────────┬──────┬─────────────────┐
│ Agent                       │ Issues│ Missing Features │
├─────────────────────────────┼───────┼──────────────────┤
│ Architect                   │   4   │     1            │
│ Performance                 │   2   │     0            │
│ API                         │   3   │     2            │
│ UX                          │   2   │     5            │
│ DevOps                      │   1   │     3            │
│ Security                    │   1*  │     0            │
│ Intern (Test)               │   5   │     8            │
└─────────────────────────────┴───────┴──────────────────┘

*Security: 1 critical, 2 medium severity
```

### Priority Matrix
```
CRITICAL (Fix in Wave 0):
- ❌ Hardcoded API keys (security)
- ⚠️ Memory leak causing OOM (perf + security risk)
- 🔥 Null pointer crash on empty arrays (intern finding)

HIGH (Wave 1, can be automated):
- 📝 Incomplete error handling docs (UX)
- 🐛 Rate limiting bypass vulnerability (API)
- 💾 Docker image bloat >500MB (DevOps)

MEDIUM (Wave 2, nice to have):
- 🔧 Missing metrics for custom endpoints (Perf)
- 📦 Over-abstracted config loader (Architect)
```

## Team Communication Patterns

### Consensus-Building Language:
- ✅ "I see your concern about X, and from my perspective..."
- ✅ "What if we combined your approach with Y?"
- ✅ "The security risk here is acceptable IF we implement Z"
- ⚠️ Avoid: "This is wrong", instead use: "This pattern has these risks..."

### Disagreement Resolution:
1. Lead Architect calls out different perspectives
2. Each agent states their priority rationale
3. Security > Functionality > Performance > Complexity (priority order)
4. Find common ground or escalate to human if critical disagreement

## Integration with Existing Workflows

This skill can be invoked at any phase:

```bash
# During planning:
/gsd-plan-phase --skill="code-review-team" "01-foundation"

# For continuous review:
Task(prompt="Team review new feature implementation", subagent_type="code-review-team")

# In execute phase for validation:
Task(prompt="Multi-agent code quality review before commit", subagent_type="code-review-team")
```

## Best Practices

### When to Use This Skill:
- ✅ New feature implementation (before committing)
- ✅ Architectural changes (>50% of codebase modified)
- ✅ Post-release incident investigation
- ✅ End-of-sprint retrospectives
- ✅ Code quality audits

### When NOT to Overuse:
- ❌ Minor bug fixes (<10 lines changed)
- ❌ Trivial config file updates
- ❌ Time-critical hotfixes (skip team review, single agent OK)

## Skill Activation Commands

```bash
# Activate for current session
/gsd-skills activate code-review-team

# Run analysis on specific directory
cd /path/to/codebase && /gsd-skills analyze --skill=code-review-team .

# Generate detailed report
node ~/.config/opencode/get-shit-done/bin/gsd-tools.cjs review-team \
  --output /path/to/report.md \
  --format markdown \
  --include-implementation-plan
```

## Example Skill Output Structure

```markdown
# Code Review Team Report - {DATE}

## Executive Summary
- Total issues found: {X}
- Critical: {Y}, High: {Z}, Medium: {W}
- Missing features identified: {N}
- Top 3 priorities: [...]

## Detailed Findings by Agent

### 🏗️ Architect Agent (4 findings)
#### Complexity Issue: Over-abstracted Config Loader
**Problem**: Config loader has unnecessary layers  
**Impact**: Confusing for new developers  
**Proposed**: Simplify to single-layer config with hooks  

---

### 🔒 Security Expert (3 findings)
#### Critical: Hardcoded API Keys
**Location**: `cmd/management/main.go:245`  
**Risk**: Full system compromise if committed  
**Fix**: Use .env file with secure defaults + validation

---

## Team Consensus & Recommendations

### Priority 1: Security Fixes (Wave 0 - This Sprint)
[Implementation plan from Lead Architect]

### Priority 2: Bug Fixes (Wave 1)
...

## Implementation Roadmap
```

## Metrics Tracked

- **Code Health Score**: Aggregated across all agent scores
- **Issue Detection Rate**: % of issues found by each agent
- **Consensus Efficiency**: Time to reach agreement on priorities
- **False Positive Rate**: Issues flagged but not real problems
- **Coverage**: % of codebase reviewed vs production critical paths

## Next Steps

1. Review team composition - Add more specialists?
2. Optimize workflow timing for different codebase sizes
3. Document common patterns from reviews
4. Create automated checks for recurring issues
5. Integrate into CI/CD pipeline