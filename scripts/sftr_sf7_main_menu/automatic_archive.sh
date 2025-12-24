#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

#global_variables
year=`date +%Y`; month=`date +%m`; day=`date +%d`; hour=`date +%H`; minute=`date +%M`; second=`date +%S`;
user="sst_user"
password="demo"
host="172.17.71.139"
port="33304"
day_threshold="45"
path="/media/secondary/archive_backup/${year}${month}${day}"
log_dir="${SCRIPT_DIR}/log"
mkdir -p "$log_dir"
log="${log_dir}/log_${year}${month}${day}"

archive() {
    mkdir -p "$path"
    for database in ${databases[@]}
    do
        echo ""
        echo "Backup $database started  at $(date)"
#        mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysqldump -u"$user" -p"$password" -h"$host" -P"$port" -CfQq --max-allowed-packet=1G --hex-blob --order-by-primary --opt --single-transaction --routines=true --triggers=true --no-data=false "$database" | gzip -c > "$file_name" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
        echo "Backup $database finished at $(date)"
        if [ -f "$file_name" ]
        then
            echo "Drop $database started  at $(date)"
#            mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "DROP DATABASE IF EXISTS $database;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
#            if [[ "$database" == "$main_database" ]]
#            then
#                mysql -u"$user" -p"$password" -h"$host" -P"$port" -D dbsaas_host -sNe "UPDATE database_history SET archived = now() WHERE name = '$main_database';"
#            fi
            echo "Drop $database finished at $(date)"
        else
            echo "Drop $database failed, backup file didn't exist"
        fi
    done
    echo ""
    echo "Backup for $main_database saved in $path"
}

main_databases=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D dbsaas_host -sNe "SELECT name FROM database_history WHERE DATE(last_check) = CURDATE() AND last_login_in_days_ago > $day_threshold ORDER BY last_login ASC;")
for main_database in ${main_databases[@]}
do
    main_database=$(echo "$main_database" | tr '[:upper:]' '[:lower:]')
    databases=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D dbsaas_host -sNe "SELECT schema_name FROM information_schema.schemata WHERE SCHEMA_NAME = '$main_database' OR REPLACE(REPLACE(REPLACE(SCHEMA_NAME,'_archive',''),'_dmart',''),'_temp','') = '$main_database' ORDER BY schema_name ASC;")
    if [[ -z "${databases[*]}" ]]
    then
        echo "Database not found";
        exit;
    fi
    archive | tee -a "$log"
done
