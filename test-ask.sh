#!/bin/bash

# Clear the screen for a clean prompt
clear

echo "================================================="
echo "  🧠 AI IDE Task Planner (Zero-Code First)"
echo "================================================="
echo "Let's construct your prompt to save tokens."
echo ""

# ==============================================================================
# 1. AGENT SELECTION
# ==============================================================================
echo "1. Select the AI agent persona (expertise level):"
echo "   1) Senior Staff Engineer (Generalist)"
echo "   2) Frontend Architect (React/Vue/Angular)"
echo "   3) Backend Engineer (Node/Go/Python)"
echo "   4) DevOps / SRE (Infrastructure, CI/CD, Cloud)"
echo "   5) Security Auditor (Zero-trust, OWASP, Compliance)"
echo "   6) Database Specialist (SQL/NoSQL, Optimization)"
echo "   7) Custom (specify your own)"
read -p "> " agent_choice
echo ""

# Set agent persona based on choice
case $agent_choice in
    1|"") agent_role="Senior Staff Engineer"
         agent_desc="You are a seasoned Senior Staff Engineer with broad expertise across the full stack, infrastructure, and system design. You prioritize clean, maintainable, and scalable solutions." ;;
    2) agent_role="Frontend Architect"
         agent_desc="You are a Senior Frontend Architect specializing in modern JavaScript frameworks (React, Vue, Angular). You prioritize performance, accessibility, UX, and component reusability." ;;
    3) agent_role="Backend Engineer"
         agent_desc="You are a Senior Backend Engineer with deep expertise in Node.js, Go, and Python. You focus on API design, data modeling, performance, and security." ;;
    4) agent_role="DevOps / SRE"
         agent_desc="You are a Senior DevOps Engineer and SRE. You specialize in CI/CD pipelines, container orchestration (Kubernetes/Docker), cloud infrastructure, monitoring, and reliability." ;;
    5) agent_role="Security Auditor"
         agent_desc="You are a Senior Security Auditor and AppSec Engineer. You focus on zero-trust architecture, OWASP Top 10, compliance (GDPR/SOC2), threat modeling, and secure coding practices. You must always include a threat model in the planning phase." ;;
    6) agent_role="Database Specialist"
         agent_desc="You are a Senior Database Specialist with expertise in SQL (PostgreSQL/MySQL) and NoSQL (MongoDB/DynamoDB). You focus on schema design, query optimization, indexing strategies, and data migrations." ;;
    7) echo "Enter custom agent role (e.g., 'Mobile Lead'):"
         read -p "> " custom_role
         agent_role="$custom_role"
         agent_desc="You are a Senior $custom_role. Follow the rules of engagement and provide expert-level guidance in your domain." ;;
    *) agent_role="Senior Staff Engineer"
         agent_desc="You are a seasoned Senior Staff Engineer with broad expertise across the full stack, infrastructure, and system design. You prioritize clean, maintainable, and scalable solutions." ;;
esac

# ==============================================================================
# 2. CORE TASK
# ==============================================================================
echo "2. What is the core objective?"
echo "   (e.g., 'Refactor the auth middleware to support JWT')"
read -p "> " task_desc

# Validate input
while [[ -z "$task_desc" ]]; do
    echo "⚠️  Task description cannot be empty. Please enter the core objective:"
    read -p "> " task_desc
done
echo ""

# ==============================================================================
# 3. CONTEXT FILES
# ==============================================================================
echo "3. Which files or directories should the AI look at?"
echo "   (e.g., 'src/middleware.js, components/Navbar.tsx' or 'src/components/' for all files in that dir)"
read -p "> " context_files

# Validate input
while [[ -z "$context_files" ]]; do
    echo "⚠️  Context files cannot be empty. Please specify at least one file or directory:"
    read -p "> " context_files
done
echo ""

