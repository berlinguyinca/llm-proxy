# Code Review Team - Multi-Agent Collaboration Skill

## Purpose
Enable teams of AI agents (5 developers + 1 security expert + 1 intern) to collaboratively review codebases, find problems, and reach consensus on solutions.

## Prerequisites
- Access to `.claude/skills/` directory
- Working with GSD (Get Shit Done) framework
- Project codebase ready for review

## Usage

### Activate Skill
```bash
/gsd-skills activate code-review-team
```

### Run Team Review
```bash
node ~/.config/opencode/get-shit-done/bin/gsd-tools.cjs code-review \
  --project . \
  --output /tmp/team-review.md \
  --format markdown
```

### Via Task Prompt
```python
Task(
  subagent_type="code-review-team",
  prompt="""
    <objective>
    Apply full multi-agent team review to codebase.
    Identify all problems, missing features, and nonsensical patterns.
    Reach consensus on solutions.
    </objective>

    <team_members>
      - architect_agent: Focus on complexity reduction and design patterns
      - performance_agent: Memory leaks, optimization opportunities
      - api_dev_agent: API design, data models, validation
      - ux_dev_agent: Documentation, error messages, user experience
      - devops_dev_agent: Deployment configs, monitoring, scalability
      - security_expert: Security vulnerabilities, auth flaws, best practices
      - intern_test_engineer: Edge cases, test coverage, documentation gaps
    </team_members>

    <workflow>
    1. Each agent independently scans codebase (30 min each)
    2. Lead architect synthesizes findings and prioritizes
    3. Team meets to discuss top 5 issues
    4. Reach consensus on implementation approaches
    5. Generate actionable report with Wave assignments
    </workflow>

    <expected_output>
    - Executive summary of all findings
    - Detailed issue list by severity
    - Team consensus decisions with reasoning
    - Implementation plan with Wave assignments
    - Code snippets for fixes (where applicable)
    </expected_output>
  """,
  model="code-review-team"
)
```

## Configuration Options

```yaml
skill:
  name: code-review-team
  config:
    team_size: 7                    # Total agents (5 devs + 1 security + 1 intern)
    parallel_execution: true        # All agents work simultaneously
    consensus_threshold: 2          # Min votes to approve solution
    priority_weights:
      security: 4                   # Security issues highest priority
      critical_bug: 3               # Crash-inducing bugs
      high_complexity: 2            # Hard to understand code
      nice_to_have: 1               # Documentation, optimization
    
    review_depth:
      surface: true                 # Syntax/style issues
      logical: true                 # Design/architecture issues  
      structural: true              # File organization problems
      security: true                # Vulnerabilities and best practices
      testability: true             # Missing edge cases
    
    communication_mode:
      async_first: true             # Agents work independently first
      consensus_meetings: true      # Scheduled sync for major issues
      voting_enabled: true          # Emoji-based voting system
    
    output_format:
      primary: markdown             # Report format
      secondary: json               # For programmatic use
      third_party_summary: true     # Executable summary section
```

## Agent Specializations

### 5 Developer Agents

1. **architect_agent** - System design patterns, complexity reduction
   - Scans for code smells and architectural debt
   - Identifies over-abstracted or under-engineered sections
   - Proposes refactoring strategies

2. **performance_agent** - Efficiency and resource usage
   - Memory leaks and allocation patterns
   - CPU hotspots and optimization opportunities
   - Resource contention issues

3. **api_dev_agent** - API design and business logic
   - Endpoint design consistency
   - Input validation completeness
   - Error handling patterns

4. **ux_dev_agent** - Developer experience and docs
   - Documentation clarity and completeness
   - Error message helpfulness
   - Setup instructions accuracy

5. **devops_dev_agent** - Operations and deployment
   - Configuration complexity
   - Health check completeness
   - Monitoring coverage gaps

### 1 Security Expert Agent
- Reviews for common vulnerabilities (SQLi, XSS, auth bypass)
- Checks authentication/authorization patterns
- Validates data exposure in responses/logs
- Recommends security best practices
- Assesses third-party dependency risks

### 1 Intern/Test Engineer Agent  
- Identifies edge cases that might crash code
- Creates test scenarios for unusual inputs
- Documents areas with high uncertainty
- Asks "what if?" questions to find gaps
- Ensures error handling is comprehensive

## Workflow in Detail

### Step 1: Independent Exploration (2-3 hours)
Each agent explores codebase:
```bash
# Agent tasks per file scanned
Agent reads:     src/**/*.go          # Core implementation
Agent reads:     cmd/**/*.go           # Entry points and CLI
Agent reviews:    pkg/**/*.go          # Package organization
Agent checks:     config/*             # Configuration files
Agent scans:      README.md            # Documentation quality
Agent validates:  .env.example         # Environment defaults
```

### Step 2: Issue Identification (30 min each)
Each agent documents findings in standardized format:
```markdown
## {Role} Finding: {ISSUE-TYPE-NN}

Severity: CRITICAL/HIGH/MEDIUM/LOW

Location: filepath:line_number OR "global pattern"

Problem: Clear description of the issue

Impact: What breaks if not fixed? Who is affected?

Current Behavior: What does the code do now?

Expected Behavior: How should it work?

Proposed Fix:
```go
// Before: problematic code
oldCode()

