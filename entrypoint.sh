
#!/usr/bin/env bash

set -e
set -u
set -o pipefail

#if [ -n "${PARAMETER_STORE:-}" ]; then
#  export PLANEACION_REPORTES_MID_PGUSER="$(aws ssm get-parameter --name /${PARAMETER_STORE}planeacion_reportes_mid/db/username --output text --query Parameter.Value)"
#  export PLANEACION_REPORTES_MID_PGPASS="$(aws ssm get-parameter --with-decryption --name /${PARAMETER_STORE}/planeacion_reportes_mid/db/password --output text --query Parameter.Value)"

exec ./main "$@"