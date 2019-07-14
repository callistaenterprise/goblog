#!/bin/sh

echo "HELLO!!"
./cockroach sql --insecure --host="goblog_cockroachdb1" --execute="DROP DATABASE IF EXISTS $1 CASCADE; CREATE DATABASE IF NOT EXISTS $1;"


#./cockroach sql --insecure --host="goblog_cockroachdb1" --execute="DROP USER IF EXISTS goblog; \
#                                        DROP DATABASE IF EXISTS $1 CASCADE; \
#                                        CREATE DATABASE IF NOT EXISTS $1; \
#                                        CREATE USER goblog WITH PASSWORD 'password'; \
#                                        GRANT ALL ON DATABASE $1 TO goblog;"