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
log_dir="${SCRIPT_DIR}/log"
mkdir -p "$log_dir"
log="${log_dir}/log_${year}${month}${day}"

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
    echo "Backup for $secondary_account_code saved in $path"
}

clear
echo "BACKUP DATABASE TO CUSTOM"
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
read -p "Enter destination path: " path
if [[ -z "$path" ]]
then
    echo "Please provide a valid destination path";
    exit;
fi
read -p "Enter account code: " account_code
account_code=$(echo "$account_code" | tr '[:upper:]' '[:lower:]')
secondary_account_code="${account_code}_${secondary}"
databases=$(mysql -u"$user" -p"$password" -D dbsaas_host -sNe "SELECT schema_name FROM information_schema.schemata WHERE REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(SCHEMA_NAME,'dbsf_nbc_',''),'dbsf_biznet_',''),'dbsf_saasmain_',''),'dbsf_saasdev_',''),'_dmart','') = '$secondary_account_code' ORDER BY schema_name ASC;")
if [[ -z "${databases[*]}" ]]
then
    echo "Database not found";
    exit;
fi
backup 2>&1 | tee -a "$log"