// After: fixed version  
fixedCode()
```

Trade-offs: Any downsides to this approach?

Alternatives considered:
- Option A (rejected): reason...
- Option B (backup): when to use instead
```

### Step 3: Lead Architect Synthesis (30 min)
Lead agent aggregates and prioritizes:
```markdown
## Executive Summary

Total Issues Found: {X}
- Critical: {Y} - Fix immediately
- High: {Z} - Address in Wave 0/1
- Medium: {W} - Can wait for Wave 2
- Low: {V} - Nice to have, prioritize if time permits

Missing Features ({N}):
- Feature A (priority: critical)
- Feature B (priority: high)  
- ...

Nonsensical Patterns ({M}):
- Anti-pattern X should be replaced with Y
- Pattern Z is over-complicating simple requirement

Top 5 Priority Issues (combined votes):
1. Hardcoded credentials (Security Expert, DevOps Agent - 3 votes)
2. Memory leak in memory manager (Performance Agent - 2 votes)  
3. Incomplete error docs (UX Agent - 2 votes)
4. Rate limit bypass possibility (API Agent - 2 votes)
5. Over-abstracted config loader (Architect Agent - 2 votes)
```

### Step 4: Consensus Meetings (as needed for critical issues)
For issues requiring team discussion:

```markdown
## Discussion: Issue-{NUM} - {TITLE}

Context: 
- Architect: "I see this, but..."
- Security: "From a security standpoint, we MUST consider..."
- UX: "End users would appreciate if..."
- Intern: "What happens if someone passes null here?"
- DevOps: "When deploying to production, this creates..."

Voting Results:
- ✅ Approve: [agents who agree]
- 🤔 Concerns: [agents with questions]  
- ❌ Reject: [agents opposed and why]

Final Decision (Consensus): {decision}
```

### Step 5: Implementation Plan
Lead architect creates wave-based plan:
```markdown
## Wave 0 - This Sprint (Critical Fixes)
- Issue: Hardcoded API keys
  Responsible: Security Expert, DevOps Agent
  Effort: 1 hour (replace with env var default + validation)
  Risk: Low, immediate security benefit

- Issue: Null pointer crash  
  Responsible: Intern Agent, API Agent
  Effort: 30 min defensive coding pattern
  Risk: Eliminate production crash

## Wave 1 - Next Sprint (High Priority)  
- Rate limiting bypass fix
- Memory leak resolution
- Documentation improvements

## Wave 2 - Backlog (Medium Priority)
- Config loader simplification
- Additional test coverage
```

### Step 6: Final Report Generation
Combine all artifacts into comprehensive report:
1. Executive summary
2. Detailed findings by agent
3. Consensus decisions with votes
4. Implementation plan by waves
5. Code snippets for common fixes
6. Checklist for prevention (avoid same issues in future)

## Team Decision-Making Patterns

### Consensus Building
```
Issue: "Should we add more validation or trust LM Studio's schema?"

Security Expert: "Trust is dangerous - validate EVERYTHING"
API Agent: "Adding validation adds latency, but prevents corruption"  
UX Agent: "Better error messages from validation vs cryptic 400s"
Intern: "What if someone manually edits config.yaml?"
DevOps: "If we're deploying to multiple machines, schema drift is bad"

Consensus decision: 
- Trust LM Studio defaults (don't override)
- But validate required fields + types before load
- Return user-friendly errors for validation failures
```

### Voting System
Agents use emoji voting during discussions:
- ✅ = Agree with approach  
- 🤔 = Have concerns but not blocking
- ❌ = Strongly disagree with explanation
- ⚠️  = Potential issue, needs attention

Decision passes when:
- Security-critical: unanimous agreement needed
- Major refactoring: ≥4 of 5 devs agree
- Nice-to-have: any positive votes OK

## Output Files Generated

### Primary Report: `team-review-report.md`
Comprehensive markdown document containing:
- Executive summary with priority matrix  
- All findings categorized by agent and severity
- Consensus decisions for each issue
- Implementation plan with Wave assignments
- Code fix examples for common patterns
- Checklist to prevent recurrence

### Secondary Files (for automation):
- `findings.json` - Machine-readable issue data
- `consensus-decisions.md` - Just the team decisions section  
- `implementation-plan.md` - Actionable tasks only

## Integration with GSD Workflow

This skill integrates at multiple points:

### 1. During Planning Phase
```bash
# After creating phase plan, run team review
/gsd-skills apply code-review-team --plan .planning/phases/01-foundation/01-01-WAVE-1-PLAN.md
```
Ensures plan addresses issues from previous reviews.

### 2. During Execution Phase  
```bash
# After each wave completes, verify quality
node ~/.config/opencode/get-shit-done/bin/gsd-tools.cjs review \
  --wave-complete .planning/phases/01-foundation/01-05-WAVE-1-SUMMARY.md
```
Catches issues before moving to next wave.

