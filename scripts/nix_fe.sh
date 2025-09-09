
POSTGRES_USER=fractalstore
POSTGRES_PASSWORD=fractalstore
POSTGRES_DB=fractalstore
PGPORT=$POSTGRES_PORT

FRACTAL_ENGINE_HOST=localhost
FRACTAL_ENGINE_DB="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:$POSTGRES_PORT/${POSTGRES_DB}"

FRACTAL_ENGINE_PORT=8720
BASE_DIR=~/.fractal-stack-1
DOGENET_WEB_PORT=8740
DOGE_RPC_PORT=23256
DOGECOIN_RPC_USER=dogecoinrpc
DOGECOIN_RPC_PASSWORD=changeme1
FRACTAL_ENGINE_DB=

go run cmd/fractal-engine/fractal_engine.go \
    --rpc-server-host 0.0.0.0 \
    --rpc-server-port $FRACTAL_ENGINE_PORT \
    --doge-net-network unix \
    --doge-net-address $BASE_DIR/dogenet.sock \
    --doge-net-web-address 0.0.0.0:$DOGENET_WEB_PORT \
    --doge-scheme http \
    --doge-host localhost \
    --doge-port $DOGE_RPC_PORT \
    --doge-user $DOGECOIN_RPC_USER \
    --doge-password $DOGECOIN_RPC_PASSWORD \
    --database-url $FRACTAL_ENGINE_DB?sslmode=disable \
    --embed-dogenet true
