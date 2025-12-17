#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#global_variables
year=`date +%Y`; month=`date +%m`; day=`date +%d`; hour=`date +%H`; minute=`date +%M`; second=`date +%S`;
user="sst_user"
password="demo"
host="172.17.71.139"
port="33304"
archived_path="/media/secondary/archive_backup"
backup_path="/media/dbtosecondary-nbc/last_backup"
log="/home/bin/main_menu/log/log_${year}${month}${day}"

restore() {
    echo ""
    database="$1"
    file_name_new="$2"
    mkdir -p "${backup_path}/${year}${month}${day}"
    file_name_old="${backup_path}/${year}${month}${day}/${database}_${year}${month}${day}_${hour}${minute}${second}.sql.gz"
    echo "Restore use backup file: $file_name_new"
    echo "Backup $database started  at $(date +'%d %B %Y %H:%M:%S')"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysqldump -u"$user" -p"$password" -h"$host" -P"$port" -CfQq --max-allowed-packet=1G --hex-blob --order-by-primary --opt --single-transaction --routines=true --triggers=true --no-data=false "$database" | gzip -c > "$file_name_old" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    echo "Backup $database finished at $(date +'%d %B %Y %H:%M:%S')"
    echo "Backup for $database saved in ${backup_path}/${year}${month}${day}"
    echo "Restore $database started  at $(date +'%d %B %Y %H:%M:%S')"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "DROP DATABASE IF EXISTS $database; CREATE DATABASE IF NOT EXISTS $database" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && gunzip -c < "$file_name_new" | mysql -u"$user" -p"$password" -h"$host" -P"$port" -f "$database" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    echo "Restore $database finished at $(date +'%d %B %Y %H:%M:%S')"
}

check() {
    read main_database archived_date <<< $(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D dbsaas_host -sNe "SELECT name, DATE_FORMAT(archived, '%Y%m%d') FROM database_history WHERE archived IS NOT NULL AND name = '$archived_database_chosen' ORDER BY archived DESC;")
    if [[ -z "$main_database" ]]
    then
        echo "Database never archived";
        exit;
    fi
    if [ ! -d "${archived_path}/${archived_date}" ]
    then
        echo "Archived files have been deleted";
        exit;
    fi
    main_database_find="$(find "${archived_path}/${archived_date}" -type f -name "${main_database}_${archived_date}_[0-9]*.sql.gz" | sort -n | tail -n 1 2>/dev/null)"
    dmart_database="${main_database}_dmart"
    temp_database="${main_database}_temp"
    dmart_database_find="$(find "${archived_path}/${archived_date}" -type f -name "${dmart_database}_${archived_date}_[0-9]*.sql.gz" | sort -n | tail -n 1 2>/dev/null)"
    if [[ -n "$main_database_find" ]]
    then
        echo "Archived files of $main_database on $(date +'%d %B %Y %H:%M:%S' -r "$main_database_find") have been found"
        read -p "Do you want to restore it? [Y/N]: " restore_confirmation
        case "$restore_confirmation" in
            [Yy]|[Yy][Ee][Ss])
                restore "$main_database" "$main_database_find" 2>&1 | tee -a "$log"
                mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "DROP DATABASE IF EXISTS ${temp_database}; CREATE DATABASE IF NOT EXISTS ${temp_database}" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
                        if [[ -f "$dmart_database_find" ]] && [[ -n "$dmart_database_find" ]]
                then
                    restore "$dmart_database" "$dmart_database_find" 2>&1 | tee -a "$log"
                fi
                ;;
            [Nn]|[Nn][Oo])
                exit;
                ;;
            *)
                echo "Invalid input. Please answer with 'Y' or 'N'";
                exit;
                ;;
        esac
    else
        echo "Archived files have been deleted";
        exit;
    fi
}

clear
echo "CHECK AND RESTORE DATABASE FROM ARCHIVE"
echo ""
read -p "Enter account code: " account_code
account_code=$(echo "$account_code" | tr '[:upper:]' '[:lower:]')
archived_databases=($(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D dbsaas_host -sNe "SELECT name FROM dbsaas_host.database_history WHERE name REGEXP '_$account_code(_|$)' AND archived IS NOT NULL ORDER BY name ASC;"))
if [[ -z "${archived_databases[*]}" ]]
then
    echo "Account code not found or never archived";
    exit;
fi
echo ""
echo "Account code of $account_code:"
number="1"
for archived_database in ${archived_databases[@]}
do
    echo "$number. ${archived_database#*_*_}"
    number="$(($number+1))"
done
sleep 1
read -p "Select the account code to restore: " number_chosen
echo ""
if [[ -z "$number_chosen" ]]
then
    echo "Please select a valid account code";
    exit;
elif [[ "$number_chosen" -gt 0 && "$number_chosen" -lt "$number" ]]
then
    index="$(($number_chosen-1))"
    archived_database_chosen="$(echo -n "${archived_databases[index]}")"
    if [[ -n "$archived_database_chosen" ]]
    then
    check
    fi
else
    echo "Please select a valid account code";
    exit;
fi
