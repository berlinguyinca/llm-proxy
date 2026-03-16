# Code Review Team - Task Templates

## Standard Agent Output Format

Each agent should format their findings as follows:

```markdown
## {AGENT-ROLE} Finding: {ISSUE-ID}-{NUMBER}

### Severity
**CRITICAL** | **HIGH** | **MEDIUM** | **LOW**

### Location
`filepath.go:line_number`  OR  
"Global pattern across files"  OR  
"Entrypoint (main.go, entry points)"

### Problem Description
Clear, concise explanation of the issue. Focus on WHAT is wrong, not HOW it's implemented in code.

### Impact Analysis
- **Who/What affected:** Users, developers, production systems
- **When triggered:** Happy path, edge cases, error conditions  
- **Severity consequence:** Data loss, security breach, crashes, confusing UX

### Current Behavior
```go
// Code snippet showing current problematic pattern
```

Or for documentation issues:
> Current README says: "Build and run proxy with docker-compose"  
> Which is misleading because Docker is optional.

### Expected Behavior
Clear description of how it should work after fix.

### Proposed Fix

#### Approach 1 (Recommended)
```go
// Fixed code pattern
func improvedCode(args Args) error {
    // Better implementation
    return nil
}
```

**Pros:** 
- Clear explanation of benefits
- Addresses root cause

**Cons:**  
- Breaking changes if any
- Migration effort required

#### Approach 2 (Alternative - when Approach 1 not feasible)
```go
// Fallback implementation
func fallbackCode(args Args) error {
    // Safer but less elegant solution
    return nil
}
```

**Use when:** Quick fix needed, or primary approach has blockers

### Trade-offs Summary
- **Benefit:** [Primary advantage]
- **Cost:** What we give up to solve this
- **Risk:** Uncertainty of implementation
- **Maintenance impact:** Will future devs understand?

### Test Scenarios Created by Intern Agent
1. Happy path with normal inputs: should work ✅
2. Edge case: empty array input - should handle gracefully  
3. Invalid input: null/nil values - should return helpful error
4. Stress test: concurrent calls with high load - no crashes
5. Error recovery: after failure, system can restart cleanly

### Alternative Approaches Considered

| Option | Why Rejected | When to Use Instead |
|--------|--------------|---------------------|
| A (rejected) | [reason] | [scenario] |
| B (backup)  | [concerns] | If security risk becomes critical |
| C (nicer)   | [overkill] | For future enhancement, not fix |

### Agent Vote
- ✅ **Approve** - This is the best solution
- 🤔 **Concerns** - Need clarification before implementing
- ❌ **Reject** - This approach has showstoppers; use alternative instead

---

## TEAM MEETING TEMPLATE

For issues requiring group consensus, use this format:

### Meeting Record: Issue-{ISSUE-ID}

#### Lead Architect Opening
> "Team, let's address {issue_title}. Here's what we're dealing with..."

#### Agent Contributions
- **Architect:** "From a design perspective..."
- **Perf:** "Performance implications of each option..."
- **API:** "API contract changes required..."
- **UX:** "How this affects user experience..."
- **DevOps:** "Deployment complexity, rollback plan..."
- **Security Expert:** "Security impact assessment..."
- **Intern:** "Edge cases we might be missing..."

#### Discussion Points
1. [Point 1 from architect or security expert]
2. [Counterpoint or concern from another agent]
3. [Clarification or additional context from intern]
4. [Consensus building discussion...]

#### Voting Results
- ✅ Approve: [agent list]
- 🤔 Concerns: [agent list with specific questions]
- ❌ Reject: [agent list with reasons and alternatives]

#### Consensus Decision
**Decision:** [clear yes/no/with conditions]

**Rationale for this choice:**
1. Security priority (if applicable)
2. Minimal disruption to existing users  
3. Clear path to implementation
4. Backward compatibility maintained

**Implementation Details:**
- Files to modify: [...]
- Breaking changes: none / version bump required / migration guide needed
- Estimated effort: X hours (Wave 0/1/2 assignment)
- Test cases: See intern's scenarios above

---

## FINAL REPORT SECTION TEMPLATE

### Executive Summary for Stakeholders

**Issue:** {clear title in one line}  
**Severity:** CRITICAL/HIGH/MEDIUM/LOW  
**Impact:** Who/what is affected and by how much  
**Decision:** Approve / Reject / Needs More Discussion  
**Resolution Path:** [summary of implementation plan]  

**Why This Matters:** One or two sentences on business impact.

**What We're Doing Now:** Immediate actions being taken.

---

## Agent Role-Specific Templates

### Architect Agent
```markdown
## Architect Finding: {ID}

### Complexity Analysis
**Current Complexity Score:** [1-10, where 10 is worst]
**Root Cause:** Why this design decision was made and what changed

### Refactoring Proposal
Before (complex):
```go
// Old complicated code
type OverEngineered struct {
    // Too many abstractions
}
```

