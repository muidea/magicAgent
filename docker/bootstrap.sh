#!/bin/sh

EXTRA_ARGS=$EXTRA_ARGS
if [ "$LISTENPORT" ]; then
    EXTRA_ARGS=$EXTRA_ARGS' -ListenPort='$LISTENPORT
fi

if [ "$ENDPOINTNAME" ]; then
    EXTRA_ARGS=$EXTRA_ARGS' -EndpointName='$ENDPOINTNAME
fi

# shellcheck disable=SC2068
/var/app/magieAgent $EXTRA_ARGS $@
