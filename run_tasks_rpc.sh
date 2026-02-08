#!/bin/bash

# =============================================================================
# MEM - Semantic Memory CLI Implementation Runner
# =============================================================================
#
# This script automates the implementation of the mem CLI based on the
# specification and implementation plan.
#
# Related Files:
#   - spec.md      (Specification)
#   - plan.md      (Implementation Plan)
#   - prompt.md    (Task prompt for the agent)
#
# Usage:
#   ./run_tasks_rpc.sh
#
# =============================================================================

set -e

PLAN_FILE="plan.md"
SPEC_FILE="spec.md"
PROMPT_FILE="prompt.md"
NOTIFY_URL="http://mule.botnet:8081/message"
PROJECT_NAME="mem"

# Send start notification
echo "📡 Notifying start of implementation loop..."
curl -X POST "$NOTIFY_URL" \
  -H "Content-Type: application/json" \
  -d "{\"message\": \"ralph loop started for $PROJECT_NAME\"}" \
  --silent --show-error 2>/dev/null || true

if [ $? -eq 0 ]; then
    echo "✅ Start notification sent!"
else
    echo "⚠️  Failed to send start notification (endpoint may be unavailable)"
fi
echo ""

# Check if required files exist
if [ ! -f "$PLAN_FILE" ]; then
    echo "❌ Error: Plan file '$PLAN_FILE' not found!"
    exit 1
fi

if [ ! -f "$SPEC_FILE" ]; then
    echo "❌ Error: Spec file '$SPEC_FILE' not found!"
    exit 1
fi

if [ ! -f "$PROMPT_FILE" ]; then
    echo "❌ Error: Prompt file '$PROMPT_FILE' not found!"
    exit 1
fi

echo "=============================================================================="
echo "🚀 Ralph loop start"
echo "=============================================================================="
echo ""
echo "📋 Specification: $SPEC_FILE"
echo "📝 Plan File:     $PLAN_FILE"
echo "💬 Prompt File:   $PROMPT_FILE"
echo ""
echo "=============================================================================="
echo ""

# Count total tasks
TOTAL_TASKS=$(grep -c "\- \[ \]" "$PLAN_FILE" 2>/dev/null || echo "0")
COMPLETED_TASKS=$(grep -c "\- \[x\]" "$PLAN_FILE" 2>/dev/null || echo "0")

echo "📊 Progress: $COMPLETED_TASKS / $(($TOTAL_TASKS + $COMPLETED_TASKS)) tasks complete"
echo ""

if [ $TOTAL_TASKS -eq 0 ]; then
    echo "✅ All tasks in $PLAN_FILE are marked as complete!"
    echo ""

    # Send completion notification
    echo "📡 Sending completion notification..."
    curl -X POST "$NOTIFY_URL" \
      -H "Content-Type: application/json" \
      -d '{"message": "spec implementation complete"}' \
      --silent --show-error 2>/dev/null || true

    echo ""
    exit 0
fi

echo "⏳ Starting incremental implementation..."
echo ""

# Loop until no unchecked boxes remain in plan file
ITERATION=0
MAX_ITERATIONS=50  # Safety limit to prevent infinite loops

while grep -q "\- \[ \]" "$PLAN_FILE" && [ $ITERATION -lt $MAX_ITERATIONS ]; do
    ITERATION=$((ITERATION + 1))
    
    echo "------------------------------------------------------------------------------"
    echo "🔄 Iteration $ITERATION - $(date '+%Y-%m-%d %H:%M:%S')"
    echo "------------------------------------------------------------------------------"
    echo ""
    
    # Count tasks before this iteration
    TASKS_BEFORE=$(grep -c "\- \[ \]" "$PLAN_FILE")
    echo "📋 Tasks remaining: $TASKS_BEFORE"
    echo ""
    
    # Show next few unchecked tasks as context
    echo "🔍 Upcoming tasks:"
    grep "\- \[ \]" "$PLAN_FILE" | head -3 | sed 's/^/   /'
    echo ""
    
    # Run the agent with a 20-minute timeout using pi-wrapper for better completion
    echo "🚀 Starting agent session..."
    
    # Read the prompt content
    PROMPT_CONTENT=$(cat "$PROMPT_FILE")
    
    # Run pi-wrapper with the prompt content
    timeout 1200s pi-wrapper "$PROMPT_CONTENT" 2>&1 || EXIT_CODE=$?
    
    # Capture exit code if timeout didn't set it
    EXIT_CODE=${EXIT_CODE:-$?}
    
    echo ""
    echo "------------------------------------------------------------------------------"
    
    if [ $EXIT_CODE -eq 124 ]; then
        echo "⏱️  Agent session timed out after 20 minutes. Continuing to next task..."
    elif [ $EXIT_CODE -ne 0 ]; then
        echo "⚠️  Agent exited with code $EXIT_CODE. Continuing to next task..."
    else
        echo "✅ Agent session completed successfully."
    fi
    
    # Count tasks after this iteration
    TASKS_AFTER=$(grep -c "\- \[ \]" "$PLAN_FILE")
    TASKS_COMPLETED=$((TASKS_BEFORE - TASKS_AFTER))
    
    echo "📝 Tasks completed in this iteration: $TASKS_COMPLETED"
    echo ""
    
    # If no tasks were completed, wait longer before next iteration
    if [ $TASKS_COMPLETED -eq 0 ]; then
        echo "⚠️  No tasks completed. The agent may need clearer instructions."
        echo "   Check prompt.md for clarity."
        sleep 5
    else
        sleep 2
    fi
done

echo "=============================================================================="

if [ $ITERATION -ge $MAX_ITERATIONS ]; then
    echo "⚠️  Reached maximum iteration limit ($MAX_ITERATIONS)"
    echo "   Some tasks may still be incomplete."
else
    echo "🎉 SUCCESS: All tasks in $PLAN_FILE are marked as complete!"
fi

echo "=============================================================================="
echo ""
echo "⚠️  IMPORTANT: Before considering this implementation complete:"
echo ""
echo "   1. ✅ Review the spec: $SPEC_FILE"
echo "   2. ✅ Review the plan: $PLAN_FILE"
echo ""

# Send completion notification
echo "📡 Sending completion notification..."
curl -X POST "$NOTIFY_URL" \
  -H "Content-Type: application/json" \
  -d '{"message": "spec implementation complete"}' \
  --silent --show-error 2>/dev/null || true

echo ""
echo "=============================================================================="