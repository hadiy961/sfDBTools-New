#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#global_variables
year=`date +%Y`; month=`date +%m`; day=`date +%d`; hour=`date +%H`; minute=`date +%M`; second=`date +%S`;
user="sst_user"
password="demo"
log="/home/bin/main_menu/log/log_${year}${month}${day}"

backup() {
    database=$(echo "${account_detail[0]}" | tr '[:upper:]' '[:lower:]')
    vip="${account_detail[1]}"
    host="${account_detail[2]}"
    port="${account_detail[3]}"
    app="${account_detail[4]}"
    case "$app" in
    "GD"|"GD-PRO"|"SF-6")
        if [[ "$vip" == "v168" ]]
        then
            app_check=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D "$database" -sNe "SELECT pversion FROM tsfclicense;")
            if [[ "$app_check" == "6.0" ]]
            then
                path="/media/dbtosecondary-nbc/sf6_gdpro/production/${year}${month}${day}"
            elif [[ "$app_check" == "7.0" || "$app_check" == "8.0" ]]
            then
                path="/media/dbtosecondary-nbc/sf7/production/${year}${month}${day}"
            fi
        else
            path="/media/dbtosecondary-nbc/sf6_gdpro/production/${year}${month}${day}"
        fi
        ;;
    "SF-7"|"SF7-NBC2")
        path="/media/dbtosecondary-nbc/sf7/production/${year}${month}${day}"
        ;;
    esac
    if [[ -z "$path" ]]
    then
        echo "Please provide a valid destination path";
        exit;
    fi
    mkdir -p "$path"
    echo ""
    echo "Backup $account_code from $vip ($host:$port)"
    for table in ${tables[@]}
    do
        table_check=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D "$database" -sNe "SHOW TABLES LIKE '$table';")
        if [[ "$table_check" == "$table" ]]
        then
            file_name="${path}/${database}_${table}_${year}${month}${day}_${hour}${minute}${second}.sql.gz"
            echo ""
            echo "Backup $table started  at $(date +'%d %B %Y %H:%M:%S')"
            mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysqldump -u"$user" -p"$password" -h"$host" -P"$port" -CfQq --max-allowed-packet=1G --hex-blob --order-by-primary --opt --single-transaction --routines=false --triggers=true --no-data=false "$database" "$table" | gzip -c > "$file_name" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
            echo "Backup $table finished at $(date +'%d %B %Y %H:%M:%S')"
        else
	    echo ""
            echo "Table $table is not found in $database"
        fi
    done
    echo ""
    echo "Backup for $account_code saved in $path"
}

clear
echo "BACKUP TABLE TO SECONDARY"
echo ""
read -p "Enter account code: " account_code
account_code=$(echo "$account_code" | tr '[:upper:]' '[:lower:]')
account_detail=($(mysql -u"$user" -p"$password" -h172.17.70.170 -P3306 -D dbsf_tools -sNe "SELECT A.dbname, B.vip, B.ip, B.port_1, B.sf FROM tmshdata A INNER JOIN v_dba_master_cluster B ON (A.host = B.virtual_ip AND A.port = B.port_2 AND B.is_master = 'Master') WHERE A.codename = '$account_code';"))
if [[ -z "${account_detail[*]}" ]]
then
    echo "Account detail not found";
    exit;
fi
read -p "Enter table (use comma if more than one): " tables
tables=$(echo "$tables" | tr '[:upper:]' '[:lower:]')
tables="${tables// /}"
tables="${tables//,/ }"
if [[ -z "$tables" ]]
then
    echo "Please enter a valid table";
    exit;
fi
backup 2>&1 | tee -a "$log"
