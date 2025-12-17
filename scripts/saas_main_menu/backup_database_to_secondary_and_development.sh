#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#global_variables
year=`date +%Y`; month=`date +%m`; day=`date +%d`; hour=`date +%H`; minute=`date +%M`; second=`date +%S`;
user="sst_user"
password="demo"
log="/home/bin/main_menu/log/log_${year}${month}${day}"

backup() {
    main_database=$(echo "${account_detail[0]}" | tr '[:upper:]' '[:lower:]')
    vip="${account_detail[1]}"
    host="${account_detail[2]}"
    port="${account_detail[3]}"
    app="${account_detail[4]}"
    dmart_database="${main_database}_dmart"
    databases=("$main_database" "$dmart_database")
    case "$app" in
    "GD"|"GD-PRO"|"SF-6")
        if [[ "$vip" == "v168" ]]
        then
            app_check=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D "$main_database" -sNe "SELECT pversion FROM tsfclicense;")
            if [[ "$app_check" == "6.0" ]]
            then
                path_secondary="/media/dbtosecondary-nbc/sf6_gdpro/production/${year}${month}${day}"
                path_development="/media/dbtolocal-nbc/sf6_gdpro/production/${year}${month}${day}"
            elif [[ "$app_check" == "7.0" || "$app_check" == "8.0" ]]
            then
                path_secondary="/media/dbtosecondary-nbc/sf7/production/${year}${month}${day}"
                path_development="/media/dbtolocal-nbc/sf7/production/${year}${month}${day}"
            fi
        else
            path_secondary="/media/dbtosecondary-nbc/sf6_gdpro/production/${year}${month}${day}"
            path_development="/media/dbtolocal-nbc/sf6_gdpro/production/${year}${month}${day}"
        fi
        ;;
    "SF-7"|"SF7-NBC2")
        path_secondary="/media/dbtosecondary-nbc/sf7/production/${year}${month}${day}"
        path_development="/media/dbtolocal-nbc/sf7/production/${year}${month}${day}"
        ;;
    esac
    if [[ -z "$path_secondary" ]] || [[ -z "$path_development" ]]
    then
        echo "Please provide a valid destination path";
        exit;
    fi
    mkdir -p "$path_secondary"
    mkdir -p "$path_development"
    echo ""
    echo "Backup $account_code from $vip ($host:$port)"
    grant="${path_secondary}/grant_${account_code}.sql"
    > $grant
    grant_users=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SELECT user FROM mysql.user WHERE user LIKE CONCAT('%$account_code','_admin') OR user LIKE CONCAT('%$account_code','_user') OR user LIKE CONCAT('%$account_code','_fin')")
    for grant_user in ${grant_users[@]}
    do
        grant_password=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SELECT password FROM mysql.user WHERE user='$grant_user';")
        echo "GRANT ALL PRIVILEGES ON database.* TO '$grant_user'@'%' IDENTIFIED BY PASSWORD '$grant_password';" >> "$grant"
    done
    for database in ${databases[@]}
    do
        file_name="${path_secondary}/${database}_${year}${month}${day}_${hour}${minute}${second}.sql.gz"
        echo ""
        echo "Backup $database started  at $(date +'%d %B %Y %H:%M:%S')"
        mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysqldump -u"$user" -p"$password" -h"$host" -P"$port" -CfQq --max-allowed-packet=1G --hex-blob --order-by-primary --opt --single-transaction --routines=true --triggers=true --no-data=false "$database" | gzip -c > "$file_name" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
        echo "Backup $database finished at $(date +'%d %B %Y %H:%M:%S')"
        cp "$file_name" "$path_development"
    done
    echo ""
    echo "Backup for $account_code saved in $path_secondary and $path_development"
}

clear
echo "BACKUP DATABASE TO SECONDARY AND DEVELOPMENT"
echo ""
read -p "Enter account code: " account_code
account_code=$(echo "$account_code" | tr '[:upper:]' '[:lower:]')
account_detail=($(mysql -u"$user" -p"$password" -h172.17.70.170 -P3306 -D dbsf_tools -sNe "SELECT A.dbname, B.vip, B.ip, B.port_1, B.sf FROM tmshdata A INNER JOIN v_dba_master_cluster B ON (A.host = B.virtual_ip AND A.port = B.port_2 AND B.is_master = 'Master') WHERE A.codename = '$account_code';"))
if [[ -z "${account_detail[*]}" ]]
then
    echo "Account detail not found";
    exit;
fi
backup 2>&1 | tee -a "$log"
