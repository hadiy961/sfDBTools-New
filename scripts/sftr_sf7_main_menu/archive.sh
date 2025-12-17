#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#global_variables
year=`date +%Y`; month=`date +%m`; day=`date +%d`; hour=`date +%H`; minute=`date +%M`; second=`date +%S`;
user="sst_user"
password="demo"
host="172.17.71.139"
port="33304"
path="/media/secondary/archive_backup/${year}${month}${day}"
log="/home/bin/main_menu/log/log_${year}${month}${day}"

archive() {
    databases=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D dbsaas_host -sNe "SELECT schema_name FROM information_schema.schemata WHERE SCHEMA_NAME = '$main_database_archived' OR REPLACE(REPLACE(REPLACE(SCHEMA_NAME,'_archive',''),'_dmart',''),'_temp','') = '$main_database_archived' ORDER BY schema_name ASC;")
    archived_date=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D dbsaas_host -sNe "SELECT archived FROM database_history WHERE name = '$main_database_archived';")
    mkdir -p "$path"
    for database in ${databases[@]}
    do
        if [[ -n "$archived_date" ]] && [[ "$archived_date" != "NULL" ]]
        then
            find /media/secondary/archive_backup/${archived_date:0:4}${archived_date:5:2}${archived_date:8:2}/ -type f -name "${database}_${archived_date:0:4}${archived_date:5:2}${archived_date:8:2}_[0-9]*.sql.gz" -exec rm {} \;
        fi
        file_name="${path}/${database}_${year}${month}${day}_${hour}${minute}${second}.sql.gz"
        echo ""
        echo "Backup $database started  at $(date +'%d %B %Y %H:%M:%S')"
        mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysqldump -u"$user" -p"$password" -h"$host" -P"$port" -CfQq --max-allowed-packet=1G --hex-blob --order-by-primary --opt --single-transaction --routines=true --triggers=true --no-data=false "$database" | gzip -c > "$file_name" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
        echo "Backup $database finished at $(date +'%d %B %Y %H:%M:%S')"
        if [ -f "$file_name" ]
        then
            echo "Drop $database started  at $(date +'%d %B %Y %H:%M:%S')"
            mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=0;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "DROP DATABASE IF EXISTS $database;" && mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SET GLOBAL max_statement_time=60;"
            if [[ "$database" == "$main_database_archived" ]]
            then
                mysql -u"$user" -p"$password" -h"$host" -P"$port" -D dbsaas_host -sNe "UPDATE database_history SET archived = now() WHERE name = '$main_database_archived';"
            fi
            echo "Drop $database finished at $(date +'%d %B %Y %H:%M:%S')"
        else
            echo "Drop $database failed, backup file didn't exist"
        fi
    done
    echo ""
    echo "Backup for $main_database_archived saved in $path"
    echo ""
}

clear
echo "ARCHIVE DATABASE (Backup and Drop)"
echo ""
read -p "Enter last login (in days ago) threshold for database to archived: " day_threshold
if ! [[ "$day_threshold" =~ ^[0-9]+$ ]]
then
    echo "Please enter a valid threshold";
    exit;
fi
database_history=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SELECT dh.name, dh.last_login_in_days_ago FROM dbsaas_host.database_history dh JOIN information_schema.schemata s ON  dh.name = s.schema_name WHERE DATE(dh.last_check) = CURDATE() AND dh.last_login_in_days_ago > $day_threshold ORDER BY dh.last_login ASC;")
if [[ -z $database_history ]]
then
    echo "No account code found";
    exit;
else
    main_databases=()
    echo ""
    echo "Account code that last logged in more than $day_threshold days ago: "
    number="1"
    while read -r main_database last_login_in_days_ago
    do
        main_databases+=("$main_database")
        echo -e "$number. [$last_login_in_days_ago days ago]\t${main_database#*_*_}"
        number="$(($number+1))"
    done <<< "$database_history"
    sleep 1
    echo ""
    echo "1. Archive all"
    echo "2. Archive with exception"
    echo ""
    echo "0. Exit"
    read -p "Select an option: " archive_option
    case "$archive_option" in
        0)
            exit;
            ;;
        1)
            for main_database_archived in ${main_databases[@]}
            do
                archive 2>&1 | tee -a "$log"
            done
            ;;
        2)
            echo ""
            read -p "Select account code that excluded (use comma if more than one): " excludeds
            excludeds="${excludeds// /}"
            excludeds="${excludeds//,/ }"
            if ! [[ "$excludeds" =~ ^[0-9\ ]+$ && "$excludeds" =~ [0-9] ]]
            then
                echo "Please select a valid account code";
                exit;
            else
                for excluded in ${excludeds[@]}
                do
                    if [[ "$excluded" -lt $number ]]
                    then
                        unset 'main_databases[$((excluded - 1))]'
                    else
                        echo "Please select a valid account code";
                        exit;
                    fi
                done
                for main_database_archived in ${main_databases[@]}
                do
                    archive 2>&1 | tee -a "$log"
                done
            fi
            ;;
        *)
            echo "Please select a valid option";
            exit;
            ;;
    esac
fi
