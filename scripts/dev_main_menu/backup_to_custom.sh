#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#global_variables
year=`date +%Y`; month=`date +%m`; day=`date +%d`; hour=`date +%H`; minute=`date +%M`; second=`date +%S`;
user="sst_user"
password="demo"
host="192.168.102.125"
port="3306"
log="/home/bin/main_menu/log/log_${year}${month}${day}"

backup() {
    mkdir -p "$path"
    for database in ${databases[@]}
    do
        file_name="${path}/${database}_${year}${month}${day}_${hour}${minute}${second}.sql.gz"
        echo ""
        echo "Backup $database started  at $(date +'%d %B %Y %H:%M:%S')"
        mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysqldump -u"$user" -p"$password" -h"$host" -P"$port" -CfQq --max-allowed-packet=1G --hex-blob --order-by-primary --opt --single-transaction --routines=true --triggers=true --no-data=false "$database" | gzip -c > "$file_name" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
        echo "Backup $database finished at $(date +'%d %B %Y %H:%M:%S')"
    done
    echo ""
    echo "Backup for $account_code saved in $path"
}

clear
echo "BACKUP DATABASE TO CUSTOM"
echo ""
read -p "Enter destination path: " path
if [[ -z "$path" ]]
then
    echo "Please provide a valid destination path";
    exit;
fi
read -p "Enter account code: " account_code
account_code=$(echo "$account_code" | tr '[:upper:]' '[:lower:]')
main_databases=($(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D dbsaas_host -sNe "SELECT schema_name FROM information_schema.schemata WHERE schema_name REGEXP '_$account_code(_|$)' AND schema_name NOT LIKE '%_dmart' AND schema_name NOT LIKE '%_temp' ORDER BY schema_name ASC;"))
if [[ -z "${main_databases[*]}" ]]
then
    echo "Account code not found";
    exit;
fi
echo ""
echo "Account code of $account_code:"
number="1"
for main_database in ${main_databases[@]}
do
    echo "$number. ${main_database#*_*_}"
    number="$(($number+1))"
done
sleep 1
read -p "Select the account code to backup: " number_chosen
if [[ -z "$number_chosen" ]]
then
    echo "Please select a valid account code";
    exit;
elif [[ "$number_chosen" -gt 0 && "$number_chosen" -lt "$number" ]]
then
    index="$(($number_chosen-1))"
    main_database_chosen="$(echo -n "${main_databases[index]}")"
    if [[ -n "$main_database_chosen" ]]
    then
        databases=("$main_database_chosen" "${main_database_chosen}_dmart")
        backup 2>&1 | tee -a "$log"
    fi
else
    echo "Please select a valid account code";
    exit;
fi