# ==============================================================================
# 4. CONSTRAINTS / NOTES
# ==============================================================================
echo "4. Are there any specific constraints, libraries, or edge cases?"
echo "   (e.g., 'Use Tailwind only', 'Don't change the database schema', or press Enter to skip)"
echo "   For multi-line input, type your constraints line by line. Type 'END' on a new line to finish."
echo ""
constraints=""
while IFS= read -r line; do
    [[ "$line" == "END" ]] && break
    [[ -n "$constraints" ]] && constraints+=$'\n'
    constraints+="$line"
done
# If empty, use default message
if [[ -z "$constraints" ]]; then
    constraints="No specific constraints provided. Rely on standard best practices."
fi
echo ""

# ==============================================================================
# 5. OPTIONAL METADATA
# ==============================================================================
echo "5. Optional metadata (press Enter to skip each):"
echo "   Priority Level (e.g., 'High', 'Medium', 'Low'):"
read -p "> " priority
echo "   Expected Outcome (what success looks like):"
read -p "> " expected_outcome
echo "   Dependencies (external systems, APIs, teams):"
read -p "> " dependencies
echo ""

# ==============================================================================
# 6. GENERATE PROMPT
# ==============================================================================
timestamp=$(date +"%Y%m%d_%H%M%S")
date_dir=$(date +"%Y-%m-%d")
target_dir="local/ask-gpt/${date_dir}"
filename="${target_dir}/ask-gpt-${timestamp}.txt"

# Ensure the target directory exists
mkdir -p "$target_dir"

# Build the prompt template
cat > "$filename" <<EOF
# TASK: 
$task_desc

# CONTEXT FILES:
$context_files

# CONSTRAINTS / NOTES:
$constraints

# METADATA:
- Agent: $agent_role
- Priority: ${priority:-"Not specified"}
- Expected Outcome: ${expected_outcome:-"Not specified"}
- Dependencies: ${dependencies:-"None specified"}

---

# RULES OF ENGAGEMENT

$agent_desc

We are currently in **PHASE 1: PLANNING**. 

1. **Zero-Code Policy:** Do not write any implementation code, functions, or CSS in this response. 
2. **Context Check:** Briefly analyze the requested task against the provided context files. If the requirements are ambiguous, ask a maximum of 3 clarifying questions.
3. **Draft the Plan:** If you have enough context, outline a step-by-step architectural plan. 
4. **Format:** Output the plan as a Markdown checklist (\`- [ ]\`). Keep descriptions concise.

# REQUIRED ENDING
End your response exactly with this phrase: 
**"Plan ready. Type 'APPROVE' to authorize coding, or request changes."**
EOF

# ==============================================================================
# 7. PREVIEW AND CONFIRMATION
# ==============================================================================
echo "================================================="
echo "📝 Prompt Preview:"
echo "================================================="
echo ""
cat "$filename"
echo ""
echo "================================================="
echo ""
echo "Do you want to:"
echo "   1) Save and exit"
echo "   2) Edit the prompt (re-enter task)"
echo "   3) Discard and cancel"
read -p "> " confirm_choice

case $confirm_choice in
    2)
        echo "Re-enter the task description:"
        read -p "> " task_desc
        # Regenerate with new task description
        sed -i "s/^# TASK: .*/# TASK: \n$task_desc/" "$filename"
        echo "✅ Prompt updated!"
        ;;
    3)
        rm "$filename"
        echo "❌ Prompt discarded. Exiting."
        exit 0
        ;;
    *)
        echo "✅ Prompt saved."
        ;;
esac

# ==============================================================================
# 8. FINAL OUTPUT
# ==============================================================================
echo "================================================="
echo "✅ Done! Your AI prompt has been generated."
echo "📄 File: $filename"
echo ""
echo "You can now:"
echo "  - Copy the contents and paste into your AI IDE"
echo "  - Use it with ChatGPT, Claude, or any AI assistant"
echo "  - The prompt includes a tailored agent persona and clear phase 1 planning instructions"
echo ""
echo "📋 Summary:"
echo "  - Agent: $agent_role"
echo "  - Priority: ${priority:-'Not specified'}"
echo "  - Context: $context_files"
echo "================================================="