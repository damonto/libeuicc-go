#!/bin/bash

EXEC_DIR=$(pwd)

function copy_files() {
  find "$1" -type f \( -name "*.h" -o -name "*.c" \) | while read -r FILE; do
    BASENAME=$(basename "$FILE")
    TARGET_FILE="$EXEC_DIR/$BASENAME"
    cp -f "$FILE" "$TARGET_FILE"
    echo "Copied $FILE to $TARGET_FILE"

    sed -i 's/<cjson\/cJSON_ex\.h>/"cJSON_ex.h"/g' "$TARGET_FILE"
  done
}

copy_files "lpac/euicc"
copy_files "lpac/cjson"
