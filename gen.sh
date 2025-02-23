#!/bin/bash
set -eux

######################
# ファイル一覧を生成する
######################

cd `dirname $0`

ls -1 | jq -R -s 'split("\n")[:-1] | { files: map({name: .}) }'
