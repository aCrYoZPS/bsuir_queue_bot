cmd="/app/bin/main"
if [ "$REMOTE_DEBUG_PORT" ]; then
  echo "Starting application with remote debugging on port $REMOTE_DEBUG_PORT"
  dlvFlags="--listen=:$REMOTE_DEBUG_PORT --headless=true --log --api-version=2 --accept-multiclient"
  execFlags=""
  if [ "$REMOTE_DEBUG_PAUSE_ON_START" ]; then
    echo "Process execution will be paused until a debug session is attached"
  else
    execFlags="$execFlags --continue"
  fi
  cmd="/bin/dlv $dlvFlags exec $cmd $execFlags -- "
fi

echo "Executing command: $cmd $*"

exec $cmd "$@"