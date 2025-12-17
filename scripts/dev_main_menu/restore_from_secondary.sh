#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#global_variables
year=`date +%Y`; month=`date +%m`; day=`date +%d`; hour=`date +%H`; minute=`date +%M`; second=`date +%S`;
user="sst_user"
password="demo"
path="/media/BACKUPDBDEV/last_backup"
log="/home/bin/main_menu/log/log_${year}${month}${day}"

restore() {
    database="$1"
    if [[ "$database" == *_dmart ]]; then
        file_name_find="$(find "/media/dbtolocal-nbc/sf7/secondary/${year}${month}${day}" -maxdepth 1 -type f -name "dbsf_*${secondary_account_code}_dmart_${year}${month}${day}_[0-9]*.sql.gz" 2>/dev/null)"
    else
        file_name_find="$(find "/media/dbtolocal-nbc/sf7/secondary/${year}${month}${day}" -maxdepth 1 -type f -name "dbsf_*${secondary_account_code}_${year}${month}${day}_[0-9]*.sql.gz" 2>/dev/null)"
    fi
    echo ""
    if [[ $(echo "$file_name_find" | wc -l)>1 ]]
    then
        echo "Found multiple backup files"
        readarray file_name_arrays <<< "$file_name_find"
        number="1"
        for file_name_array in ${file_name_arrays[@]}
        do
                timestamp=$(echo $file_name_array | sed 's/\.sql\.enc$//' | rev | cut -d'_' -f1-2 | rev)
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

create_grant() {
    create_user="$1"
    create_password="$2"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "CREATE USER IF NOT EXISTS '${create_user}'@'%' IDENTIFIED BY '${create_password}';" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "GRANT ALL PRIVILEGES ON ${main_database}.* to '${create_user}'@'%';" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "GRANT ALL PRIVILEGES ON ${dmart_database}.* to '${create_user}'@'%';" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "GRANT ALL PRIVILEGES ON ${temp_database}.* to '${create_user}'@'%';" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "FLUSH PRIVILEGES;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
}

clear
echo "RESTORE DATABASE FROM SECONDARY"
echo ""
echo "1. From Secondary Training"
echo "2. From Secondary Dev"
echo "3. From Secondary (Custom)"
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
echo "Paste account detail here (from SF Cola) and then press Enter and Ctrl+D"
echo ""
account_detail="/home/bin/main_menu/account_detail.txt"
> "$account_detail"
$(cat > "$account_detail")
account_code=$(cat "$account_detail" | grep -w "Instance:" | awk '{print $2}' | tr '[:upper:]' '[:lower:]')
host=$(cat "$account_detail" | grep -w "Host:" | awk '{print $2}')
port=$(cat "$account_detail" | grep -w "Port:" | awk '{print $2}')
main_database=$(cat "$account_detail" | grep -w "Database:" | awk '{print $2}' | tr '[:upper:]' '[:lower:]')
dmart_database=$(cat "$account_detail" | grep -w "Database DMART:" | awk '{print $3}' | tr '[:upper:]' '[:lower:]')
admin_user=$(cat "$account_detail" | grep -w "User Admin:" | awk '{print $3}')
admin_password=$(cat "$account_detail" | grep -w "Pass Admin:" | awk '{print $3}')
fin_user=$(cat "$account_detail" | grep -w "User Fin:" | awk '{print $3}')
fin_password=$(cat "$account_detail" | grep -w "Pass Fin:" | awk '{print $3}')
user_user=$(cat "$account_detail" | grep -w "User User:" | awk '{print $3}')
user_password=$(cat "$account_detail" | grep -w "Pass User:" | awk '{print $3}')
database_type=$(cat "$account_detail" | grep -w "Database Type:" | awk '{print $3}' | tr '[:upper:]' '[:lower:]')
temp_database="${main_database}_temp"
if [[ ! -s "$account_detail" ]]
then
    echo "Please provide account detail";
    exit;
fi
if [[ -z "$account_code" || -z "$host" || -z "$port" || -z "$main_database" || -z "$dmart_database" || -z "$admin_user" || -z "$admin_password" || -z "$fin_user" || -z "$fin_password" || -z "$user_user" || -z "$user_password" || -z "$database_type" ]]
then
    echo "Please paste all account detail";
    exit;
fi
if [[ "$database_type" != "mariadb" ]]
then
    echo "Account is not using mariadb";
    exit;
fi
if [[ "$host" != "192.168.102.125" ]]
then
    echo "Please use the correct host";
    exit;
fi
if [[ "$port" != "3306" ]]
then
    echo "Please use the correct port";
    exit;
fi
secondary_account_code="${account_code}_${secondary}"
create_grant "$admin_user" "$admin_password"
create_grant "$fin_user" "$fin_password"
create_grant "$user_user" "$user_password"
restore "$dmart_database" 2>&1 | tee -a "$log"
restore "$main_database" 2>&1 | tee -a "$log"
mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "DROP DATABASE IF EXISTS ${temp_database}; CREATE DATABASE IF NOT EXISTS ${temp_database};" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
