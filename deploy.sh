#!/bin/sh
# make sure you've configured this server in ~/.ssh/config
# XXX: $1 and $2 are NOT safe for untrusted input
ssh update-server 'docker exec -i $(docker ps -qf name=watchtower-db) psql -U postgres updates' <<EOF
    INSERT INTO version (image, version)
    VALUES ('$1', '$2')
    ON CONFLICT (image)
    DO UPDATE SET version = EXCLUDED.version;
EOF