After (simple):
```go
// Direct and clear
type Simple struct {
    // Minimal necessary fields
}
```

### Pattern Applied
- ✅ DRY principle - eliminating duplication
- ✅ KISS principle - avoiding over-abstraction  
- ✅ YAGNI principle - no premature optimization
- ✅ SOLID principles - where applicable
```

### Security Expert Template
```markdown
## Security Finding: {SECU-{NUMBER}]

### CVE Reference (if known)
N/A or CVE-YYYY-XXXXX pattern

### Vulnerability Type
- [ ] SQL Injection
- [ ] Cross-Site Scripting (XSS)
- [ ] Authentication Bypass
- [ ] Insecure Deserialization  
- [ ] Path Traversal
- [ ] Command Injection
- [ ] Information Disclosure
- [ ] CSRF/Session Hijacking
- [ ] Hardcoded Secrets
- [ ] Other: {specify}

### Risk Assessment
**Exploitability:** Easy/Moderate/Difficult (requires attack vector)  
**Impact:** Data breach / Service disruption / Auth compromise  
**Likelihood:** Low/Medium/High based on common attack patterns  

### Attack Vector Description
How could someone exploit this? Be concrete.

### Mitigation Implementation
```go
// Before: vulnerable code
func handleRequest(req Request) Response {
    // Dangerous direct DB access
}

// After: fixed with parameterized queries and validation
func handleRequest(req Request) Response {
    validated, err := validateInput(req)  ← Always validate first
    if err != nil { return ErrorResponse(err) }
    sanitized, err := sanitizeOutput(result)  ← Never trust output
    // ... rest of implementation
}
```

### Testing Recommendations for Intern Agent
1. Fuzz test input parsing functions  
2. Test with SQL injection payloads in all text inputs
3. Check header injection attempts on content-type fields
4. Verify rate limiting works under attack simulation
5. Audit file access permissions if file operations involved
```

### Intern/Test Engineer Template
```markdown
## Testing Finding: {TEST-{NUMBER}]

### Edge Case Identified
**Scenario:** [clear description, e.g., "empty model name passed"]  
**Current Code Reaction:** [what happens now - nil panic? confusing error?]  
**Expected Behavior:** [what should happen instead]

### Test Case Created
```go
func TestEmptyModelName(t *testing.T) {
    // Test setup: create proxy with empty model name config
    
    // Execute: call LoadModel() endpoint
    
    // Verify: returns user-friendly error, not panic
    require.Error(t, result)
    assert.Contains(t, result.Error, "model_name cannot be blank")
}
```

### Boundary Conditions Tested
- ✅ Empty strings ("" length 0)
- ✅ Whitespace only ("   ")  
- ✅ Very long inputs (>1000 chars)
- ✅ Special characters in model paths (../)
- ✅ Unicode in configuration fields
- ✅ Null/nil pointers from missing dependencies

### Documentation Gaps Identified
Missing documentation for:
1. How to configure X without breaking Y
2. Error message meanings  
3. Which fields are required vs optional
4. What happens during upgrades/migrations
```

---

## Consensus Voting Emoji Legend

Use these when team discusses issues:

- ✅ **Green Check** = Strongly approve this solution
  - Agent understands and agrees with approach
  - Has considered trade-offs carefully
  
- 🤔 **Thinking Face** = Concerns raised but not blocking
  - Needs clarification or discussion  
  - "I'd prefer X, but Y works if we add Z"
  
- ❌ **Red X** = Reject this approach
  - Has fundamental blocker (security risk, impossible, etc.)
  - Better alternative exists
  
- ⚠️ **Warning Sign** = Potential issue needs attention
  - "This works but could cause problems later"
  - Add to checklist for future review

### Voting Thresholds

| Issue Type | Votes Needed | Who Decides |
|------------|--------------|-------------|
| Critical (security, crash) | All 7 must agree (unanimous) | Lead Architect or Security Expert overrides if critical |
| High priority | ≥5 of 7 votes approve | Lead Architect resolves deadlock |
| Medium priority | ≥3 of 5 dev agents agree | Can proceed with alternative approach |  
| Nice-to-have | Any positive vote OK | Optional, can wait for capacity |

---

## Report Checklist

Before finalizing team review report:

- [ ] All agent findings documented with severity and location
- [ ] Voting results captured for each consensus decision  
- [ ] Implementation plan includes Wave assignments (0/1/2)
- [ ] Code fix examples include before/after comparisons
- [ ] Test scenarios created by intern included
- [ ] Trade-offs clearly explained for major refactors
- [ ] Checklist added to prevent same issues in future reviews
- [ ] Metrics section completed with issue counts by severity and agent

---

## Summary

This template ensures:
1. **Consistent formatting** across all agent findings  
2. **Clear decision-making** through voting mechanisms
3. **Actionable output** with code examples and test cases
4. **Comprehensive coverage** from architect to intern perspectives  
5. **Measurable improvement** through severity tracking and metrics

Use these templates in all team review sessions for maximum effectiveness.