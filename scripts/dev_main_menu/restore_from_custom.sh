#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#global_variables
year=`date +%Y`; month=`date +%m`; day=`date +%d`; hour=`date +%H`; minute=`date +%M`; second=`date +%S`;
user="sst_user"
password="demo"
last_backup_path="/media/BACKUPDBDEV/last_backup"
log="/home/bin/main_menu/log/log_${year}${month}${day}"

restore() {
    database="$1"
    file_name_new="$2"
    echo ""
    if [ ! -f "$file_name_new" ] || [[ -z "$file_name_new" ]]
    then
        echo "Backup file not found";
        exit;
    fi
    if [[ "$file_name_new" != *.gz ]] && [[ "$file_name_new" != *.enc ]]
    then
        echo "Backup file is not in .gz or .enc format";
        exit;
    fi
    mkdir -p "${last_backup_path}/${year}${month}${day}"
    file_name_old="${last_backup_path}/${year}${month}${day}/${database}_${year}${month}${day}_${hour}${minute}${second}.sql.gz"
    echo "Restore use backup file: $file_name_new"
    echo "Backup $database started  at $(date +'%d %B %Y %H:%M:%S')"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysqldump -u"$user" -p"$password" -h"$host" -P"$port" -CfQq --max-allowed-packet=1G --hex-blob --order-by-primary --opt --single-transaction --routines=true --triggers=true --no-data=false "$database" | gzip -c > "$file_name_old" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    echo "Backup $database finished at $(date +'%d %B %Y %H:%M:%S')"
    echo "Backup for $database saved in ${last_backup_path}/${year}${month}${day}"
    echo "Restore $database started  at $(date +'%d %B %Y %H:%M:%S')"
    mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "DROP DATABASE IF EXISTS $database; CREATE DATABASE IF NOT EXISTS $database;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    if [[ "$file_name_new" == *.gz ]]
    then
        mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && gunzip -c < "$file_name_new" | mysql -u"$user" -p"$password" -h"$host" -P"$port" -f "$database" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    fi
    if [[ "$file_name_new" == *.enc ]]
    then
        mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && openssl enc -d -aes-256-cbc -md sha1 -k "${file_name_new##*/}" -in "$file_name_new" | gunzip -c | mysql -u"$user" -p"$password" -h"$host" -P"$port" -f "$database" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
    fi
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
echo "RESTORE DATABASE FROM CUSTOM"
echo ""
read -p "Enter main database backup file path: " main_database_backup_file_path
if [[ -z "$main_database_backup_file_path" ]]
then
    echo "Please provide a valid main database backup file path";
    exit;
fi
read -p "Enter dmart database backup file path (Optional): " dmart_database_backup_file_path
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
create_grant "$admin_user" "$admin_password"
create_grant "$fin_user" "$fin_password"
create_grant "$user_user" "$user_password"
if [[ -n "$dmart_database_backup_file_path" ]]
then
    restore "$dmart_database" "$dmart_database_backup_file_path" 2>&1 | tee -a "$log"
fi
restore "$main_database" "$main_database_backup_file_path" 2>&1 | tee -a "$log"
mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "DROP DATABASE IF EXISTS ${temp_database}; CREATE DATABASE IF NOT EXISTS ${temp_database};" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
