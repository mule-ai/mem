#!/bin/bash

# =============================================================================
# mem - Semantic Memory CLI Implementation Runner
# =============================================================================
#
# This script automates the incremental implementation of the mem
# (Semantic Memory CLI) project.
#
# Related Files:
#   - spec.md    (Product Specification)
#   - plan.md    (Implementation Plan)
#
# Usage:
#   ./run_tasks.sh
#
# =============================================================================

PI_CMD="/data/jbutler/git/jbutlerdev/pi-mono/packages/coding-agent/binaries/linux-x64/pi"
PLAN_FILE="plan.md"
SPEC_FILE="spec.md"
NOTIFY_URL="http://mule.botnet:8081/message"
PROJECT_NAME="mem"

# Send start notification
echo "📡 Notifying start of implementation loop..."
curl -X POST "$NOTIFY_URL" \
  -H "Content-Type: application/json" \
  -d "{\"message\": \"pi loop started for $PROJECT_NAME\"}" \
  --silent --show-error

if [ $? -eq 0 ]; then
    echo "✅ Start notification sent!"
else
    echo "⚠️  Failed to send start notification (endpoint may be unavailable)"
fi
echo ""

# Check if required files exist
if [ ! -f "$PLAN_FILE" ]; then
    echo "❌ Error: Plan file '$PLAN_FILE' not found!"
    echo "   Expected location: ./plan.md"
    exit 1
fi

if [ ! -f "$SPEC_FILE" ]; then
    echo "❌ Error: Spec file '$SPEC_FILE' not found!"
    echo "   Expected location: ./spec.md"
    exit 1
fi

echo "=============================================================================="
echo "🧠 mem - Semantic Memory CLI Implementation"
echo "=============================================================================="
echo ""
echo "📋 Specification: $SPEC_FILE"
echo "📝 Plan File:     $PLAN_FILE"
echo ""
echo "🎯 Goal: Build persistent semantic memory CLI with vector embeddings"
echo ""
echo "=============================================================================="
echo ""

# Count total tasks
TOTAL_TASKS=$(grep -c "\- \[ \]" "$PLAN_FILE")
COMPLETED_TASKS=$(grep -c "\- \[x\]" "$PLAN_FILE")

echo "📊 Progress: $COMPLETED_TASKS / $(($TOTAL_TASKS + $COMPLETED_TASKS)) tasks complete"
echo ""

if [ $TOTAL_TASKS -eq 0 ]; then
    echo "✅ All tasks in $PLAN_FILE are marked as complete!"
    echo ""
    echo "⚠️  REMINDER: Make sure to verify the final phase:"
    echo "   'Phase 7: End-to-End Integration Test'"
    echo ""

    # Send completion notification
    echo "📡 Sending completion notification..."
    curl -X POST "$NOTIFY_URL" \
      -H "Content-Type: application/json" \
      -d '{"message": "mem implementation complete"}' \
      --silent --show-error

    if [ $? -eq 0 ]; then
        echo "✅ Notification sent successfully!"
    else
        echo "⚠️  Failed to send notification (endpoint may be unavailable)"
    fi

    echo ""
    exit 0
fi

echo "⏳ Starting incremental implementation..."
echo ""

# Loop until no unchecked boxes remain in plan file
ITERATION=0
while grep -q "\- \[ \]" "$PLAN_FILE"; do
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
    
    # Run the agent with a 20-minute timeout (in case tool calls hang)
    echo "🚀 Starting agent session..."
    timeout 1200s $PI_CMD -p --timeout 600000 @prompt.md
    EXIT_CODE=$?
    
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
    
    # Brief pause between iterations
    sleep 2
done

echo "=============================================================================="
echo "🎉 SUCCESS: All tasks in $PLAN_FILE are marked as complete!"
echo "=============================================================================="
echo ""
echo "⚠️  IMPORTANT: Before considering this implementation complete:"
echo ""
echo "   1. ✅ Review the spec: $SPEC_FILE"
echo "   2. ✅ Review the plan: $PLAN_FILE"
echo "   3. ❌ Complete Phase 7: End-to-End Integration Test"
echo ""
echo "   Build and test the CLI:"
echo "   cd /data/jbutler/git/jbutlerdev/mem"
echo "   go build -o mem ./cmd/mem"
echo "   ./mem store 'Test memory'"
echo "   ./mem query 'test'"
echo ""

# Send completion notification
echo "📡 Sending completion notification..."
curl -X POST "$NOTIFY_URL" \
  -H "Content-Type: application/json" \
  -d '{"message": "mem implementation complete"}' \
  --silent --show-error

if [ $? -eq 0 ]; then
    echo "✅ Notification sent successfully!"
else
    echo "⚠️  Failed to send notification (endpoint may be unavailable)"
fi

echo ""
echo "=============================================================================="
