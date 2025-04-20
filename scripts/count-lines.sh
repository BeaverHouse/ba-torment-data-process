#!/bin/bash

# Count to lines of all .go files in the current directory and subdirectories
# This is Bash script, see count-lines.ps1 for the PowerShell version.
# Written by Austin Lee.

total_lines=0
file_count=0

echo -e "\033[32mCalculating Go file line count...\033[0m"

# Find all .go files recursively and process them
while IFS= read -r -d '' file; do
    lines=$(wc -l < "$file")
    total_lines=$((total_lines + lines))
    file_count=$((file_count + 1))
    
    # Print information about each file
    echo -e "\033[33m$file: $lines lines\033[0m"
done < <(find . -name "*.go" -print0)

# Print final result
echo -e "\n\033[32mFinal result\033[0m"
echo -e "\033[36mTotal file count: $file_count\033[0m"
echo -e "\033[36mTotal line count: $total_lines\033[0m" 