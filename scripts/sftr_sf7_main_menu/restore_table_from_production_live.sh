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
path="/media/dbtosecondary-nbc/last_backup"
log_dir="${SCRIPT_DIR}/log"
mkdir -p "$log_dir"
log="${log_dir}/log_${year}${month}${day}"

restore() {
    database_find="$(find /media/dbtosecondary-nbc/sf7/production/${year}${month}${day} -maxdepth 1 -type f -name "dbsf_*_${account_code}_*_${year}${month}${day}_[0-9]*.sql.gz" | sort -n | tail -n 1 2>/dev/null)"
    if [[ ! -f "$database_find" ]]
    then
        echo "Backup file not found";
        exit;
    fi
    database="$(basename "$database_find" .sql.gz | sed -E 's/_[0-9]{8}_[0-9]{6}$//' | sed -E 's/_[^_]+$//')_${secondary}"
    database_check=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SHOW DATABASES LIKE '$database';")
    if [[ "$database_check" == "$database" ]]
    then
        for table in ${tables[@]}
        do
            echo ""
            file_name_find="$(find /media/dbtosecondary-nbc/sf7/production/${year}${month}${day} -maxdepth 1 -type f -name "${database//"_$secondary"/}_${table}_${year}${month}${day}_[0-9]*.sql.gz" 2>/dev/null)"
            if [[ $(echo "$file_name_find" | wc -l)>1 ]]
            then
                echo "Found multiple backup files"
                readarray file_name_arrays <<< "$file_name_find"
                number="1"
                for file_name_array in ${file_name_arrays[@]}
                do
                    timestamp=$(echo $file_name_array | sed 's/\.sql\.gz$//' | rev | cut -d'_' -f1-2 | rev)
                    date_part=$(echo $timestamp | cut -d'_' -f1)
                    time_part=$(echo $timestamp | cut -d'_' -f2)
                    date_string="${date_part}${time_part}"
                    formatted_time=$(date -d "${date_part:0:4}-${date_part:4:2}-${date_part:6:2} ${time_part:0:2}:${time_part:2:2}:${time_part:4:2}" "+%d %B %Y %H:%M:%S")
                    echo "$number. $formatted_time"
                    number="$(($number+1))"
                done
                sleep 1
                read -p "Select a backup file to restore: " file_name_chosen
                if [[ -z "$file_name_chosen" ]]
                then
                    echo "Please select a valid backup file";
                    exit;
                fi
                index="$(($file_name_chosen-1))"
                file_name_new="$(echo -n "${file_name_arrays[index]}")"
            else
                file_name_new="$file_name_find"
            fi
            if [ ! -f "$file_name_new" ] || [[ -z "$file_name_new" ]]
            then
                echo "Backup file not found";
                exit;
            fi
            mkdir -p "${path}/${year}${month}${day}"
            file_name_old="${path}/${year}${month}${day}/${database}_${table}_${year}${month}${day}_${hour}${minute}${second}.sql.gz"
            echo "Restore use backup file: $file_name_new"
            echo "Backup $table started  at $(date +'%d %B %Y %H:%M:%S')"
            mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysqldump -u"$user" -p"$password" -h"$host" -P"$port" -CfQq --max-allowed-packet=1G --hex-blob --order-by-primary --opt --single-transaction --routines=false --triggers=true --no-data=false "$database" "$table" | gzip -c > "$file_name_old" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
            echo "Backup $table finished at $(date +'%d %B %Y %H:%M:%S')"
            echo "Backup for $table saved in ${path}/${year}${month}${day}"
            echo "Restore $table started  at $(date +'%d %B %Y %H:%M:%S')"
            mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && gunzip -c < "$file_name_new" | mysql -u"$user" -p"$password" -h"$host" -P"$port" -f "$database" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
            echo "Restore $table finished at $(date +'%d %B %Y %H:%M:%S')"
        done
    else
        echo ""
        echo "Database $database is not found"
    fi
}

clear
echo "RESTORE TABLE FROM PRODUCTION (LIVE)"
echo ""
echo "1. To Secondary Training"
echo "2. To Secondary Dev"
echo "3. To Secondary (Custom)"
echo ""
echo "0. Exit"
read -p "Select an option: " option_secondary
case "$option_secondary" in
0)
    exit;
    ;;
1)
    secondary="secondary_training"
    ;;
2)
    secondary="secondary_dev"
    ;;
3)
    read -p "Please type the detail: " custom
    secondary="secondary_$custom"
    ;;
*)
    echo "Please select a valid option";
    exit;
    ;;
esac
echo ""
read -p "Enter account code: " account_code
account_code=$(echo "$account_code" | tr '[:upper:]' '[:lower:]')
if [[ -z "$account_code" ]]
then
    echo "Please enter a valid account code";
    exit;
fi
secondary_account_code="${account_code}_${secondary}"
read -p "Enter table (use comma if more than one): " tables
tables=$(echo "$tables" | tr '[:upper:]' '[:lower:]')
tables="${tables// /}"
tables="${tables//,/ }"
if [[ -z "$tables" ]]
then
    echo "Please enter a valid table";
    exit;
fi
restore 2>&1 | tee -a "$log"
