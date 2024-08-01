#!/usr/bin/bash
pushd /app
nav-db-filler /build/ksnav_kdb.json
nav -f /build/ksnav_nav.json > /out_img/run_init_process



