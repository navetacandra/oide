#!/bin/sh
PROJECT_PATH="/app"
if [ -d "$PROJECT_PATH" ] && [ -f "$PROJECT_PATH/Procfile" ]; then
  cd $PROJECT_PATH
  local TYPE="$(echo "$(grep -m 1 '^TYPE=' $PROJECT_PATH/Procfile | cut -d '=' -f2-)" | tr -d '"')"
  local PREPARE="$(echo "$(grep -m 1 '^PREPARE=' $PROJECT_PATH/Procfile | cut -d '=' -f2-)" | tr -d '"')"
  local BUILD="$(echo "$(grep -m 1 '^BUILD=' $PROJECT_PATH/Procfile | cut -d '=' -f2-)" | tr -d '"')"
  local START="$(echo "$(grep -m 1 '^START=' $PROJECT_PATH/Procfile | cut -d '=' -f2-)" | tr -d '"')"
  if [ -n "${TYPE:-}" ] && [ "$TYPE" = "web" ]; then
    cp -rf "$PROJECT_PATH/*" /var/www/locahost/htdocs
    cd /var/www/locahost/htdocs
    httpd -D BACKGROUND
  fi
  if [ -n "${PREPARE:-}" ]; then
    echo "Run prepare script: '$PREPARE'"
    echo "Run prepare script: '$PREPARE'" >> /app/log
    eval "nohup $PREPARE 2>&1 > /app/log &"
    echo "" >> /app/log
    echo "" >> /app/log
  fi
  if [ -n "${BUILD:-}" ]; then
    echo "Run build script: '$BUILD'"
    echo "Run build script: '$BUILD'" >> /app/log
    eval "nohup $BUILD 2>&1 > /app/log &"
    echo "" >> /app/log
    echo "" >> /app/log
  fi
  if [ -n "${START:-}" ]; then
    echo "Runing with script: '$START'"
    echo "Runing with script: '$START'" >> /app/log
    eval "nohup $START 2>&1 > /app/log &"
  fi
fi
