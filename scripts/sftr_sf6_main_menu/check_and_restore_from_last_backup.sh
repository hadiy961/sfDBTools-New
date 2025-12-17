#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#global_variables
year=`date +%Y`; month=`date +%m`; day=`date +%d`; hour=`date +%H`; minute=`date +%M`; second=`date +%S`;
user="sst_user"
password="demo"
host="172.17.70.40"
port="13310"
path="/media/dbtosecondary-nbc/last_backup"
log="/home/bin/main_menu/log/log_${year}${month}${day}"

restore() {
    echo ""
    database="$1"
    file_name_new="$2"
    mkdir -p "${path}/${year}${month}${day}"
    file_name_old="${path}/${year}${month}${day}/${database}_${year}${month}${day}_${hour}${minute}${second}.sql.gz"
    echo "Restore use backup file: $file_name_new"
    echo "Backup $database started  at $(date +'%d %B %Y %H:%M:%S')"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysqldump -u"$user" -p"$password" -h"$host" -P"$port" -CfQq --max-allowed-packet=1G --hex-blob --order-by-primary --opt --single-transaction --routines=true --triggers=true --no-data=false "$database" | gzip -c > "$file_name_old" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    echo "Backup $database finished at $(date +'%d %B %Y %H:%M:%S')"
    echo "Backup for $database saved in ${path}/${year}${month}${day}"
    echo "Restore $database started  at $(date +'%d %B %Y %H:%M:%S')"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "DROP DATABASE IF EXISTS $database; CREATE DATABASE IF NOT EXISTS $database;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && gunzip -c < "$file_name_new" | mysql -u"$user" -p"$password" -h"$host" -P"$port" -f "$database" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    echo "Restore $database finished at $(date +'%d %B %Y %H:%M:%S')"
}

check() {
    echo "Please wait while searching for available backup files"
    echo ""
    file_name_find=$(find ${path}/*/ -type f -name "${main_database_chosen}_[0-9]*.sql.gz" 2>/dev/null | sort -n)
    if [[ $(echo "$file_name_find" | wc -l)>0 ]] && [[ -n "$file_name_find" ]]
    then
        echo "Available backup files of "$secondary_account_code":"
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
        echo ""
        echo "0. Exit"
        sleep 1
        read -p "Select the date of backup files to restore: " date_chosen
        echo ""
        if [[ -z "$date_chosen" ]]
        then
            echo "Please select a valid date";
            exit;
        elif [[ "$date_chosen" == "0" ]]
        then
            exit
        elif [[ "$date_chosen" -gt 0 && "$date_chosen" -lt "$number" ]]
        then
            index="$(($date_chosen-1))"
            file_name_chosen="$(echo -n "${file_name_arrays[index]}")"
            if [[ -f "$file_name_chosen" ]] || [[ -z "$file_name_chosen" ]]
            then
                dmart_database="${main_database_chosen}_dmart"
                dmart_file_name_chosen="$(echo "$file_name_chosen" | sed "s/\($main_database_chosen\)\(_[0-9]*\)/\1_dmart\2/")"
                echo "Backup files of $secondary_account_code on $(date +'%d %B %Y %H:%M:%S' -r "$file_name_chosen") have been found"
                read -p "Do you want to restore it? [Y/N]: " restore_confirmation
                case "$restore_confirmation" in
                    [Yy]|[Yy][Ee][Ss])
                        restore "$main_database_chosen" "$file_name_chosen" 2>&1 | tee -a "$log"
                        if [[ -f "$dmart_file_name_chosen" ]] || [[ -z "$dmart_file_name_chosen" ]]
                        then
                            restore "$dmart_database" "$dmart_file_name_chosen" 2>&1 | tee -a "$log"
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
                echo "No backup file available. Please check manually";
                exit;
            fi
        else
            echo "Please select a valid date";
            exit;
        fi
    else
        echo "No backup file available"
    fi
}

clear
echo "CHECK AND RESTORE DATABASE FROM LAST BACKUP"
echo ""
read -p "Enter account code: " account_code
account_code=$(echo "$account_code" | tr '[:upper:]' '[:lower:]')
main_databases=($(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D dbsaas_host -sNe "SELECT schema_name FROM information_schema.schemata WHERE schema_name REGEXP '_$account_code(_|$)' AND schema_name NOT LIKE '%_dmart' AND schema_name NOT LIKE '%_temp' ORDER BY schema_name ASC;"))
if [[ -z "${main_databases[*]}" ]]
then
    echo "Account code not found or never backed up";
    exit;
fi
echo ""
echo "Secondary account of $account_code:"
number="1"
for main_database in ${main_databases[@]}
do
    secondary_account_code="${main_database#*_}"
    secondary_account_code="${secondary_account_code#*_}"
    echo "$number. $secondary_account_code"
    number="$(($number+1))"
done
sleep 1
read -p "Select the secondary account to restore: " number_chosen
echo ""
if [[ -z "$number_chosen" ]]
then
    echo "Please select a valid secondary account";
    exit;
elif [[ "$number_chosen" -gt 0 && "$number_chosen" -lt "$number" ]]
then
    index="$(($number_chosen-1))"
    main_database_chosen="$(echo -n "${main_databases[index]}")"
    if [[ -n "$main_database_chosen" ]]
    then
    check
    fi
else
    echo "Please select a valid secondary account";
    exit;
fi
