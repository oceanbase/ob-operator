#!/bin/bash
set -e

mkdir -p /home/admin/data-file/wallet
ln -sf /home/admin/data-file/wallet /home/admin/oceanbase/wallet

exec /home/admin/oceanbase/bin/oceanbase-helper.real "$@"
