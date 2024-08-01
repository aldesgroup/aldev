package templates

const GitHookCOMMITMSG = `#!/bin/bash

# Path to the commit message file
COMMIT_MSG_FILE=$1

# Read the first line of the commit message
COMMIT_MSG=$(head -n 1 "$COMMIT_MSG_FILE")

# Define the allowed prefixes
PREFIXES=("dev:" "feat:" "fix:")

# Flag to indicate if the message is valid
IS_VALID=false

# Check if the commit message starts with any of the allowed prefixes
for PREFIX in "${PREFIXES[@]}"; do
    if [[ "$COMMIT_MSG" == "$PREFIX"* ]]; then
        IS_VALID=true
        break
    fi
done

# If the message is not valid, print an error and exit with status 1
if [ "$IS_VALID" = false ]; then
    echo "Error: Commit message must start with one of the following prefixes: ${PREFIXES[*]}"
    exit 1
fi`
