#!/bin/bash

# Directory
DATA_DIR="test_data"

if [ -d "$DATA_DIR" ]; then
    echo "Test data directory '$DATA_DIR' already exists."
    exit 0
fi

echo "Scaffolding test data in '$DATA_DIR'..."
mkdir -p "$DATA_DIR/personal"
mkdir -p "$DATA_DIR/work/project_alpha"
mkdir -p "$DATA_DIR/archive"

# Welcome Note
cat > "$DATA_DIR/welcome.md" <<EOF
# Welcome to Mend

This is a scaffolded test environment.

## Features
- **File Tree**: Navigate using arrow keys.
- **Notes**: View rendered Markdown.
- **Hints**: Toggle hints with Space.

Try hinting: **Space out**
EOF

# Personal Note
cat > "$DATA_DIR/personal/todo.md" <<EOF
# Personal Todos

- [ ] Buy groceries
- [ ] Call mom
- [ ] Fix the **sink**
EOF

# Work Note
cat > "$DATA_DIR/work/project_alpha/specs.md" <<EOF
# Project Alpha Specs

## Overview
This is a **top secret** project.

## Requirements
1. Fast
2. Reliable
3. __Secure__
EOF

# Archive Note
cat > "$DATA_DIR/archive/old_ideas.md" <<EOF
# Old Ideas

Just some random thoughts from the past.
EOF

echo "Done."
