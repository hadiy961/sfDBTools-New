#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

#global_variables
year=`date +%Y`; month=`date +%m`; day=`date +%d`; hour=`date +%H`; minute=`date +%M`; second=`date +%S`;
user="sst_user"
password="demo"
host="172.17.71.139"
port="33304"
path="/media/BACKUPDBDEV/archive_backup/${year}${month}${day}"
log_dir="${SCRIPT_DIR}/log"
mkdir -p "$log_dir"
log="${log_dir}/log_${year}${month}${day}"

clear
echo "SET"
echo ""
read -p "Enter account code: " account_code
account_code=$(echo "$account_code" | tr '[:upper:]' '[:lower:]')
archived_databases=($(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D dbsaas_host -sNe "SELECT name FROM dbsaas_host.database_history WHERE name REGEXP '_aaa(_|$)' AND archived IS NOT NULL ORDER BY name ASC;"))
if [[ -z "${archived_databases[*]}" ]]
then
    echo "Account code not found or never archived";
    exit;
fi
echo ""
echo "Archived databases of $account_code:"
number="1"
for archived_database in ${archived_databases[@]}
do
    echo "$number. $archived_database"
    number="$(($number+1))"
done
sleep 1
read -p "Select the archived database to restore: " number_chosen
echo ""
if [[ -z "$number_chosen" ]]
then
    echo "Please select a valid database";
    exit;
elif [[ "$number_chosen" -gt 0 && "$number_chosen" -lt "$number" ]]
then
    index="$(($number_chosen-1))"
    archived_database_chosen="$(echo -n "${archived_databases[index]}")"
    if [[ -n "$archived_database_chosen" ]]
    then
    databases=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D dbsaas_host -sNe "SELECT schema_name FROM information_schema.schemata WHERE SCHEMA_NAME = '$archived_database_chosen' OR REPLACE(REPLACE(REPLACE(SCHEMA_NAME,'_archive',''),'_dmart',''),'_temp','') = '$archived_database_chosen' ORDER BY schema_name ASC;")
        if [[ -z "${databases[*]}" ]]
        then
            echo "Database not found";
            exit;
        fi
    fi
else
    echo "Please select a valid main database";
    exit;
fi
#archived_database=$(echo "$archived_database" | tr '[:upper:]' '[:lower:]')
for database in ${databases[@]}
do
    echo $database
done
