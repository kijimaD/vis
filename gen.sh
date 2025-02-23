#!/bin/bash
set -eux

######################
# ファイル一覧を生成する
######################

cd `dirname $0`

ls files | jq -R -s 'split("\n")[:-1] | { files: map({name: .}) }' > files.json
