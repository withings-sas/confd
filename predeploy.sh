set -x
set -e
set -o pipefail

ENVIRONMENT=$1
COMMIT=$2

case "$ENVIRONMENT" in
        prod)
                NICE=0
                ;;
        *)
                NICE=10
                ;;
esac


BUILD_SERVER=fr-hq-build-02.corp.withings.com
BUILD_USER=scaleweb
BUILD_PATH=/home/$BUILD_USER/confd-$ENVIRONMENT

ssh $BUILD_USER@$BUILD_SERVER "mkdir -p $BUILD_PATH"

rsync -az --delete --exclude="bin/" ./ $BUILD_USER@$BUILD_SERVER:$BUILD_PATH/

ssh $BUILD_USER@$BUILD_SERVER "cd $BUILD_PATH; bash build.sh"

rsync -avz --dry-run $BUILD_USER@$BUILD_SERVER:$BUILD_PATH/bin/ bin/