### 3. Before Major Commits
```bash
# Run lightweight check on proposed changes
git diff HEAD~1 | node ~/.config/opencode/get-shit-done/bin/gsd-tools.cjs code-review-diff
```

### 4. End-of-Sprint Retrospective  
Run full team review:
```bash
node ~/.config/opencode/get-shit-done/bin/gsd-tools.cjs code-review \
  --sprint .planning/phases/01-foundation \
  --report /tmp/sprint-retro-team-review.md
```

## Skill Metrics

After each run, generates metrics:
```yaml
team_review_metrics:
  total_issues_found: 23
  severity_breakdown:
    critical: 2        # Fix immediately
    high: 8            # Address this sprint
    medium: 9          # Backlog
    low: 4             # Nice to have
  
  findings_by_agent:
    architect: {count: 4, avg_severity: "medium"}
    performance: {count: 2, avg_severity: "high"}  
    api_dev: {count: 3, avg_severity: "medium"}
    ux_dev: {count: 2, avg_severity: "low"}
    devops: {count: 1, avg_severity: "critical"}
    security: {count: 4, avg_severity: "high"}
    intern: {count: 5, avg_severity: "medium"}
  
  consensus_efficiency:
    issues_resolved_in_first_meeting: 70%
    avg_time_to_consensus_minutes: 25
  
  false_positive_rate: 8%
```

## Best Practices

### When to Use This Skill

✅ **DO use:**
- After completing a major feature or phase
- Before significant architectural changes  
- During post-incident code investigation
- End-of-sprint retrospectives
- When reviewing third-party contributions
- Codebase is approaching 5k+ LOC

❌ **DON'T overuse:**
- For 5-line bug fixes (single agent review OK)
- When in time-critical hotfix situation  
- During active development sprints (schedule for sprint end)
- For configuration-only changes

### Team Composition Guidelines

- Keep team size fixed at 7 agents for consistency
- Security expert should NEVER be junior/intern - must have senior expertise
- Intern/test engineer role is intentionally separate from developers
- If adding specialists (e.g., ML agent), they integrate into existing roles

## Example Session

```bash
# Activate skill
/gsd-skills activate code-review-team

# Run review on current phase
node ~/.config/opencode/get-shit-done/bin/gsd-tools.cjs code-review \
  --phase 01-foundation \
  --output /tmp/team-review-01f.md \
  --include-voting-log

# Review completes, generates:
/tmp/team-review-01f.md              # Full markdown report
/tmp/team-review-01f-findings.json   # Structured data
/tmp/team-review-01f-consensus.md    # Team decisions only
```

### Sample Output (Excerpt)

```markdown
# Code Review Team Report - Phase 01 Foundation

## Executive Summary

**Total Issues Found: 23**  
✅ Critical: 2, High: 8, Medium: 9, Low: 4  
⏱️ Consensus meetings held: 3 (for critical issues only)  

### Top Priority Issues

1. **CRITICAL - Security**: Hardcoded API key in `cmd/management/main.go:245`
   Votes: Security Expert, DevOps Agent, UX Agent ✅
   
2. **HIGH - Performance**: Memory leak in memory manager when evicting models
   Votes: Performance Agent, Architect Agent, Security Expert ✅

3. **MEDIUM - Complexity**: Over-abstracted config loader with unnecessary hooks
   Votes: Architect Agent, UX Agent (dev experience), API Agent ⚠️

## Team Consensus Decisions

### Decision 1: Remove Hardcoded Keys (Wave 0)
**Why:** Security risk if repo is public or committed to private repo by mistake  
**Approach:** Use `.env` file with secure validation and default values  
**Implementation Time:** ~30 minutes  
**Risk:** Minimal - just adds environment variable handling

### Decision 2: Fix Memory Leak in Eviction (Wave 1)
**Why:** Can cause OOM under heavy load, affects production stability  
**Approach:** Clear old model's references before removing from registry  
**Implementation Time:** ~45 minutes  
**Risk:** None - defensive coding pattern

```

## Troubleshooting

### "Agent failed to complete review"
- Ensure codebase is accessible and not locked in other operations
- Check file permissions for read access
- Increase `time_allowed_minutes` config if needed

### "Consensus not reached on critical issue"  
- Lead Architect facilitates discussion with all perspectives represented
- If after 2 rounds no consensus: escalate to human reviewer
- Document rationale for final decision with full debate summary

### "Review taking too long"
- For <5k LOC: reduce time allocation per agent to 15 min
- Use `--skip-surface-level` flag to skip style issues  
- Focus only on high-priority areas: security, bugs, critical paths

## Related Skills

- **Code Quality Audit** - Similar team structure but different focus (linting, formatting, best practices)
- **Security Assessment** - Deep security-focused review with penetration testing agents
- **Performance Tuning Team** - Specialized for profiling and optimization workloads

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-03-15 | Initial release, 7-agent team structure |
| 1.1 | TBD | Add ML model review specialist role |

## See Also

- [GSD Framework Documentation](https://github.com/get-shit-done/framework)
- [Multi-Agent Collaboration Guide](/cli/agents.md#collaboration-workflows)
- [Code Review Best Practices](docs/CODE-REVIEW-GUIDELINES.md)